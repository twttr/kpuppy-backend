package handler

import (
	"bytes"
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

const validHash = "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5"

func setupUserHandlerTest(t *testing.T) (*UserHandler, func()) {
	tmpFile, err := os.CreateTemp("", "test_handler_*.db")
	require.NoError(t, err)
	tmpFile.Close()

	db, err := sqlite.NewDB(tmpFile.Name())
	require.NoError(t, err)

	err = sqlite.RunMigrations(db)
	require.NoError(t, err)

	userRepo := sqlite.NewUserRepository(db)
	userService := usecase.NewUserService(userRepo)
	handler := NewUserHandler(userService)

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return handler, cleanup
}

func TestUserHandler_Provision_NewUser(t *testing.T) {
	handler, cleanup := setupUserHandlerTest(t)
	defer cleanup()

	e := echo.New()

	reqBody := domain.ProvisionRequest{
		UserHash: validHash,
		Avatar:   strPtr("https://example.com/avatar.jpg"),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users/provision", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Provision(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp domain.ProvisionResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.UserID)
	assert.NotEmpty(t, resp.DisplayName)
}

func TestUserHandler_Provision_NewUser_DisplayNameReturned(t *testing.T) {
	handler, cleanup := setupUserHandlerTest(t)
	defer cleanup()

	e := echo.New()

	reqBody := domain.ProvisionRequest{
		UserHash: validHash,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users/provision", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Provision(c)
	require.NoError(t, err)

	var resp domain.ProvisionResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	expectedDisplayName := domain.GeneratePseudonym(validHash)
	assert.Equal(t, expectedDisplayName, resp.DisplayName)
}

func TestUserHandler_Provision_ExistingUser(t *testing.T) {
	handler, cleanup := setupUserHandlerTest(t)
	defer cleanup()

	e := echo.New()

	reqBody := domain.ProvisionRequest{
		UserHash: validHash,
	}
	body, _ := json.Marshal(reqBody)

	req1 := httptest.NewRequest(http.MethodPost, "/users/provision", bytes.NewReader(body))
	req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	err := handler.Provision(c1)
	require.NoError(t, err)

	var resp1 domain.ProvisionResponse
	json.Unmarshal(rec1.Body.Bytes(), &resp1)

	req2 := httptest.NewRequest(http.MethodPost, "/users/provision", bytes.NewReader(body))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	err = handler.Provision(c2)
	require.NoError(t, err)

	var resp2 domain.ProvisionResponse
	json.Unmarshal(rec2.Body.Bytes(), &resp2)

	assert.Equal(t, resp1.UserID, resp2.UserID)
	assert.Equal(t, resp1.DisplayName, resp2.DisplayName)
}

func TestUserHandler_Provision_EmptyHash(t *testing.T) {
	handler, cleanup := setupUserHandlerTest(t)
	defer cleanup()

	e := echo.New()

	reqBody := domain.ProvisionRequest{
		UserHash: "",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users/provision", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Provision(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var apiErr domain.APIError
	json.Unmarshal(rec.Body.Bytes(), &apiErr)
	assert.Equal(t, domain.ErrUserHashEmpty.Error(), apiErr.Error)
}

func TestUserHandler_Provision_InvalidHashLength(t *testing.T) {
	handler, cleanup := setupUserHandlerTest(t)
	defer cleanup()

	e := echo.New()

	reqBody := domain.ProvisionRequest{
		UserHash: "tooshort",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users/provision", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Provision(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var apiErr domain.APIError
	json.Unmarshal(rec.Body.Bytes(), &apiErr)
	assert.Equal(t, domain.ErrInvalidUserHash.Error(), apiErr.Error)
}

func TestUserHandler_Provision_InvalidJSON(t *testing.T) {
	handler, cleanup := setupUserHandlerTest(t)
	defer cleanup()

	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/users/provision", bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Provision(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func strPtr(s string) *string {
	return &s
}
