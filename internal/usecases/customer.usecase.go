package usecases

import (
	"SalaryAdvance/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CustomerUseCase struct {
	customerRepo domain.CustomerRepository
	validator    *validator.Validate
}

func NewCustomerUseCase(customerRepo domain.CustomerRepository) *CustomerUseCase {
	return &CustomerUseCase{
		customerRepo: customerRepo,
		validator:    validator.New(),
	}
}

func (uc *CustomerUseCase) ImportCustomers(ctx context.Context, file io.Reader) ([]*domain.Customer, []map[string]interface{}, error) {
	var imported []*domain.Customer
	var logs []map[string]interface{}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %v", err)
	}

	log.Printf("Raw JSON input: %s", string(data))

	var input []struct {
		CustomerName string      `json:"customerName"`
		AccountNo    interface{} `json:"accountNo"`
	}
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, nil, fmt.Errorf("invalid JSON format: %v", err)
	}

	for i, in := range input {
		logEntry := map[string]interface{}{
			"record_index": i + 1,
			"verified":     false,
			"errors":       []string{},
		}

		if in.CustomerName == "" {
			logEntry["errors"] = append(logEntry["errors"].([]string), "customer name is required")
		}

		accountNoStr := ""
		switch v := in.AccountNo.(type) {
		case float64:
			accountNoStr = fmt.Sprintf("%.0f", v)
		case string:
			accountNoStr = v
		default:
			logEntry["errors"] = append(logEntry["errors"].([]string), "account number is in invalid format/type")
			logEntry["attempted_name"] = in.CustomerName
			logEntry["attempted_account_no"] = fmt.Sprintf("%v", in.AccountNo)
			logs = append(logs, logEntry)
			continue
		}

		if accountNoStr == "" {
			logEntry["errors"] = append(logEntry["errors"].([]string), "account number is required")
		}

		trimmedName := strings.TrimSpace(in.CustomerName)
		strippedAccount := strings.TrimLeft(accountNoStr, "0")

		if strippedAccount == "" || !isNumeric(strippedAccount) {
			logEntry["errors"] = append(logEntry["errors"].([]string), "account number is in invalid format/type")
		}

		verified := true
		var normalized *domain.Customer

		if len(logEntry["errors"].([]string)) == 0 {

			existing, err := uc.customerRepo.FindByNameAndAccountNo(ctx, trimmedName, accountNoStr)
			if err != nil {
				logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("database error: %v", err))
				verified = false
			} else if existing == nil {
				logEntry["errors"] = append(logEntry["errors"].([]string), "name or account number does not match existing records in customers table")
				verified = false
			} else {

				duplicate, err := uc.customerRepo.CheckDuplicateInValidCustomers(ctx, trimmedName, accountNoStr)
				if err != nil {
					logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("error checking duplicates in valid_customers: %v", err))
					verified = false
				} else if duplicate != nil {
					logEntry["errors"] = append(logEntry["errors"].([]string), "record already exists in valid_customers")
					verified = false
				} else {

					normalized = &domain.Customer{
						CustomerId:      fmt.Sprintf("CUST-%s", uuid.New().String()[:8]),
						CustomerName:    trimmedName,
						AccountNo:       domain.AccountNo(accountNoStr),
						Mobile:          "",
						BranchName:      "",
						BranchCode:      "",
						ProductName:     "",
						CustomerBalance: 0,
						CreatedAt:       time.Now(),
						UpdatedAt:       time.Now(),
					}

					if err := uc.validator.Struct(normalized); err != nil {
						logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("validation failed: %v", err))
						verified = false
					} else {
						_, err := uc.customerRepo.Create(ctx, normalized)
						if err != nil {
							logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("failed to save to valid_customers: %v", err))
							verified = false
						} else {
							imported = append(imported, normalized)
						}
					}
				}
			}
		} else {
			verified = false
		}

		logEntry["verified"] = verified
		if verified {
			logEntry["normalized_record"] = normalized
		} else {
			logEntry["attempted_name"] = in.CustomerName
			logEntry["attempted_account_no"] = accountNoStr
		}

		logs = append(logs, logEntry)
	}

	if len(imported) == 0 {
		return nil, logs, errors.New("no valid customers imported; see logs for details")
	}

	for _, l := range logs {
		if !l["verified"].(bool) {
			fmt.Printf("Unverified record %d: attempted_name=%s, attempted_account_no=%s, errors=%v\n",
				l["record_index"], l["attempted_name"], l["attempted_account_no"], l["errors"])
		}
	}

	return imported, logs, nil
}

