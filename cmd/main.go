package main

import (
	"SalaryAdvance/api/routes"
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/internal/services"
	"SalaryAdvance/migration"
	"SalaryAdvance/pkg/config"
	"SalaryAdvance/pkg/database"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	cfg := config.LoadConfig()

	db := database.ConnectDB()

	if err := migration.MigrateModels(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	} else {
		log.Println("Database migration completed successfully")
	}
	err = db.Table("customers").AutoMigrate(&domain.Customer{})
	if err != nil {
		panic("failed to auto-migrate customers: " + err.Error())
	}
	err = db.Table("valid_customers").AutoMigrate(&domain.Customer{})
	if err != nil {
		panic("failed to auto-migrate valid_customers: " + err.Error())
	}

	db.AutoMigrate(&domain.User{}, &domain.Invite{}, &domain.Customer{})

	var admin domain.User
	if err := db.Where("email = ?", cfg.AdminEmail).First(&admin).Error; err != nil {
		hashedPassword, _ := services.HashPassword(cfg.AdminPassword)
		admin = domain.User{
			Email:    cfg.AdminEmail,
			Password: hashedPassword,
			Role:     "admin",
		}
		db.Create(&admin)
		log.Println("Admin user seeded successfully")
	}

	router := gin.Default()

	router.Use(cors.Default())

	routes.SetupRoutes(router.Group("/api/v0"), &cfg, db)

	log.Printf("Server is running on port %s\n", cfg.AppPort)
	if err := router.Run(cfg.AppPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
