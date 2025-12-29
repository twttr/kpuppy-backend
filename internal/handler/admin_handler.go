package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/twttr/kpuppy-backend/internal/domain"
	"github.com/twttr/kpuppy-backend/internal/usecase"
)

type AdminHandler struct {
	userService    *usecase.UserService
	commentService *usecase.CommentService
}

func NewAdminHandler(userService *usecase.UserService, commentService *usecase.CommentService) *AdminHandler {
	return &AdminHandler{
		userService:    userService,
		commentService: commentService,
	}
}

func (h *AdminHandler) ListUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage := 20

	users, total, err := h.userService.List(c.Request().Context(), page, perPage)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, domain.NewAPIError(err, domain.CodeValidationError))
	}

	totalPages := (total + perPage - 1) / perPage

	responses := make([]domain.UserResponse, 0, len(users))
	for _, u := range users {
		responses = append(responses, u.ToResponse())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"users": responses,
		"pagination": domain.PaginationResponse{
			Page:       page,
			TotalPages: totalPages,
			TotalItems: total,
			HasMore:    page < totalPages,
		},
	})
}

func (h *AdminHandler) BanUser(c echo.Context) error {
	userID := c.Param("id")
	if err := h.userService.SetBanned(c.Request().Context(), userID, true); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, domain.NewAPIError(err, domain.CodeValidationError))
		}
		return c.JSON(http.StatusInternalServerError, domain.NewAPIError(err, domain.CodeValidationError))
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "banned"})
}

func (h *AdminHandler) UnbanUser(c echo.Context) error {
	userID := c.Param("id")
	if err := h.userService.SetBanned(c.Request().Context(), userID, false); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, domain.NewAPIError(err, domain.CodeValidationError))
		}
		return c.JSON(http.StatusInternalServerError, domain.NewAPIError(err, domain.CodeValidationError))
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "unbanned"})
}

func (h *AdminHandler) ListComments(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage := 20

	comments, total, err := h.commentService.List(c.Request().Context(), page, perPage, true)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, domain.NewAPIError(err, domain.CodeValidationError))
	}

	totalPages := (total + perPage - 1) / perPage

	responses := make([]domain.CommentResponse, 0, len(comments))
	for _, co := range comments {
		responses = append(responses, co.ToResponse())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"comments": responses,
		"pagination": domain.PaginationResponse{
			Page:       page,
			TotalPages: totalPages,
			TotalItems: total,
			HasMore:    page < totalPages,
		},
	})
}

func (h *AdminHandler) DeleteComment(c echo.Context) error {
	commentID := c.Param("id")
	comment, err := h.commentService.AdminDelete(c.Request().Context(), commentID)
	if err != nil {
		if errors.Is(err, domain.ErrCommentNotFound) {
			return c.JSON(http.StatusNotFound, domain.NewAPIError(err, domain.CodeCommentNotFound))
		}
		return c.JSON(http.StatusInternalServerError, domain.NewAPIError(err, domain.CodeValidationError))
	}
	status := "deleted"
	if !comment.IsDeleted() {
		status = "restored"
	}
	return c.JSON(http.StatusOK, map[string]string{"status": status})
}
