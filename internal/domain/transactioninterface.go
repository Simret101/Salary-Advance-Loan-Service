package domain

import (
	"context"
	"io"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) (*Transaction, error)
	GetByCustomerID(ctx context.Context, customerID string) ([]Transaction, error)
	GetAll(ctx context.Context) ([]Transaction, error)
}
type TransactionUseCase interface {
	AddTransaction(ctx context.Context, transaction *Transaction) (*Transaction, error)
	GetTransactionsForCustomer(ctx context.Context, customerID string) ([]Transaction, error)
	GetAll(ctx context.Context) ([]Transaction, error)
	ImportTransactions(ctx context.Context, file io.Reader) ([]*Transaction, []map[string]interface{}, error)
	GenerateSyntheticTransactions(ctx context.Context, customerID string) ([]Transaction, error)
}
