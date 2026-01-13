package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twttr/kpuppy-backend/internal/domain"
	"github.com/twttr/kpuppy-backend/internal/repository/sqlite"
	"github.com/twttr/kpuppy-backend/internal/usecase"
)

func setupAdminHandlerTest(t *testing.T) (*AdminHandler, *usecase.UserService, *usecase.CommentService, func()) {
	tmpFile, err := os.CreateTemp("", "test_admin_handler_*.db")
	require.NoError(t, err)
	tmpFile.Close()

	db, err := sqlite.NewDB(tmpFile.Name())
	require.NoError(t, err)

	err = sqlite.RunMigrations(db)
	require.NoError(t, err)

	userRepo := sqlite.NewUserRepository(db)
	contentRepo := sqlite.NewContentRepository(db)
	commentRepo := sqlite.NewCommentRepository(db)

	userService := usecase.NewUserService(userRepo)
	commentService := usecase.NewCommentService(commentRepo, contentRepo, userRepo)

	handler := NewAdminHandler(userService, commentService)

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return handler, userService, commentService, cleanup
}

func TestAdminHandler_ListUsers(t *testing.T) {
	handler, userService, _, cleanup := setupAdminHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	userService.Provision(ctx, &domain.ProvisionRequest{UserHash: "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5"})
	userService.Provision(ctx, &domain.ProvisionRequest{UserHash: "b7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b6"})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/api/users?page=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ListUsers(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	users := resp["users"].([]interface{})
	assert.Len(t, users, 2)
}

func TestAdminHandler_ListUsers_DefaultPage(t *testing.T) {
	handler, _, _, cleanup := setupAdminHandlerTest(t)
	defer cleanup()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/api/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ListUsers(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminHandler_BanUser_Success(t *testing.T) {
	handler, userService, _, cleanup := setupAdminHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	user, _ := userService.Provision(ctx, &domain.ProvisionRequest{UserHash: "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5"})

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/admin/api/users/"+user.ID+"/ban", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(user.ID)

	err := handler.BanUser(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "banned", resp["status"])
}

func TestAdminHandler_BanUser_NotFound(t *testing.T) {
	handler, _, _, cleanup := setupAdminHandlerTest(t)
	defer cleanup()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/admin/api/users/nonexistent/ban", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := handler.BanUser(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminHandler_UnbanUser_Success(t *testing.T) {
	handler, userService, _, cleanup := setupAdminHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	user, _ := userService.Provision(ctx, &domain.ProvisionRequest{UserHash: "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5"})
	userService.SetBanned(ctx, user.ID, true)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/admin/api/users/"+user.ID+"/unban", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(user.ID)

	err := handler.UnbanUser(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "unbanned", resp["status"])
}

func TestAdminHandler_UnbanUser_NotFound(t *testing.T) {
	handler, _, _, cleanup := setupAdminHandlerTest(t)
	defer cleanup()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/admin/api/users/nonexistent/unban", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := handler.UnbanUser(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminHandler_ListComments(t *testing.T) {
	handler, userService, commentService, cleanup := setupAdminHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	user, _ := userService.Provision(ctx, &domain.ProvisionRequest{UserHash: "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5"})

	commentService.CreateComment(ctx, 12345, user.ID, &domain.CreateCommentRequest{Text: "comment 1"})
	commentService.CreateComment(ctx, 12345, user.ID, &domain.CreateCommentRequest{Text: "comment 2"})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/api/comments?page=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ListComments(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	comments := resp["comments"].([]interface{})
	assert.Len(t, comments, 2)
}

func TestAdminHandler_ListComments_ShowsDeleted(t *testing.T) {
	handler, userService, commentService, cleanup := setupAdminHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	user, _ := userService.Provision(ctx, &domain.ProvisionRequest{UserHash: "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5"})

	comment, _ := commentService.CreateComment(ctx, 12345, user.ID, &domain.CreateCommentRequest{Text: "to delete"})
	commentService.AdminDelete(ctx, comment.ID)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/api/comments", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ListComments(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	comments := resp["comments"].([]interface{})
	assert.Len(t, comments, 1)
}

func TestAdminHandler_DeleteComment_Success(t *testing.T) {
	handler, userService, commentService, cleanup := setupAdminHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	user, _ := userService.Provision(ctx, &domain.ProvisionRequest{UserHash: "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5"})
	comment, _ := commentService.CreateComment(ctx, 12345, user.ID, &domain.CreateCommentRequest{Text: "test"})

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/admin/api/comments/"+comment.ID, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(comment.ID)

	err := handler.DeleteComment(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "deleted", resp["status"])
}

func TestAdminHandler_DeleteComment_NotFound(t *testing.T) {
	handler, _, _, cleanup := setupAdminHandlerTest(t)
	defer cleanup()

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/admin/api/comments/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := handler.DeleteComment(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
