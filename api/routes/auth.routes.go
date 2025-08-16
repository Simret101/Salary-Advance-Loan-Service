package routes

import (
	"SalaryAdvance/api/controllers"
	"SalaryAdvance/api/middleware"
	"SalaryAdvance/internal/repositories"
	"SalaryAdvance/internal/services"
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/internal/usecases"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupAuthRoutes(userRoute *gin.RouterGroup, db *gorm.DB, jwtService domain.JWTService) {
	userRepo := repositories.NewUserRepository(db)
	inviteRepo := repositories.NewInviteRepository(db)
	emailService := services.NewEmailService()
	authUsecase := usecases.NewAuthUseCase(userRepo, jwtService)
	inviteUsecase := usecases.NewInviteUseCase(inviteRepo, userRepo, emailService, jwtService)
	authCtrl := controllers.NewAuthController(authUsecase)
	inviteCtrl := controllers.NewInviteController(inviteUsecase)

	auth := userRoute.Group("/auth")
	{
		auth.POST("/login", authCtrl.Login)
		auth.POST("/register", authCtrl.Register)
		auth.POST("/refresh", authCtrl.Refresh)
	}

	invite := userRoute.Group("/invite")
	invite.Use(middleware.NewAuthMiddleware(jwtService).RequireAuth(), middleware.NewAuthMiddleware(jwtService).RequireAdmin())
	{
		invite.POST("/send", inviteCtrl.SendInvite)
	}
}