func (uc *CustomerUseCase) ImportTransactions(ctx context.Context, file io.Reader, allowOverdraft bool) ([]*domain.Transaction, []map[string]interface{}, error) {
	var transactions []*domain.Transaction
	var logs []map[string]interface{}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %v", err)
	}

	var input []struct {
		FromAccount string  `json:"fromAccount"`
		ToAccount   string  `json:"toAccount"`
		Amount      float64 `json:"amount"`
		Date        string  `json:"date"`
	}
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, nil, fmt.Errorf("invalid JSON format: %v", err)
	}

	
	for i, in := range input {
		logEntry := map[string]interface{}{
			"record_index": i + 1,
			"verified":     false,
			"errors":       []string{},
		}

		
		if in.FromAccount == "" || in.ToAccount == "" {
			logEntry["errors"] = append(logEntry["errors"].([]string), "fromAccount and toAccount are required")
		}
		if in.Amount <= 0 {
			logEntry["errors"] = append(logEntry["errors"].([]string), "amount must be positive")
		}

		
		parsedDate, err := time.Parse("2006-01-02", in.Date)
		if err != nil {
			logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("invalid date format: %v", err))
		}

		if len(logEntry["errors"].([]string)) > 0 {
			logEntry["attempted_from_account"] = in.FromAccount
			logEntry["attempted_to_account"] = in.ToAccount
			logEntry["attempted_amount"] = in.Amount
			logs = append(logs, logEntry)
			continue
		}

		
		fromCustomer, err := uc.customerRepo.CheckDuplicateInValidCustomers(ctx, "", in.FromAccount)
		if err != nil {
			logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("error checking fromAccount: %v", err))
		} else if fromCustomer == nil {
			logEntry["errors"] = append(logEntry["errors"].([]string), "fromAccount not found in valid_customers")
		}

		toCustomer, err := uc.customerRepo.CheckDuplicateInValidCustomers(ctx, "", in.ToAccount)
		if err != nil {
			logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("error checking toAccount: %v", err))
		} else if toCustomer == nil {
			logEntry["errors"] = append(logEntry["errors"].([]string), "toAccount not found in valid_customers")
		}

		if len(logEntry["errors"].([]string)) > 0 {
			logEntry["attempted_from_account"] = in.FromAccount
			logEntry["attempted_to_account"] = in.ToAccount
			logEntry["attempted_amount"] = in.Amount
			logs = append(logs, logEntry)
			continue
		}

		
		if !allowOverdraft && fromCustomer.CustomerBalance < in.Amount {
			logEntry["errors"] = append(logEntry["errors"].([]string), "insufficient balance for fromAccount")
			logEntry["attempted_from_account"] = in.FromAccount
			logEntry["attempted_to_account"] = in.ToAccount
			logEntry["attempted_amount"] = in.Amount
			logs = append(logs, logEntry)
			continue
		}

		
		transaction := &domain.Transaction{
			TransactionID: fmt.Sprintf("TXN-%s", uuid.New().String()[:8]),
			FromAccount:   domain.AccountNo(in.FromAccount),
			ToAccount:     domain.AccountNo(in.ToAccount),
			Amount:        in.Amount,
			Date:          parsedDate,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		
		if err := uc.validator.Struct(transaction); err != nil {
			logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("validation failed: %v", err))
			logs = append(logs, logEntry)
			continue
		}

	
		_, err = uc.customerRepo.CreateTransaction(ctx, transaction)
		if err != nil {
			logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("failed to save transaction: %v", err))
			logs = append(logs, logEntry)
			continue
		}

		
		fromCustomer.CustomerBalance -= in.Amount
		toCustomer.CustomerBalance += in.Amount
		if _, err := uc.customerRepo.Update(ctx, fromCustomer); err != nil {
			logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("failed to update fromCustomer balance: %v", err))
			logs = append(logs, logEntry)
			continue
		}
		if _, err := uc.customerRepo.Update(ctx, toCustomer); err != nil {
			logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("failed to update toCustomer balance: %v", err))
			logs = append(logs, logEntry)
			continue
		}

		logEntry["verified"] = true
		logEntry["transaction"] = transaction
		logs = append(logs, logEntry)
		transactions = append(transactions, transaction)
	}

	
	customers, err := uc.customerRepo.FindAll(ctx)
	if err != nil {
		return transactions, logs, fmt.Errorf("failed to fetch customers for synthetic transactions: %v", err)
	}

	for _, customer := range customers {
		hasTransactions, err := uc.customerRepo.HasTransactions(ctx, string(customer.AccountNo))
		if err != nil {
			logs = append(logs, map[string]interface{}{
				"record_index": 0,
				"verified":     false,
				"errors":       []string{fmt.Sprintf("error checking transactions for customer %s: %v", customer.AccountNo, err)},
			})
			continue
		}

		if !hasTransactions {
			
			syntheticTransaction := &domain.Transaction{
				TransactionID: fmt.Sprintf("TXN-%s", uuid.New().String()[:8]),
				FromAccount:   customer.AccountNo,
				ToAccount:     domain.AccountNo("SYNTHETIC-" + string(customer.AccountNo)),
				Amount:        100.0, 
				Date:          time.Now(),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}

			
			if err := uc.validator.Struct(syntheticTransaction); err != nil {
				logs = append(logs, map[string]interface{}{
					"record_index": 0,
					"verified":     false,
					"errors":       []string{fmt.Sprintf("validation failed for synthetic transaction: %v", err)},
				})
				continue
			}

			
			_, err = uc.customerRepo.CreateTransaction(ctx, syntheticTransaction)
			if err != nil {
				logs = append(logs, map[string]interface{}{
					"record_index": 0,
					"verified":     false,
					"errors":       []string{fmt.Sprintf("failed to save synthetic transaction: %v", err)},
				})
				continue
			}

			
			if !allowOverdraft && customer.CustomerBalance < syntheticTransaction.Amount {
				logs = append(logs, map[string]interface{}{
					"record_index": 0,
					"verified":     false,
					"errors":       []string{"insufficient balance for synthetic transaction"},
				})
				continue
			}

			customer.CustomerBalance -= syntheticTransaction.Amount
			if _, err := uc.customerRepo.Update(ctx, customer); err != nil {
				logs = append(logs, map[string]interface{}{
					"record_index": 0,
					"verified":     false,
					"errors":       []string{fmt.Sprintf("failed to update customer balance for synthetic transaction: %v", err)},
				})
				continue
			}

			transactions = append(transactions, syntheticTransaction)
			logs = append(logs, map[string]interface{}{
				"record_index": 0,
				"verified":     true,
				"transaction":  syntheticTransaction,
				"synthetic":    true,
			})
		}
	}

	if len(transactions) == 0 {
		return nil, logs, errors.New("no valid transactions imported; see logs for details")
	}

	return transactions, logs, nil
}

