package handler

import (
	"embed"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/twttr/kpuppy-backend/internal/usecase"
)

type AdminWebHandler struct {
	userService    *usecase.UserService
	commentService *usecase.CommentService
	templates      map[string]*template.Template
	basePath       string
}

type DashboardStats struct {
	TotalUsers    int
	TotalComments int
	BannedUsers   int
}

type CommentView struct {
	ID                 string
	KinopubItemID      int64
	Text               string
	Spoiler            bool
	IsDeleted          bool
	CreatedAtFormatted string
	User               UserView
}

type UserView struct {
	ID                 string
	Username           string
	Avatar             *string
	IsBanned           bool
	CreatedAtFormatted string
}

func NewAdminWebHandler(userService *usecase.UserService, commentService *usecase.CommentService, fs embed.FS, basePath string) *AdminWebHandler {
	funcMap := template.FuncMap{
		"add":      func(a, b int) int { return a + b },
		"subtract": func(a, b int) int { return a - b },
		"slice": func(s string, start, end int) string {
			if len(s) < end {
				end = len(s)
			}
			if start > len(s) {
				return ""
			}
			return s[start:end]
		},
	}

	templates := make(map[string]*template.Template)
	pages := []string{"dashboard", "comments", "users"}

	for _, page := range pages {
		tmpl := template.Must(
			template.New("").Funcs(funcMap).ParseFS(fs, "templates/layout.html", "templates/"+page+".html"),
		)
		templates[page] = tmpl
	}

	return &AdminWebHandler{
		userService:    userService,
		commentService: commentService,
		templates:      templates,
		basePath:       basePath,
	}
}

func (h *AdminWebHandler) Dashboard(c echo.Context) error {
	ctx := c.Request().Context()

	users, totalUsers, _ := h.userService.List(ctx, 1, 1)
	_, totalComments, _ := h.commentService.List(ctx, 1, 1, false)

	bannedCount := 0
	allUsers, _, _ := h.userService.List(ctx, 1, 1000)
	for _, u := range allUsers {
		if u.IsBanned {
			bannedCount++
		}
	}

	recentComments, _, _ := h.commentService.List(ctx, 1, 5, false)
	commentViews := make([]CommentView, 0, len(recentComments))
	for _, co := range recentComments {
		var userView UserView
		if co.User != nil {
			userView = UserView{
				ID:       co.User.ID,
				Username: co.User.KinopubUsername,
				Avatar:   co.User.Avatar,
			}
		}
		commentViews = append(commentViews, CommentView{
			ID:                 co.ID,
			Text:               truncate(co.Text, 100),
			Spoiler:            co.Spoiler,
			IsDeleted:          co.IsDeleted(),
			CreatedAtFormatted: co.CreatedAt.Format("Jan 2, 2006 15:04"),
			User:               userView,
		})
	}

	_ = users

	data := map[string]interface{}{
		"Stats": DashboardStats{
			TotalUsers:    totalUsers,
			TotalComments: totalComments,
			BannedUsers:   bannedCount,
		},
		"RecentComments": commentViews,
	}

	return h.render(c, "dashboard", data)
}

func (h *AdminWebHandler) Comments(c echo.Context) error {
	ctx := c.Request().Context()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage := 20
	search := c.QueryParam("search")
	userFilter := c.QueryParam("user")

	comments, total, _ := h.commentService.List(ctx, page, perPage, true)

	commentViews := make([]CommentView, 0, len(comments))
	for _, co := range comments {
		if search != "" && !containsIgnoreCase(co.Text, search) {
			continue
		}
		var userView UserView
		if co.User != nil {
			if userFilter != "" && !containsIgnoreCase(co.User.KinopubUsername, userFilter) {
				continue
			}
			userView = UserView{
				ID:       co.User.ID,
				Username: co.User.KinopubUsername,
				Avatar:   co.User.Avatar,
			}
		} else if userFilter != "" {
			continue
		}
		commentViews = append(commentViews, CommentView{
			ID:                 co.ID,
			KinopubItemID:      co.KinopubItemID,
			Text:               co.Text,
			Spoiler:            co.Spoiler,
			IsDeleted:          co.IsDeleted(),
			CreatedAtFormatted: co.CreatedAt.Format("Jan 2, 2006 15:04"),
			User:               userView,
		})
	}

	totalPages := (total + perPage - 1) / perPage

	data := map[string]interface{}{
		"Comments":   commentViews,
		"Page":       page,
		"TotalPages": totalPages,
		"TotalItems": total,
		"Search":     search,
		"UserFilter": userFilter,
	}

	return h.render(c, "comments", data)
}

func (h *AdminWebHandler) Users(c echo.Context) error {
	ctx := c.Request().Context()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage := 20

	users, total, _ := h.userService.List(ctx, page, perPage)

	userViews := make([]UserView, 0, len(users))
	for _, u := range users {
		userViews = append(userViews, UserView{
			ID:                 u.ID,
			Username:           u.KinopubUsername,
			Avatar:             u.Avatar,
			IsBanned:           u.IsBanned,
			CreatedAtFormatted: u.CreatedAt.Format("Jan 2, 2006 15:04"),
		})
	}

	totalPages := (total + perPage - 1) / perPage

	data := map[string]interface{}{
		"Users":      userViews,
		"Page":       page,
		"TotalPages": totalPages,
		"TotalItems": total,
	}

	return h.render(c, "users", data)
}

func (h *AdminWebHandler) DeleteComment(c echo.Context) error {
	commentID := c.Param("id")

	referer := c.Request().Referer()
	if referer == "" {
		referer = h.basePath + "/admin/comments"
	}

	if _, err := h.commentService.AdminDelete(c.Request().Context(), commentID); err != nil {
		return c.Redirect(http.StatusSeeOther, h.basePath+"/admin/comments?error=delete_failed")
	}

	return c.Redirect(http.StatusSeeOther, referer)
}

func (h *AdminWebHandler) BanUser(c echo.Context) error {
	userID := c.Param("id")

	if err := h.userService.SetBanned(c.Request().Context(), userID, true); err != nil {
		return c.Redirect(http.StatusSeeOther, h.basePath+"/admin/users?error=ban_failed")
	}

	return c.Redirect(http.StatusSeeOther, h.basePath+"/admin/users")
}

func (h *AdminWebHandler) UnbanUser(c echo.Context) error {
	userID := c.Param("id")

	if err := h.userService.SetBanned(c.Request().Context(), userID, false); err != nil {
		return c.Redirect(http.StatusSeeOther, h.basePath+"/admin/users?error=unban_failed")
	}

	return c.Redirect(http.StatusSeeOther, h.basePath+"/admin/users")
}

func (h *AdminWebHandler) ToggleSpoiler(c echo.Context) error {
	commentID := c.Param("id")

	referer := c.Request().Referer()
	if referer == "" {
		referer = h.basePath + "/admin/comments"
	}

	if _, err := h.commentService.AdminToggleSpoiler(c.Request().Context(), commentID); err != nil {
		return c.Redirect(http.StatusSeeOther, h.basePath+"/admin/comments?error=toggle_failed")
	}

	return c.Redirect(http.StatusSeeOther, referer)
}

func (h *AdminWebHandler) render(c echo.Context, name string, data map[string]interface{}) error {
	tmpl, ok := h.templates[name]
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "template not found")
	}

	data["BasePath"] = h.basePath

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().WriteHeader(http.StatusOK)

	if err := tmpl.ExecuteTemplate(c.Response().Writer, "layout.html", data); err != nil {
		return err
	}
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

var _ = time.Now
