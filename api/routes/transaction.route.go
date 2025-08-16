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

func SetupTransactionRoutes(txRoute *gin.RouterGroup, db *gorm.DB, jwtService domain.JWTService) {
	transactionRepo := repositories.NewTransactionRepository(db)
	customerRepo := repositories.NewCustomerRepository(db)
	ratingRepo := repositories.NewRatingRepository(db)
	transactionUseCase := usecases.NewTransactionUseCase(transactionRepo, customerRepo)
	ratingUseCase := usecases.NewRatingUseCase(ratingRepo, transactionRepo, customerRepo, transactionUseCase)
	ctrl := controllers.NewTransactionController(transactionUseCase, ratingUseCase)


	authTxRoute := txRoute.Group("/").Use(middleware.NewAuthMiddleware(jwtService).RequireAuth())
	{
		authTxRoute.POST("", ctrl.AddTransaction)
		authTxRoute.GET("/customer/:customerID", ctrl.GetTransactionsForCustomer)
		authTxRoute.GET("", ctrl.GetAllTransactions)
		authTxRoute.POST("/import", ctrl.ImportTransactions)
	}
}
