package controllers

import (
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/internal/services"
	"SalaryAdvance/pkg/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authUseCase domain.AuthUseCase
	rateLimiter *services.LoginRateLimiter
}

func NewAuthController(uc domain.AuthUseCase, rl *services.LoginRateLimiter) *AuthController {
	return &AuthController{authUseCase: uc,
		rateLimiter: rl,
	}
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(config.GetStatusCode(config.ErrBadRequest), gin.H{"error": err.Error()})
		return
	}
	allowed, err := ctrl.rateLimiter.CheckAndIncrement(req.Email)
	if !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many login attempts, try again later"})
		return
	}
	access, refresh, err := ctrl.authUseCase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(config.GetStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	ctrl.rateLimiter.Reset(req.Email)
	c.JSON(http.StatusOK, gin.H{"access_token": access, "refresh_token": refresh})
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(config.GetStatusCode(config.ErrBadRequest), gin.H{"error": err.Error()})
		return
	}
	err := ctrl.authUseCase.RegisterFromInvite(c.Request.Context(), req.Token, req.Password)
	if err != nil {
		c.JSON(config.GetStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "registered successfully"})
}

func (ctrl *AuthController) Refresh(c *gin.Context) {
	var req domain.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(config.GetStatusCode(config.ErrBadRequest), gin.H{"error": err.Error()})
		return
	}
	access, err := ctrl.authUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(config.GetStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": access})
}
