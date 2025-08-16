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

func SetupRatingRoutes(ratingRoute *gin.RouterGroup, db *gorm.DB, jwtService domain.JWTService) {
	ratingRepo := repositories.NewRatingRepository(db)
	transactionRepo := repositories.NewTransactionRepository(db)
	customerRepo := repositories.NewCustomerRepository(db)
	transactionUseCase := usecases.NewTransactionUseCase(transactionRepo, customerRepo)
	ratingUseCase := usecases.NewRatingUseCase(ratingRepo, transactionRepo, customerRepo, transactionUseCase)
	ctrl := controllers.NewRatingController(ratingUseCase)

	
	authRatingRoute := ratingRoute.Group("/").Use(middleware.NewAuthMiddleware(jwtService).RequireAuth())
	{
		authRatingRoute.GET("", ctrl.GetAllRatings)
		authRatingRoute.GET("/customer/:customerID", ctrl.GetRating)
	}
}
