package handler

import (
	"github.com/dip-roy/go-backend/internal/service"
	"github.com/dip-roy/go-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type registerRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type authResponse struct {
	User         userResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body registerRequest true "Register input"
// @Success 201 {object} response.Response{data=authResponse}
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, formatValidationErrors(h.validate.Struct(req)))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, formatValidationErrors(err))
		return
	}

	out, err := h.authService.Register(c.Request.Context(), service.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.Err(c, err)
		return
	}

	response.Created(c, authResponse{
		User:         toUserResponse(out.User),
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	})
}

// Login godoc
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param body body loginRequest true "Login input"
// @Success 200 {object} response.Response{data=authResponse}
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, formatValidationErrors(err))
		return
	}

	out, err := h.authService.Login(c.Request.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, authResponse{
		User:         toUserResponse(out.User),
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	})
}

// Refresh godoc
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body refreshRequest true "Refresh token"
// @Success 200 {object} response.Response{data=authResponse}
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, formatValidationErrors(err))
		return
	}

	out, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, authResponse{
		User:         toUserResponse(out.User),
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	})
}

// Logout godoc
// @Summary Logout
// @Tags auth
// @Accept json
// @Produce json
// @Param body body logoutRequest true "Refresh token to revoke"
// @Success 200 {object} response.Response
// @Security BearerAuth
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, formatValidationErrors(err))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, gin.H{"message": "logged out"})
}
