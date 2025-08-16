package routes

import (
	"SalaryAdvance/api/controllers"
	"SalaryAdvance/api/middleware"
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/internal/repositories"

	"SalaryAdvance/internal/usecases"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupCustomerRoutes(customerRoute *gin.RouterGroup, db *gorm.DB, jwtService domain.JWTService) {
	repo := repositories.NewCustomerRepository(db)
	uc := usecases.NewCustomerUseCase(repo)
	ctrl := controllers.NewCustomerController(uc)

	authCustomerRoute := customerRoute.Group("/")
	authCustomerRoute.Use(middleware.NewAuthMiddleware(jwtService).RequireAuth())
	{
		authCustomerRoute.POST("/import", ctrl.ImportCustomers)
		authCustomerRoute.GET("/:id", ctrl.GetCustomer)
		authCustomerRoute.GET("/", ctrl.GetAllCustomers)
	}
}
