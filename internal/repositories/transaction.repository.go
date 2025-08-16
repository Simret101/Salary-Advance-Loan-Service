package repositories

import (
	"context"
	"fmt"

	"SalaryAdvance/internal/domain"

	"gorm.io/gorm"
)

type TransactionRepositoryImpl struct {
	DB *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepositoryImpl {
	return &TransactionRepositoryImpl{DB: db}
}

func (r *TransactionRepositoryImpl) Create(ctx context.Context, transaction *domain.Transaction) (*domain.Transaction, error) {
	result := r.DB.WithContext(ctx).Create(transaction)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create transaction for customer %s: %v", transaction.CustomerID, result.Error)
	}
	return transaction, nil
}

func (r *TransactionRepositoryImpl) GetByCustomerID(ctx context.Context, customerID string) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := r.DB.WithContext(ctx).Where("customer_id = ?", customerID).Find(&transactions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions for customer %s: %v", customerID, err)
	}
	return transactions, nil
}

func (r *TransactionRepositoryImpl) GetAll(ctx context.Context) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := r.DB.WithContext(ctx).Find(&transactions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all transactions: %v", err)
	}
	return transactions, nil
}
