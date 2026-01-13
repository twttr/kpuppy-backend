package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/twttr/kpuppy-backend/internal/domain"
	"github.com/twttr/kpuppy-backend/internal/usecase"
)

type UserHandler struct {
	userService *usecase.UserService
}

func NewUserHandler(userService *usecase.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Provision(c echo.Context) error {
	var req domain.ProvisionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.NewAPIError(err, domain.CodeValidationError))
	}

	user, err := h.userService.Provision(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, domain.NewAPIError(err, domain.CodeValidationError))
	}

	return c.JSON(http.StatusOK, domain.ProvisionResponse{UserID: user.ID, DisplayName: user.DisplayName})
}
