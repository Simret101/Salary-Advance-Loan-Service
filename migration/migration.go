package migration

import (
	"SalaryAdvance/internal/domain" 

	"gorm.io/gorm"
)

func MigrateModels(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.User{},
		&domain.Customer{},
		&domain.Transaction{},
		&domain.Rating{},
	)
}
