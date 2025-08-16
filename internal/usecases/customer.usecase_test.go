package usecases

import (
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/internal/mocks"
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCustomerUseCase_ImportCustomers(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewCustomerRepository(t)
	uc := NewCustomerUseCase(mockRepo)

	tests := []struct {
		name              string
		inputJSON         string
		mockSetup         func()
		expectedCustomers []*domain.Customer
		expectedLogs      []map[string]interface{}
		expectedErr       error
	}{
		{
			name: "Valid customer import",
			inputJSON: `[
				{"customerName": "John Doe", "accountNo": "12345"},
				{"customerName": "Jane Smith", "accountNo": 67890}
			]`,
			mockSetup: func() {
				mockRepo.On("FindByNameAndAccountNo", ctx, "John Doe", "12345").
					Return(&domain.Customer{ID: 1, CustomerName: "John Doe", AccountNo: "12345"}, nil).Once()
				mockRepo.On("CheckDuplicateInValidCustomers", ctx, "John Doe", "12345").
					Return(nil, nil).Once()
				mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Customer")).
					Return(&domain.Customer{CustomerId: "CUST-12345678", CustomerName: "John Doe", AccountNo: "12345"}, nil).Once()
				mockRepo.On("FindByNameAndAccountNo", ctx, "Jane Smith", "67890").
					Return(&domain.Customer{ID: 2, CustomerName: "Jane Smith", AccountNo: "67890"}, nil).Once()
				mockRepo.On("CheckDuplicateInValidCustomers", ctx, "Jane Smith", "67890").
					Return(nil, nil).Once()
				mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Customer")).
					Return(&domain.Customer{CustomerId: "CUST-87654321", CustomerName: "Jane Smith", AccountNo: "67890"}, nil).Once()
			},
			expectedCustomers: []*domain.Customer{
				{CustomerId: "CUST-12345678", CustomerName: "John Doe", AccountNo: "12345"},
				{CustomerId: "CUST-87654321", CustomerName: "Jane Smith", AccountNo: "67890"},
			},
			expectedLogs: []map[string]interface{}{
				{
					"record_index":      1,
					"verified":          true,
					"normalized_record": mock.Anything,
				},
				{
					"record_index":      2,
					"verified":          true,
					"normalized_record": mock.Anything,
				},
			},
			expectedErr: nil,
		},
		{
			name:              "Invalid JSON",
			inputJSON:         `[{invalid json}]`,
			mockSetup:         func() {},
			expectedCustomers: nil,
			expectedLogs:      nil,
			expectedErr:       errors.New("invalid JSON format: invalid character 'i' looking for beginning of object key string"),
		},
		{
			name: "Missing customer name",
			inputJSON: `[
				{"customerName": "", "accountNo": "12345"}
			]`,
			mockSetup:         func() {},
			expectedCustomers: nil,
			expectedLogs: []map[string]interface{}{
				{
					"record_index":         1,
					"verified":             false,
					"errors":               []string{"customer name is required"},
					"attempted_name":       "",
					"attempted_account_no": "12345",
				},
			},
			expectedErr: errors.New("no valid customers imported; see logs for details"),
		},
		{
			name: "Duplicate customer",
			inputJSON: `[
				{"customerName": "John Doe", "accountNo": "12345"}
			]`,
			mockSetup: func() {
				mockRepo.On("FindByNameAndAccountNo", ctx, "John Doe", "12345").
					Return(&domain.Customer{ID: 1, CustomerName: "John Doe", AccountNo: "12345"}, nil).Once()
				mockRepo.On("CheckDuplicateInValidCustomers", ctx, "John Doe", "12345").
					Return(&domain.Customer{CustomerId: "CUST-12345678"}, nil).Once()
			},
			expectedCustomers: nil,
			expectedLogs: []map[string]interface{}{
				{
					"record_index":         1,
					"verified":             false,
					"errors":               []string{"record already exists in valid_customers"},
					"attempted_name":       "John Doe",
					"attempted_account_no": "12345",
				},
			},
			expectedErr: errors.New("no valid customers imported; see logs for details"),
		},
		{
			name: "Invalid account number",
			inputJSON: `[
				{"customerName": "John Doe", "accountNo": "abc"}
			]`,
			mockSetup:         func() {},
			expectedCustomers: nil,
			expectedLogs: []map[string]interface{}{
				{
					"record_index":         1,
					"verified":             false,
					"errors":               []string{"account number is in invalid format/type"},
					"attempted_name":       "John Doe",
					"attempted_account_no": "abc",
				},
			},
			expectedErr: errors.New("no valid customers imported; see logs for details"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			reader := bytes.NewReader([]byte(tt.inputJSON))
			customers, logs, err := uc.ImportCustomers(ctx, reader)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if len(tt.expectedCustomers) > 0 {
				assert.Len(t, customers, len(tt.expectedCustomers))
				for i, expected := range tt.expectedCustomers {
					assert.Equal(t, expected.CustomerName, customers[i].CustomerName)
					assert.Equal(t, expected.AccountNo, customers[i].AccountNo)
					assert.NotEmpty(t, customers[i].CustomerId)
				}
			} else {
				assert.Nil(t, customers)
			}

			if len(tt.expectedLogs) > 0 {
				assert.Len(t, logs, len(tt.expectedLogs))
				for i, expectedLog := range tt.expectedLogs {
					assert.Equal(t, expectedLog["record_index"], logs[i]["record_index"])
					assert.Equal(t, expectedLog["verified"], logs[i]["verified"])
					if !expectedLog["verified"].(bool) {
						assert.NotEmpty(t, logs[i]["errors"])
						if expectedLog["attempted_name"] != nil {
							assert.Equal(t, expectedLog["attempted_name"], logs[i]["attempted_name"])
						}
						if expectedLog["attempted_account_no"] != nil {
							assert.Equal(t, expectedLog["attempted_account_no"], logs[i]["attempted_account_no"])
						}
					} else {
						assert.NotNil(t, logs[i]["normalized_record"])
					}
				}
			} else {
				assert.Nil(t, logs)
			}
		})
	}
}

func TestCustomerUseCase_ImportTransactions(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewCustomerRepository(t)
	uc := NewCustomerUseCase(mockRepo)

	tests := []struct {
		name                 string
		inputJSON            string
		allowOverdraft       bool
		mockSetup            func()
		expectedTransactions []*domain.Transaction
		expectedLogs         []map[string]interface{}
		expectedErr          error
	}{
		{
			name: "Valid transaction import",
			inputJSON: `[
				{"fromAccount": "12345", "toAccount": "67890", "amount": 100.0, "date": "2025-01-01"}
			]`,
			allowOverdraft: true,
			mockSetup: func() {
				mockRepo.On("CheckDuplicateInValidCustomers", ctx, "", "12345").
					Return(&domain.Customer{ID: 1, CustomerId: "CUST-12345678", AccountNo: "12345", CustomerBalance: 1000.0}, nil).Once()
				mockRepo.On("CheckDuplicateInValidCustomers", ctx, "", "67890").
					Return(&domain.Customer{ID: 2, CustomerId: "CUST-87654321", AccountNo: "67890", CustomerBalance: 500.0}, nil).Once()
				mockRepo.On("CreateTransaction", ctx, mock.AnythingOfType("*domain.Transaction")).
					Return(&domain.Transaction{TransactionID: "TXN-12345678"}, nil).Once()
				mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.Customer")).
					Return(&domain.Customer{}, nil).Times(2)
				mockRepo.On("FindAll", ctx).
					Return([]*domain.Customer{
						{ID: 1, CustomerId: "CUST-12345678", AccountNo: "12345", CustomerBalance: 900.0},
					}, nil).Once()
				mockRepo.On("HasTransactions", ctx, "12345").
					Return(true, nil).Once()
			},
			expectedTransactions: []*domain.Transaction{
				{TransactionID: "TXN-12345678", FromAccount: "12345", ToAccount: "67890", Amount: 100.0},
			},
			expectedLogs: []map[string]interface{}{
				{
					"record_index": 1,
					"verified":     true,
					"transaction":  mock.Anything,
				},
			},
			expectedErr: nil,
		},
		{
			name:                 "Invalid JSON",
			inputJSON:            `[{invalid json}]`,
			allowOverdraft:       false,
			mockSetup:            func() {},
			expectedTransactions: nil,
			expectedLogs:         nil,
			expectedErr:          errors.New("invalid JSON format: invalid character 'i' looking for beginning of object key string"),
		},
		{
			name: "Insufficient balance",
			inputJSON: `[
				{"fromAccount": "12345", "toAccount": "67890", "amount": 2000.0, "date": "2025-01-01"}
			]`,
			allowOverdraft: false,
			mockSetup: func() {
				mockRepo.On("CheckDuplicateInValidCustomers", ctx, "", "12345").
					Return(&domain.Customer{ID: 1, CustomerId: "CUST-12345678", AccountNo: "12345", CustomerBalance: 1000.0}, nil).Once()
				mockRepo.On("CheckDuplicateInValidCustomers", ctx, "", "67890").
					Return(&domain.Customer{ID: 2, CustomerId: "CUST-87654321", AccountNo: "67890"}, nil).Once()
				mockRepo.On("FindAll", ctx).
					Return([]*domain.Customer{
						{ID: 1, CustomerId: "CUST-12345678", AccountNo: "12345", CustomerBalance: 1000.0},
					}, nil).Once()
				mockRepo.On("HasTransactions", ctx, "12345").
					Return(true, nil).Once()
			},
			expectedTransactions: nil,
			expectedLogs: []map[string]interface{}{
				{
					"record_index":           1,
					"verified":               false,
					"errors":                 []string{"insufficient balance for fromAccount"},
					"attempted_from_account": "12345",
					"attempted_to_account":   "67890",
					"attempted_amount":       2000.0,
				},
			},
			expectedErr: errors.New("no valid transactions imported; see logs for details"),
		},
		{
			name:           "Synthetic transaction",
			inputJSON:      `[]`,
			allowOverdraft: true,
			mockSetup: func() {
				mockRepo.On("FindAll", ctx).
					Return([]*domain.Customer{
						{ID: 1, CustomerId: "CUST-12345678", AccountNo: "12345", CustomerBalance: 1000.0},
					}, nil).Once()
				mockRepo.On("HasTransactions", ctx, "12345").
					Return(false, nil).Once()
				mockRepo.On("CreateTransaction", ctx, mock.AnythingOfType("*domain.Transaction")).
					Return(&domain.Transaction{TransactionID: "TXN-12345678"}, nil).Once()
				mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.Customer")).
					Return(&domain.Customer{}, nil).Once()
			},
			expectedTransactions: []*domain.Transaction{
				{TransactionID: "TXN-12345678", FromAccount: "12345", ToAccount: "SYNTHETIC-12345", Amount: 100.0},
			},
			expectedLogs: []map[string]interface{}{
				{
					"record_index": 0,
					"verified":     true,
					"transaction":  mock.Anything,
					"synthetic":    true,
				},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			reader := bytes.NewReader([]byte(tt.inputJSON))
			transactions, logs, err := uc.ImportTransactions(ctx, reader, tt.allowOverdraft)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if len(tt.expectedTransactions) > 0 {
				assert.Len(t, transactions, len(tt.expectedTransactions))
				for i, expected := range tt.expectedTransactions {
					assert.Equal(t, expected.FromAccount, transactions[i].FromAccount)
					assert.Equal(t, expected.ToAccount, transactions[i].ToAccount)
					assert.Equal(t, expected.Amount, transactions[i].Amount)
					assert.NotEmpty(t, transactions[i].TransactionID)
				}
			} else {
				assert.Nil(t, transactions)
			}

			if len(tt.expectedLogs) > 0 {
				assert.Len(t, logs, len(tt.expectedLogs))
				for i, expectedLog := range tt.expectedLogs {
					assert.Equal(t, expectedLog["record_index"], logs[i]["record_index"])
					assert.Equal(t, expectedLog["verified"], logs[i]["verified"])
					if !expectedLog["verified"].(bool) {
						assert.NotEmpty(t, logs[i]["errors"])
						if expectedLog["attempted_from_account"] != nil {
							assert.Equal(t, expectedLog["attempted_from_account"], logs[i]["attempted_from_account"])
						}
						if expectedLog["attempted_to_account"] != nil {
							assert.Equal(t, expectedLog["attempted_to_account"], logs[i]["attempted_to_account"])
						}
						if expectedLog["attempted_amount"] != nil {
							assert.Equal(t, expectedLog["attempted_amount"], logs[i]["attempted_amount"])
						}
					} else {
						assert.NotNil(t, logs[i]["transaction"])
						if expectedLog["synthetic"] != nil {
							assert.Equal(t, expectedLog["synthetic"], logs[i]["synthetic"])
						}
					}
				}
			} else {
				assert.Nil(t, logs)
			}
		})
	}
}

func TestCustomerUseCase_CalculateCustomerRating(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewCustomerRepository(t)
	uc := NewCustomerUseCase(mockRepo)

	tests := []struct {
		name           string
		customerID     string
		mockSetup      func()
		expectedRating float64
		expectedErr    error
	}{
		{
			name:       "Customer with transactions",
			customerID: "1",
			mockSetup: func() {
				customer := &domain.Customer{
					ID:              1,
					CustomerId:      "CUST-12345678",
					AccountNo:       "12345",
					CustomerBalance: 1000.0,
				}
				transactions := []*domain.Transaction{
					{
						TransactionID: "TXN-1",
						FromAccount:   "12345",
						ToAccount:     "67890",
						Amount:        500.0,
						Date:          time.Now().AddDate(0, 0, -365),
					},
					{
						TransactionID: "TXN-2",
						FromAccount:   "12345",
						ToAccount:     "67890",
						Amount:        500.0,
						Date:          time.Now(),
					},
				}
				mockRepo.On("FindByID", ctx, "1").Return(customer, nil).Once()
				mockRepo.On("GetTransactionsByAccount", ctx, "12345").Return(transactions, nil).Once()
			},
			expectedRating: 4.9, 
			expectedErr:    nil,
		},
		{
			name:       "Customer with no transactions",
			customerID: "1",
			mockSetup: func() {
				customer := &domain.Customer{ID: 1, CustomerId: "CUST-12345678", AccountNo: "12345"}
				mockRepo.On("FindByID", ctx, "1").Return(customer, nil).Once()
				mockRepo.On("GetTransactionsByAccount", ctx, "12345").Return([]*domain.Transaction{}, nil).Once()
			},
			expectedRating: 1.0,
			expectedErr:    nil,
		},
		{
			name:       "Customer not found",
			customerID: "1",
			mockSetup: func() {
				mockRepo.On("FindByID", ctx, "1").Return(nil, errors.New("not found")).Once()
			},
			expectedRating: 0,
			expectedErr:    errors.New("failed to find customer: not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			rating, err := uc.CalculateCustomerRating(ctx, tt.customerID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Equal(t, 0.0, rating)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expectedRating, rating, 0.1)
			}
		})
	}
}

