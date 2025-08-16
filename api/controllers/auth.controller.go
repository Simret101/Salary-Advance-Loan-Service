package controllers

import (
	"net/http"
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/pkg/config"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authUseCase domain.AuthUseCase
}

func NewAuthController(uc domain.AuthUseCase) *AuthController {
	return &AuthController{authUseCase: uc}
}




func (ctrl *AuthController) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(config.GetStatusCode(config.ErrBadRequest), gin.H{"error": err.Error()})
		return
	}
	access, refresh, err := ctrl.authUseCase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(config.GetStatusCode(err), gin.H{"error": err.Error()})
		return
	}
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