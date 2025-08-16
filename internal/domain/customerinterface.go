package domain

import (
	"context"
)

type CustomerRepository interface {
	Create(ctx context.Context, customer *Customer) (*Customer, error)
	FindByNameAndAccountNo(ctx context.Context, name string, accountNo string) (*Customer, error)
	FindByID(ctx context.Context, id string) (*Customer, error)
	FindAll(ctx context.Context) ([]*Customer, error)
	CheckDuplicateInValidCustomers(ctx context.Context, name string, accountNo string) (*Customer, error)

	GetAll(ctx context.Context) ([]*Customer, error)
}

type CustomerUseCase interface {
	ImportCustomers(ctx context.Context, customers []*Customer) ([]*Customer, []map[string]interface{}, error)
	GetCustomer(ctx context.Context, id string) (*Customer, error)
	GetAllCustomers(ctx context.Context) ([]*Customer, error)
}
