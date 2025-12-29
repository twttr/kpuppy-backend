package handler

import (
	"bytes"
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
	"github.com/twttr/kpuppy-backend/internal/websocket"
)

func setupCommentHandlerTest(t *testing.T) (*CommentHandler, *usecase.UserService, func()) {
	tmpFile, err := os.CreateTemp("", "test_comment_handler_*.db")
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

	hub := websocket.NewHub()

	handler := NewCommentHandler(commentService, hub)

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return handler, userService, cleanup
}

func provisionUser(t *testing.T, userService *usecase.UserService, username string) *domain.User {
	user, err := userService.Provision(context.Background(), &domain.ProvisionRequest{Username: username})
	require.NoError(t, err)
	return user
}

func TestCommentHandler_GetComments_EmptyContent(t *testing.T) {
	handler, _, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/content/12345/comments", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("kinopubItemId")
	c.SetParamValues("12345")

	err := handler.GetComments(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp domain.CommentsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Empty(t, resp.Comments)
}

func TestCommentHandler_GetComments_InvalidKinopubItemID(t *testing.T) {
	handler, _, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/content/invalid/comments", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("kinopubItemId")
	c.SetParamValues("invalid")

	err := handler.GetComments(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCommentHandler_CreateComment_Success(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	user := provisionUser(t, userService, "testuser")

	e := echo.New()
	reqBody := domain.CreateCommentRequest{
		Text:    "test comment",
		Spoiler: false,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/content/12345/comments", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-ID", user.ID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("kinopubItemId")
	c.SetParamValues("12345")

	err := handler.CreateComment(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp domain.CommentResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "test comment", resp.Text)
	assert.NotEmpty(t, resp.ID)
}

func TestCommentHandler_CreateComment_NoUserID(t *testing.T) {
	handler, _, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	e := echo.New()
	reqBody := domain.CreateCommentRequest{
		Text: "test comment",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/content/12345/comments", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("kinopubItemId")
	c.SetParamValues("12345")

	err := handler.CreateComment(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCommentHandler_CreateComment_EmptyText(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	user := provisionUser(t, userService, "testuser")

	e := echo.New()
	reqBody := domain.CreateCommentRequest{
		Text: "",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/content/12345/comments", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-ID", user.ID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("kinopubItemId")
	c.SetParamValues("12345")

	err := handler.CreateComment(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCommentHandler_ReplyToComment_Success(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	user := provisionUser(t, userService, "testuser")

	e := echo.New()

	createReqBody := domain.CreateCommentRequest{Text: "parent comment"}
	createBody, _ := json.Marshal(createReqBody)
	createReq := httptest.NewRequest(http.MethodPost, "/content/12345/comments", bytes.NewReader(createBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set("X-User-ID", user.ID)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)
	createCtx.SetParamNames("kinopubItemId")
	createCtx.SetParamValues("12345")

	err := handler.CreateComment(createCtx)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var parentComment domain.CommentResponse
	json.Unmarshal(createRec.Body.Bytes(), &parentComment)

	replyReqBody := domain.CreateCommentRequest{Text: "reply comment"}
	replyBody, _ := json.Marshal(replyReqBody)
	replyReq := httptest.NewRequest(http.MethodPost, "/comments/"+parentComment.ID+"/reply", bytes.NewReader(replyBody))
	replyReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	replyReq.Header.Set("X-User-ID", user.ID)
	replyRec := httptest.NewRecorder()
	replyCtx := e.NewContext(replyReq, replyRec)
	replyCtx.SetParamNames("commentId")
	replyCtx.SetParamValues(parentComment.ID)

	err = handler.ReplyToComment(replyCtx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, replyRec.Code)

	var reply domain.CommentResponse
	json.Unmarshal(replyRec.Body.Bytes(), &reply)
	assert.Equal(t, "reply comment", reply.Text)
}

func TestCommentHandler_UpdateComment_Success(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	user := provisionUser(t, userService, "testuser")

	e := echo.New()

	createReqBody := domain.CreateCommentRequest{Text: "original text"}
	createBody, _ := json.Marshal(createReqBody)
	createReq := httptest.NewRequest(http.MethodPost, "/content/12345/comments", bytes.NewReader(createBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set("X-User-ID", user.ID)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)
	createCtx.SetParamNames("kinopubItemId")
	createCtx.SetParamValues("12345")

	err := handler.CreateComment(createCtx)
	require.NoError(t, err)

	var comment domain.CommentResponse
	json.Unmarshal(createRec.Body.Bytes(), &comment)

	spoiler := true
	updateReqBody := domain.UpdateCommentRequest{Text: "updated text", Spoiler: &spoiler}
	updateBody, _ := json.Marshal(updateReqBody)
	updateReq := httptest.NewRequest(http.MethodPatch, "/comments/"+comment.ID, bytes.NewReader(updateBody))
	updateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	updateReq.Header.Set("X-User-ID", user.ID)
	updateRec := httptest.NewRecorder()
	updateCtx := e.NewContext(updateReq, updateRec)
	updateCtx.SetParamNames("commentId")
	updateCtx.SetParamValues(comment.ID)

	err = handler.UpdateComment(updateCtx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, updateRec.Code)

	var updated domain.CommentResponse
	json.Unmarshal(updateRec.Body.Bytes(), &updated)
	assert.Equal(t, "updated text", updated.Text)
	assert.True(t, updated.Spoiler)
}

func TestCommentHandler_UpdateComment_NotOwner(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	user1 := provisionUser(t, userService, "user1")
	user2 := provisionUser(t, userService, "user2")

	e := echo.New()

	createReqBody := domain.CreateCommentRequest{Text: "original text"}
	createBody, _ := json.Marshal(createReqBody)
	createReq := httptest.NewRequest(http.MethodPost, "/content/12345/comments", bytes.NewReader(createBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set("X-User-ID", user1.ID)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)
	createCtx.SetParamNames("kinopubItemId")
	createCtx.SetParamValues("12345")

	err := handler.CreateComment(createCtx)
	require.NoError(t, err)

	var comment domain.CommentResponse
	json.Unmarshal(createRec.Body.Bytes(), &comment)

	updateReqBody := domain.UpdateCommentRequest{Text: "hacked text"}
	updateBody, _ := json.Marshal(updateReqBody)
	updateReq := httptest.NewRequest(http.MethodPatch, "/comments/"+comment.ID, bytes.NewReader(updateBody))
	updateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	updateReq.Header.Set("X-User-ID", user2.ID)
	updateRec := httptest.NewRecorder()
	updateCtx := e.NewContext(updateReq, updateRec)
	updateCtx.SetParamNames("commentId")
	updateCtx.SetParamValues(comment.ID)

	err = handler.UpdateComment(updateCtx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusForbidden, updateRec.Code)
}

func TestCommentHandler_DeleteComment_Success(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	user := provisionUser(t, userService, "testuser")

	e := echo.New()

	createReqBody := domain.CreateCommentRequest{Text: "to be deleted"}
	createBody, _ := json.Marshal(createReqBody)
	createReq := httptest.NewRequest(http.MethodPost, "/content/12345/comments", bytes.NewReader(createBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set("X-User-ID", user.ID)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)
	createCtx.SetParamNames("kinopubItemId")
	createCtx.SetParamValues("12345")

	err := handler.CreateComment(createCtx)
	require.NoError(t, err)

	var comment domain.CommentResponse
	json.Unmarshal(createRec.Body.Bytes(), &comment)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/comments/"+comment.ID, nil)
	deleteReq.Header.Set("X-User-ID", user.ID)
	deleteRec := httptest.NewRecorder()
	deleteCtx := e.NewContext(deleteReq, deleteRec)
	deleteCtx.SetParamNames("commentId")
	deleteCtx.SetParamValues(comment.ID)

	err = handler.DeleteComment(deleteCtx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, deleteRec.Code)
}

func TestCommentHandler_DeleteComment_NotOwner(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	user1 := provisionUser(t, userService, "user1")
	user2 := provisionUser(t, userService, "user2")

	e := echo.New()

	createReqBody := domain.CreateCommentRequest{Text: "to be deleted"}
	createBody, _ := json.Marshal(createReqBody)
	createReq := httptest.NewRequest(http.MethodPost, "/content/12345/comments", bytes.NewReader(createBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set("X-User-ID", user1.ID)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)
	createCtx.SetParamNames("kinopubItemId")
	createCtx.SetParamValues("12345")

	err := handler.CreateComment(createCtx)
	require.NoError(t, err)

	var comment domain.CommentResponse
	json.Unmarshal(createRec.Body.Bytes(), &comment)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/comments/"+comment.ID, nil)
	deleteReq.Header.Set("X-User-ID", user2.ID)
	deleteRec := httptest.NewRecorder()
	deleteCtx := e.NewContext(deleteReq, deleteRec)
	deleteCtx.SetParamNames("commentId")
	deleteCtx.SetParamValues(comment.ID)

	err = handler.DeleteComment(deleteCtx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusForbidden, deleteRec.Code)
}

func TestCommentHandler_GetComments_ReturnsAllComments(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	user := provisionUser(t, userService, "testuser")

	e := echo.New()

	for i := 0; i < 5; i++ {
		reqBody := domain.CreateCommentRequest{Text: "comment"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/content/12345/comments", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("X-User-ID", user.ID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("kinopubItemId")
		c.SetParamValues("12345")
		handler.CreateComment(c)
	}

	req := httptest.NewRequest(http.MethodGet, "/content/12345/comments", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("kinopubItemId")
	c.SetParamValues("12345")

	err := handler.GetComments(c)
	require.NoError(t, err)

	var resp domain.CommentsResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	assert.Len(t, resp.Comments, 5)
}

func TestCommentHandler_CreateComment_BannedUser(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	ctx := context.Background()
	user := provisionUser(t, userService, "testuser")
	userService.SetBanned(ctx, user.ID, true)

	e := echo.New()
	reqBody := domain.CreateCommentRequest{Text: "test comment"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/content/12345/comments", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-ID", user.ID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("kinopubItemId")
	c.SetParamValues("12345")

	err := handler.CreateComment(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCommentHandler_CreateComment_UserNotFound(t *testing.T) {
	handler, _, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	e := echo.New()
	reqBody := domain.CreateCommentRequest{Text: "test comment"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/content/12345/comments", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-ID", "nonexistent-user")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("kinopubItemId")
	c.SetParamValues("12345")

	err := handler.CreateComment(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCommentHandler_CreateComment_CommentTooLong(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	user := provisionUser(t, userService, "testuser")

	e := echo.New()
	longText := make([]byte, 1001)
	for i := range longText {
		longText[i] = 'a'
	}
	reqBody := domain.CreateCommentRequest{Text: string(longText)}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/content/12345/comments", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-ID", user.ID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("kinopubItemId")
	c.SetParamValues("12345")

	err := handler.CreateComment(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCommentHandler_ReplyToComment_NoUserID(t *testing.T) {
	handler, _, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	e := echo.New()
	reqBody := domain.CreateCommentRequest{Text: "reply"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/comments/some-id/reply", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("commentId")
	c.SetParamValues("some-id")

	err := handler.ReplyToComment(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCommentHandler_ReplyToComment_ParentNotFound(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	user := provisionUser(t, userService, "testuser")

	e := echo.New()
	reqBody := domain.CreateCommentRequest{Text: "reply"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/comments/nonexistent/reply", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-ID", user.ID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("commentId")
	c.SetParamValues("nonexistent")

	err := handler.ReplyToComment(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCommentHandler_ReplyToComment_NestedReplies(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	user := provisionUser(t, userService, "testuser")

	e := echo.New()

	createReqBody := domain.CreateCommentRequest{Text: "parent comment"}
	createBody, _ := json.Marshal(createReqBody)
	createReq := httptest.NewRequest(http.MethodPost, "/content/12345/comments", bytes.NewReader(createBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set("X-User-ID", user.ID)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)
	createCtx.SetParamNames("kinopubItemId")
	createCtx.SetParamValues("12345")
	handler.CreateComment(createCtx)

	var parentComment domain.CommentResponse
	json.Unmarshal(createRec.Body.Bytes(), &parentComment)

	replyReqBody := domain.CreateCommentRequest{Text: "reply"}
	replyBody, _ := json.Marshal(replyReqBody)
	replyReq := httptest.NewRequest(http.MethodPost, "/comments/"+parentComment.ID+"/reply", bytes.NewReader(replyBody))
	replyReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	replyReq.Header.Set("X-User-ID", user.ID)
	replyRec := httptest.NewRecorder()
	replyCtx := e.NewContext(replyReq, replyRec)
	replyCtx.SetParamNames("commentId")
	replyCtx.SetParamValues(parentComment.ID)
	handler.ReplyToComment(replyCtx)

	var reply domain.CommentResponse
	json.Unmarshal(replyRec.Body.Bytes(), &reply)

	nestedReplyBody, _ := json.Marshal(domain.CreateCommentRequest{Text: "nested reply"})
	nestedReplyReq := httptest.NewRequest(http.MethodPost, "/comments/"+reply.ID+"/reply", bytes.NewReader(nestedReplyBody))
	nestedReplyReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	nestedReplyReq.Header.Set("X-User-ID", user.ID)
	nestedReplyRec := httptest.NewRecorder()
	nestedReplyCtx := e.NewContext(nestedReplyReq, nestedReplyRec)
	nestedReplyCtx.SetParamNames("commentId")
	nestedReplyCtx.SetParamValues(reply.ID)

	err := handler.ReplyToComment(nestedReplyCtx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, nestedReplyRec.Code)

	var nestedReply domain.CommentResponse
	json.Unmarshal(nestedReplyRec.Body.Bytes(), &nestedReply)
	assert.Equal(t, "nested reply", nestedReply.Text)
	assert.Equal(t, reply.ID, *nestedReply.ParentID)
}

func TestCommentHandler_UpdateComment_NoUserID(t *testing.T) {
	handler, _, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	e := echo.New()
	reqBody := domain.UpdateCommentRequest{Text: "updated"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/comments/some-id", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("commentId")
	c.SetParamValues("some-id")

	err := handler.UpdateComment(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCommentHandler_UpdateComment_NotFound(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	user := provisionUser(t, userService, "testuser")

	e := echo.New()
	reqBody := domain.UpdateCommentRequest{Text: "updated"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/comments/nonexistent", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-User-ID", user.ID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("commentId")
	c.SetParamValues("nonexistent")

	err := handler.UpdateComment(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCommentHandler_DeleteComment_NoUserID(t *testing.T) {
	handler, _, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	e := echo.New()

	req := httptest.NewRequest(http.MethodDelete, "/comments/some-id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("commentId")
	c.SetParamValues("some-id")

	err := handler.DeleteComment(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCommentHandler_DeleteComment_NotFound(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	user := provisionUser(t, userService, "testuser")

	e := echo.New()

	req := httptest.NewRequest(http.MethodDelete, "/comments/nonexistent", nil)
	req.Header.Set("X-User-ID", user.ID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("commentId")
	c.SetParamValues("nonexistent")

	err := handler.DeleteComment(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCommentHandler_DeleteComment_AlreadyDeleted(t *testing.T) {
	handler, userService, cleanup := setupCommentHandlerTest(t)
	defer cleanup()

	user := provisionUser(t, userService, "testuser")

	e := echo.New()

	createReqBody := domain.CreateCommentRequest{Text: "to delete"}
	createBody, _ := json.Marshal(createReqBody)
	createReq := httptest.NewRequest(http.MethodPost, "/content/12345/comments", bytes.NewReader(createBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set("X-User-ID", user.ID)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)
	createCtx.SetParamNames("kinopubItemId")
	createCtx.SetParamValues("12345")
	handler.CreateComment(createCtx)

	var comment domain.CommentResponse
	json.Unmarshal(createRec.Body.Bytes(), &comment)

	deleteReq1 := httptest.NewRequest(http.MethodDelete, "/comments/"+comment.ID, nil)
	deleteReq1.Header.Set("X-User-ID", user.ID)
	deleteRec1 := httptest.NewRecorder()
	deleteCtx1 := e.NewContext(deleteReq1, deleteRec1)
	deleteCtx1.SetParamNames("commentId")
	deleteCtx1.SetParamValues(comment.ID)
	handler.DeleteComment(deleteCtx1)

	deleteReq2 := httptest.NewRequest(http.MethodDelete, "/comments/"+comment.ID, nil)
	deleteReq2.Header.Set("X-User-ID", user.ID)
	deleteRec2 := httptest.NewRecorder()
	deleteCtx2 := e.NewContext(deleteReq2, deleteRec2)
	deleteCtx2.SetParamNames("commentId")
	deleteCtx2.SetParamValues(comment.ID)

	err := handler.DeleteComment(deleteCtx2)
	require.NoError(t, err)

	assert.Equal(t, http.StatusGone, deleteRec2.Code)
}
