package domain

import (
	"context"
	"io"
)

type CustomerRepository interface {
	Create(ctx context.Context, customer *Customer) (*Customer, error)
	FindByNameAndAccountNo(ctx context.Context, name string, accountNo string) (*Customer, error)
	FindByID(ctx context.Context, id string) (*Customer, error)
	FindAll(ctx context.Context) ([]*Customer, error)
	CheckDuplicateInValidCustomers(ctx context.Context, name string, accountNo string) (*Customer, error)

	GetAll(ctx context.Context) ([]*Customer, error)
	CreateTransaction(ctx context.Context, transaction *Transaction) (*Transaction, error)
	Update(ctx context.Context, customer *Customer) (*Customer, error)
	HasTransactions(ctx context.Context, accountNo string) (bool, error)
	GetTransactionsByAccount(ctx context.Context, accountNo string) ([]*Transaction, error)
	GetTransactionsByCustomerId(ctx context.Context, customerId string) ([]*Transaction, error)
}

type CustomerUseCase interface {
	ImportCustomers(ctx context.Context, customers []*Customer) ([]*Customer, []map[string]interface{}, error)
	GetCustomer(ctx context.Context, id string) (*Customer, error)
	GetAllCustomers(ctx context.Context) ([]*Customer, error)
	ImportTransactions(ctx context.Context, file io.Reader, allowOverdraft bool) ([]*Transaction, []map[string]interface{}, error)
	CalculateCustomerRating(ctx context.Context, id string) (float64, error)
}