func TestCustomerUseCase_GetCustomer(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewCustomerRepository(t)
	uc := NewCustomerUseCase(mockRepo)

	tests := []struct {
		name             string
		customerID       string
		mockSetup        func()
		expectedCustomer *domain.Customer
		expectedErr      error
	}{
		{
			name:       "Valid customer",
			customerID: "1",
			mockSetup: func() {
				mockRepo.On("FindByID", ctx, "1").
					Return(&domain.Customer{ID: 1, CustomerId: "CUST-12345678", CustomerName: "John Doe", AccountNo: "12345"}, nil).Once()
			},
			expectedCustomer: &domain.Customer{ID: 1, CustomerId: "CUST-12345678", CustomerName: "John Doe", AccountNo: "12345"},
			expectedErr:      nil,
		},
		{
			name:       "Customer not found",
			customerID: "1",
			mockSetup: func() {
				mockRepo.On("FindByID", ctx, "1").Return(nil, errors.New("not found")).Once()
			},
			expectedCustomer: nil,
			expectedErr:      errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			customer, err := uc.GetCustomer(ctx, tt.customerID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, customer)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCustomer, customer)
			}
		})
	}
}

func TestCustomerUseCase_GetAllCustomers(t *testing.T) {
	ctx := context.Background()
	mockRepo := mocks.NewCustomerRepository(t)
	uc := NewCustomerUseCase(mockRepo)

	tests := []struct {
		name              string
		mockSetup         func()
		expectedCustomers []*domain.Customer
		expectedErr       error
	}{
		{
			name: "Multiple customers",
			mockSetup: func() {
				mockRepo.On("FindAll", ctx).
					Return([]*domain.Customer{
						{ID: 1, CustomerId: "CUST-12345678", CustomerName: "John Doe", AccountNo: "12345"},
						{ID: 2, CustomerId: "CUST-87654321", CustomerName: "Jane Smith", AccountNo: "67890"},
					}, nil).Once()
			},
			expectedCustomers: []*domain.Customer{
				{ID: 1, CustomerId: "CUST-12345678", CustomerName: "John Doe", AccountNo: "12345"},
				{ID: 2, CustomerId: "CUST-87654321", CustomerName: "Jane Smith", AccountNo: "67890"},
			},
			expectedErr: nil,
		},
		{
			name: "No customers",
			mockSetup: func() {
				mockRepo.On("FindAll", ctx).Return([]*domain.Customer{}, nil).Once()
			},
			expectedCustomers: []*domain.Customer{},
			expectedErr:       nil,
		},
		{
			name: "Repository error",
			mockSetup: func() {
				mockRepo.On("FindAll", ctx).Return(nil, errors.New("database error")).Once()
			},
			expectedCustomers: nil,
			expectedErr:       errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			customers, err := uc.GetAllCustomers(ctx)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, customers)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCustomers, customers)
			}
		})
	}
}