func (uc *CustomerUseCase) CalculateCustomerRating(ctx context.Context, id string) (float64, error) {
	customer, err := uc.customerRepo.FindByID(ctx, id)
	if err != nil {
		return 0, fmt.Errorf("failed to find customer: %v", err)
	}

	transactions, err := uc.customerRepo.GetTransactionsByAccount(ctx, string(customer.AccountNo))
	if err != nil {
		return 0, fmt.Errorf("failed to fetch transactions: %v", err)
	}

	if len(transactions) == 0 {
		return 1.0, nil 
	}

	countScore := math.Min(float64(len(transactions))/10.0, 1.0) 


	var totalVolume float64
	for _, tx := range transactions {
		if tx.FromAccount == customer.AccountNo {
			totalVolume += tx.Amount
		}
	}
	volumeScore := math.Min(totalVolume/10000.0, 1.0) 

	
	var firstDate, lastDate time.Time
	for i, tx := range transactions {
		if i == 0 {
			firstDate = tx.Date
			lastDate = tx.Date
		} else {
			if tx.Date.Before(firstDate) {
				firstDate = tx.Date
			}
			if tx.Date.After(lastDate) {
				lastDate = tx.Date
			}
		}
	}
	durationDays := lastDate.Sub(firstDate).Hours() / 24
	durationScore := math.Min(durationDays/365.0, 1.0) 

	
	var balances []float64
	currentBalance := customer.CustomerBalance
	for i := len(transactions) - 1; i >= 0; i-- {
		tx := transactions[i]
		if tx.FromAccount == customer.AccountNo {
			currentBalance += tx.Amount
		} else if tx.ToAccount == customer.AccountNo {
			currentBalance -= tx.Amount
		}
		balances = append(balances, currentBalance)
	}

	
	var mean, sumSquaredDiff float64
	for _, balance := range balances {
		mean += balance
	}
	mean /= float64(len(balances))
	for _, balance := range balances {
		sumSquaredDiff += math.Pow(balance-mean, 2)
	}
	stdDev := math.Sqrt(sumSquaredDiff / float64(len(balances)))
	stabilityScore := math.Max(1.0-stdDev/10000.0, 0.0) 

	
	totalRating := (0.3*countScore + 0.3*volumeScore + 0.2*durationScore + 0.2*stabilityScore) * 10.0
	totalRating = math.Round(totalRating*10) / 10 
	if totalRating < 1.0 {
		totalRating = 1.0
	} else if totalRating > 10.0 {
		totalRating = 10.0
	}

	return totalRating, nil
}

func (uc *CustomerUseCase) GetCustomer(ctx context.Context, id string) (*domain.Customer, error) {
	return uc.customerRepo.FindByID(ctx, id)
}

func (uc *CustomerUseCase) GetAllCustomers(ctx context.Context) ([]*domain.Customer, error) {
	return uc.customerRepo.FindAll(ctx)
}

func isNumeric(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}
