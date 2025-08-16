package repositories

import (
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/pkg/config"
	"context"
	"strings"
	
	"gorm.io/gorm"
)

type CustomerRepositoryImpl struct {
	DB *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepositoryImpl {
	return &CustomerRepositoryImpl{DB: db}
}

func (r *CustomerRepositoryImpl) Create(ctx context.Context, customer *domain.Customer) (*domain.Customer, error) {
	result := r.DB.WithContext(ctx).Table("valid_customers").Create(customer)
	if result.Error != nil {
		return nil, config.ErrInternalServer
	}
	return customer, nil
}

func (r *CustomerRepositoryImpl) FindByNameAndAccountNo(ctx context.Context, name string, accountNo string) (*domain.Customer, error) {
	var customer domain.Customer
	trimmedName := strings.ToLower(strings.TrimSpace(name))
	strippedAccount := strings.TrimLeft(accountNo, "0")
	if strippedAccount == "" {
		return nil, nil 
	}
	if err := r.DB.WithContext(ctx).
		Table("customers").
		Where("LOWER(TRIM(customer_name)) = ? AND regexp_replace(CAST(account_no AS TEXT), '^0+', '', 'g') = ?", trimmedName, strippedAccount).
		First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil 
		}
		return nil, config.ErrInternalServer
	}
	return &customer, nil
}

func (r *CustomerRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Customer, error) {
	var customer domain.Customer
	if err := r.DB.WithContext(ctx).Table("valid_customers").Where("id = ?", id).First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, config.ErrNotFound
		}
		return nil, config.ErrInternalServer
	}
	return &customer, nil
}

func (r *CustomerRepositoryImpl) FindAll(ctx context.Context) ([]*domain.Customer, error) {
	var customers []*domain.Customer
	if err := r.DB.WithContext(ctx).Table("valid_customers").Find(&customers).Error; err != nil {
		return nil, config.ErrInternalServer
	}
	return customers, nil
}
func (r *CustomerRepositoryImpl) GetAll(ctx context.Context) ([]*domain.Customer, error) {
	var customers []*domain.Customer
	err := r.DB.WithContext(ctx).Find(&customers).Error
	if err != nil {
		return nil, err
	}
	return customers, nil
}

func (r *CustomerRepositoryImpl) CheckDuplicateInValidCustomers(ctx context.Context, name string, accountNo string) (*domain.Customer, error) {
	var customer domain.Customer
	trimmedName := strings.ToLower(strings.TrimSpace(name))
	strippedAccount := strings.TrimLeft(accountNo, "0")
	if strippedAccount == "" {
		return nil, nil
	}
	if err := r.DB.WithContext(ctx).
		Table("valid_customers").
		Where("LOWER(TRIM(customer_name)) = ? AND regexp_replace(CAST(account_no AS TEXT), '^0+', '', 'g') = ?", trimmedName, strippedAccount).
		First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, config.ErrInternalServer
	}
	return &customer, nil
}

func (r *CustomerRepositoryImpl) CreateTransaction(ctx context.Context, transaction *domain.Transaction) (*domain.Transaction, error) {
	result := r.DB.WithContext(ctx).Table("transactions").Create(transaction)
	if result.Error != nil {
		return nil, config.ErrInternalServer
	}
	return transaction, nil
}

func (r *CustomerRepositoryImpl) Update(ctx context.Context, customer *domain.Customer) (*domain.Customer, error) {
	result := r.DB.WithContext(ctx).Table("valid_customers").Save(customer)
	if result.Error != nil {
		return nil, config.ErrInternalServer
	}
	return customer, nil
}

func (r *CustomerRepositoryImpl) HasTransactions(ctx context.Context, accountNo string) (bool, error) {
	var count int64
	strippedAccount := strings.TrimLeft(accountNo, "0")
	if strippedAccount == "" {
		return false, nil
	}
	err := r.DB.WithContext(ctx).Table("transactions").
		Where("from_account = ? OR to_account = ?", accountNo, accountNo).
		Count(&count).Error
	if err != nil {
		return false, config.ErrInternalServer
	}
	return count > 0, nil
}

func (r *CustomerRepositoryImpl) GetTransactionsByAccount(ctx context.Context, accountNo string) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction
	err := r.DB.WithContext(ctx).Table("transactions").
		Where("from_account = ? OR to_account = ?", accountNo, accountNo).
		Find(&transactions).Error
	if err != nil {
		return nil, config.ErrInternalServer
	}
	return transactions, nil
}
func (r *CustomerRepositoryImpl) GetTransactionsByCustomerId(ctx context.Context, customerId string) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction
	result := r.DB.WithContext(ctx).Table("transactions").Where("customer_id = ?", customerId).Find(&transactions)
	if result.Error != nil {
		return nil, config.ErrInternalServer
	}
	return transactions, nil
}