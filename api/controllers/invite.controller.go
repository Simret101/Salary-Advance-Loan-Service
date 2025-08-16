package controllers

import (
	"net/http"
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/pkg/config"

	"github.com/gin-gonic/gin"
)

type InviteController struct {
	inviteUseCase domain.InviteUseCase
}

func NewInviteController(uc domain.InviteUseCase) *InviteController {
	return &InviteController{inviteUseCase: uc}
}




func (ctrl *InviteController) SendInvite(c *gin.Context) {
	var req domain.SendInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(config.GetStatusCode(config.ErrBadRequest), gin.H{"error": err.Error()})
		return
	}
	adminID := c.GetUint("user_id")
	link, err := ctrl.inviteUseCase.SendInvite(c.Request.Context(), adminID, req.Email)
	if err != nil {
		c.JSON(config.GetStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"invite_link": link})
}