package handler

import (
	"github.com/dip-roy/go-backend/internal/service"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	authService service.AuthService
	userService service.UserService
	validate    *validator.Validate
}

func New(authSvc service.AuthService, userSvc service.UserService) *Handler {
	return &Handler{
		authService: authSvc,
		userService: userSvc,
		validate:    validator.New(),
	}
}

type validationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (h *Handler) bindAndValidate(v interface{}, obj interface{}) []validationError {
	if errs, ok := obj.(validator.ValidationErrors); ok {
		var out []validationError
		for _, e := range errs {
			out = append(out, validationError{Field: e.Field(), Message: e.Tag()})
		}
		return out
	}
	return nil
}

func formatValidationErrors(err error) []validationError {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil
	}
	var out []validationError
	for _, e := range errs {
		out = append(out, validationError{Field: e.Field(), Message: e.Tag()})
	}
	return out
}
