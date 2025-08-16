package routes

import (
	"SalaryAdvance/internal/services"
	"SalaryAdvance/pkg/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.RouterGroup, cfg *config.Config, db *gorm.DB) {

	jwtService := services.NewJWTService(cfg)

	SetupCustomerRoutes(r.Group("/customers"), db, jwtService)
	SetupAuthRoutes(r.Group("/user"), db, jwtService)
}
