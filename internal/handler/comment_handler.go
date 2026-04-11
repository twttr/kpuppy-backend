package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/twttr/kpuppy-backend/internal/domain"
	custommw "github.com/twttr/kpuppy-backend/internal/middleware"
	"github.com/twttr/kpuppy-backend/internal/usecase"
	"github.com/twttr/kpuppy-backend/internal/websocket"
)

type CommentHandler struct {
	commentService *usecase.CommentService
	hub            *websocket.Hub
}

func NewCommentHandler(commentService *usecase.CommentService, hub *websocket.Hub) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		hub:            hub,
	}
}

func (h *CommentHandler) GetComments(c echo.Context) error {
	kinopubItemID, err := strconv.ParseInt(c.Param("kinopubItemId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, domain.NewAPIError(err, domain.CodeValidationError))
	}

	response, err := h.commentService.GetComments(c.Request().Context(), kinopubItemID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, domain.NewAPIError(err, domain.CodeValidationError))
	}

	return c.JSON(http.StatusOK, response)
}

func (h *CommentHandler) CreateComment(c echo.Context) error {
	kinopubItemID, err := strconv.ParseInt(c.Param("kinopubItemId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, domain.NewAPIError(err, domain.CodeValidationError))
	}

	// Fix #7: user ID comes from middleware (verified via X-User-Hash), not client header
	userID, ok := custommw.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, domain.NewAPIError(domain.ErrUserNotProvisioned, domain.CodeUserNotProvisioned))
	}

	var req domain.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.NewAPIError(err, domain.CodeValidationError))
	}
	req.UserID = userID

	comment, err := h.commentService.CreateComment(c.Request().Context(), kinopubItemID, userID, &req)
	if err != nil {
		return h.handleError(c, err)
	}

	roomID := strconv.FormatInt(kinopubItemID, 10)
	h.hub.BroadcastToRoom(roomID, websocket.Message{
		Event: "comment_created",
		Data:  comment.ToResponse(),
	})

	return c.JSON(http.StatusCreated, comment.ToResponse())
}

func (h *CommentHandler) ReplyToComment(c echo.Context) error {
	commentID := c.Param("commentId")

	// Fix #7: user ID comes from middleware (verified via X-User-Hash), not client header
	userID, ok := custommw.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, domain.NewAPIError(domain.ErrUserNotProvisioned, domain.CodeUserNotProvisioned))
	}

	var req domain.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.NewAPIError(err, domain.CodeValidationError))
	}
	req.UserID = userID

	comment, err := h.commentService.ReplyToComment(c.Request().Context(), commentID, userID, &req)
	if err != nil {
		return h.handleError(c, err)
	}

	parent, _ := h.commentService.GetByID(c.Request().Context(), commentID)
	if parent != nil {
		content, _ := h.getContentByCommentID(c, parent.ContentID)
		if content != nil {
			roomID := strconv.FormatInt(content.KinopubItemID, 10)
			h.hub.BroadcastToRoom(roomID, websocket.Message{
				Event: "comment_created",
				Data:  comment.ToResponse(),
			})
		}
	}

	return c.JSON(http.StatusCreated, comment.ToResponse())
}

func (h *CommentHandler) UpdateComment(c echo.Context) error {
	commentID := c.Param("commentId")

	// Fix #7: user ID comes from middleware (verified via X-User-Hash), not client header
	userID, ok := custommw.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, domain.NewAPIError(domain.ErrUserNotProvisioned, domain.CodeUserNotProvisioned))
	}

	var req domain.UpdateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.NewAPIError(err, domain.CodeValidationError))
	}
	req.UserID = userID

	comment, err := h.commentService.UpdateComment(c.Request().Context(), commentID, userID, &req)
	if err != nil {
		return h.handleError(c, err)
	}

	content, _ := h.getContentByCommentID(c, comment.ContentID)
	if content != nil {
		roomID := strconv.FormatInt(content.KinopubItemID, 10)
		h.hub.BroadcastToRoom(roomID, websocket.Message{
			Event: "comment_updated",
			Data:  comment.ToResponse(),
		})
	}

	return c.JSON(http.StatusOK, comment.ToResponse())
}

func (h *CommentHandler) DeleteComment(c echo.Context) error {
	commentID := c.Param("commentId")

	// Fix #7: user ID comes from middleware (verified via X-User-Hash), not client header
	userID, ok := custommw.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, domain.NewAPIError(domain.ErrUserNotProvisioned, domain.CodeUserNotProvisioned))
	}

	comment, err := h.commentService.GetByID(c.Request().Context(), commentID)
	if err != nil {
		return h.handleError(c, err)
	}

	contentID := comment.ContentID

	if err := h.commentService.DeleteComment(c.Request().Context(), commentID, userID); err != nil {
		return h.handleError(c, err)
	}

	content, _ := h.getContentByCommentID(c, contentID)
	if content != nil {
		roomID := strconv.FormatInt(content.KinopubItemID, 10)
		h.hub.BroadcastToRoom(roomID, websocket.Message{
			Event: "comment_deleted",
			Data:  map[string]string{"id": commentID},
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *CommentHandler) getContentByCommentID(c echo.Context, contentID string) (*domain.Content, error) {
	return h.commentService.GetContentByID(c.Request().Context(), contentID)
}

func (h *CommentHandler) handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return c.JSON(http.StatusUnauthorized, domain.NewAPIError(err, domain.CodeUserNotProvisioned))
	case errors.Is(err, domain.ErrUserBanned):
		return c.JSON(http.StatusForbidden, domain.NewAPIError(err, domain.CodeUserBanned))
	case errors.Is(err, domain.ErrCommentNotFound):
		return c.JSON(http.StatusNotFound, domain.NewAPIError(err, domain.CodeCommentNotFound))
	case errors.Is(err, domain.ErrContentNotFound):
		return c.JSON(http.StatusNotFound, domain.NewAPIError(err, domain.CodeContentNotFound))
	case errors.Is(err, domain.ErrNotCommentOwner):
		return c.JSON(http.StatusForbidden, domain.NewAPIError(err, domain.CodeForbidden))
	case errors.Is(err, domain.ErrCommentDeleted):
		return c.JSON(http.StatusGone, domain.NewAPIError(err, domain.CodeCommentNotFound))
	case errors.Is(err, domain.ErrCommentTooLong), errors.Is(err, domain.ErrCommentEmpty):
		return c.JSON(http.StatusBadRequest, domain.NewAPIError(err, domain.CodeValidationError))
	default:
		return c.JSON(http.StatusInternalServerError, domain.NewAPIError(err, domain.CodeValidationError))
	}
}
