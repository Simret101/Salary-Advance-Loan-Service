package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strconv"
	"time"

	"SalaryAdvance/internal/domain"

	"github.com/google/uuid"
)

type TransactionUseCase struct {
	transactionRepo domain.TransactionRepository
	customerRepo    domain.CustomerRepository
}

func NewTransactionUseCase(transactionRepo domain.TransactionRepository, customerRepo domain.CustomerRepository) *TransactionUseCase {
	return &TransactionUseCase{
		transactionRepo: transactionRepo,
		customerRepo:    customerRepo,
	}
}

func (uc *TransactionUseCase) AddTransaction(ctx context.Context, txn *domain.Transaction) (*domain.Transaction, error) {
	if txn.CustomerID == "" {
		return nil, errors.New("customer ID is required")
	}
	if txn.Amount <= 0 {
		return nil, errors.New("invalid transaction amount")
	}
	if txn.TransactionType != "Debit" && txn.TransactionType != "Credit" {
		return nil, errors.New("transaction type must be Debit or Credit")
	}
	txn.Status = "pending"
	txn.CreatedAt = time.Now()
	createdTxn, err := uc.transactionRepo.Create(ctx, txn)
	if err != nil {
		return nil, err
	}
	return createdTxn, nil
}

func (uc *TransactionUseCase) GetTransactionsForCustomer(ctx context.Context, customerID string) ([]domain.Transaction, error) {
	if customerID == "" {
		return nil, errors.New("customer ID is required")
	}
	return uc.transactionRepo.GetByCustomerID(ctx, customerID)
}

func (uc *TransactionUseCase) GetAll(ctx context.Context) ([]domain.Transaction, error) {
	return uc.transactionRepo.GetAll(ctx)
}

func (uc *TransactionUseCase) ImportTransactions(ctx context.Context, file io.Reader) ([]*domain.Transaction, []map[string]interface{}, error) {
	var imported []*domain.Transaction
	var logs []map[string]interface{}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %v", err)
	}

	var input []struct {
		ID                  string  `json:"id"`
		FromAccount         string  `json:"fromAccount"`
		ToAccount           string  `json:"toAccount"`
		Amount              float64 `json:"amount"`
		Remark              string  `json:"remark"`
		TransactionType     string  `json:"transactionType"`
		RequestId           string  `json:"requestId"`
		Reference           string  `json:"reference"`
		ThirdPartyReference string  `json:"thirdPartyReference"`
		InstitutionId       string  `json:"institutionId"`
		ClearedBalance      float64 `json:"clearedBalance"`
		TransactionDate     string  `json:"transactionDate"`
		BillerId            string  `json:"billerId"`
	}
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, nil, fmt.Errorf("invalid JSON format: %v", err)
	}

	
	customerTransactionMap := make(map[string]bool)
	for _, tx := range input {
		customer, err := uc.customerRepo.FindByID(ctx, tx.FromAccount)
		if err == nil && customer != nil {
			customerTransactionMap[customer.CustomerId] = true
		}
	}

	
	customers, err := uc.customerRepo.GetAll(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch customers: %v", err)
	}
	for _, customer := range customers {
		if !customerTransactionMap[customer.CustomerId] {
			syntheticTxns, err := uc.GenerateSyntheticTransactions(ctx, customer.CustomerId)
			if err != nil {
				logs = append(logs, map[string]interface{}{
					"customer_id": customer.CustomerId,
					"verified":    false,
					"errors":      []string{fmt.Sprintf("failed to generate synthetic transactions: %v", err)},
				})
				continue
			}
			for _, tx := range syntheticTxns {
				imported = append(imported, &tx)
				logs = append(logs, map[string]interface{}{
					"record_index":      len(logs) + 1,
					"verified":          true,
					"synthetic":         true,
					"customer_id":       customer.CustomerId,
					"normalized_record": tx,
				})
			}
		}
	}

	
	for i, in := range input {
		logEntry := map[string]interface{}{
			"record_index": i + 1,
			"verified":     false,
			"errors":       []string{},
		}

		
		if in.FromAccount == "" {
			logEntry["errors"] = append(logEntry["errors"].([]string), "fromAccount is required")
		}
		if in.Amount <= 0 {
			logEntry["errors"] = append(logEntry["errors"].([]string), "amount must be positive")
		}
		transactionDateMs, err := strconv.ParseInt(in.TransactionDate, 10, 64)
		if err != nil {
			logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("invalid transaction date: %v", err))
		}
		transactionDate := time.UnixMilli(transactionDateMs)

		verified := true
		var normalized *domain.Transaction

		if len(logEntry["errors"].([]string)) == 0 {
			
			customer, err := uc.customerRepo.FindByID(ctx, in.FromAccount)
			if err != nil {
				logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("database error finding customer: %v", err))
				verified = false
			} else if customer == nil {
				logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("no customer found for account %s", in.FromAccount))
				verified = false
			} else {
			
				normalizedType := "Debit"
				if in.ToAccount != "" {
					toCustomer, err := uc.customerRepo.FindByID(ctx, in.ToAccount)
					if err == nil && toCustomer != nil && toCustomer.CustomerId == customer.CustomerId {
						normalizedType = "Credit"
					}
				}
				switch in.TransactionType {
				case "Derash Bill Payment", "DSTV Payment", "OtherBank Transaction", "mpesa Transaction", "M-PESA Transaction", "YIMULU", "SAFARI AIRTIME", "telebirr Transaction":
					normalizedType = "Debit"
				case "withInBank Transaction", "Bank2Bank Transaction":
					
				default:
					logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("unsupported transaction type: %s", in.TransactionType))
					verified = false
				}

				if verified {
					
					currentBalance := customer.CustomerBalance
					if normalizedType == "Debit" {
						if currentBalance < in.Amount {
							logEntry["errors"] = append(logEntry["errors"].([]string), "insufficient balance for debit")
							verified = false
						} else {
							currentBalance -= in.Amount
						}
					} else {
						currentBalance += in.Amount
					}

					if verified {
						normalized = &domain.Transaction{
							CustomerID:          customer.CustomerId,
							FromAccount:         in.FromAccount,
							ToAccount:           in.ToAccount,
							Amount:              in.Amount,
							Remark:              in.Remark,
							TransactionType:     normalizedType,
							RequestId:           in.RequestId,
							Reference:           in.Reference,
							ThirdPartyReference: in.ThirdPartyReference,
							InstitutionId:       in.InstitutionId,
							ClearedBalance:      currentBalance,
							TransactionDate:     transactionDate,
							BillerId:            in.BillerId,
							Status:              "completed",
							CreatedAt:           time.Now(),
							UpdatedAt:           time.Now(),
						}
						_, err := uc.transactionRepo.Create(ctx, normalized)
						if err != nil {
							logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("failed to save transaction: %v", err))
							verified = false
						} else {
							
							customer.CustomerBalance = currentBalance
							_, err = uc.customerRepo.Create(ctx, customer)
							if err != nil {
								logEntry["errors"] = append(logEntry["errors"].([]string), fmt.Sprintf("failed to update customer balance: %v", err))
								verified = false
							} else {
								imported = append(imported, normalized)
							}
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
			logEntry["attempted_from_account"] = in.FromAccount
			logEntry["attempted_amount"] = in.Amount
			logEntry["attempted_transaction_type"] = in.TransactionType
		}
		logs = append(logs, logEntry)
	}

	return imported, logs, nil
}

func (uc *TransactionUseCase) GenerateSyntheticTransactions(ctx context.Context, customerID string) ([]domain.Transaction, error) {
	if customerID == "" {
		return nil, errors.New("customer ID is required")
	}

	customer, err := uc.customerRepo.FindByID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find customer: %v", err)
	}
	if customer == nil {
		return nil, errors.New("customer not found")
	}

	currentBalance := customer.CustomerBalance
	numTxns := rand.Intn(4) + 2 
	txns := make([]domain.Transaction, 0, numTxns)

	for i := 0; i < numTxns; i++ {
		isCredit := rand.Float32() < 0.5
		amount := 50 + rand.Float64()*450 
		transactionType := "Debit"
		if isCredit {
			transactionType = "Credit"
		}

		if transactionType == "Debit" && currentBalance < amount {
			continue 
		}

		if transactionType == "Debit" {
			currentBalance -= amount
		} else {
			currentBalance += amount
		}

		daysBack := rand.Intn(365)
		txnDate := time.Now().AddDate(0, 0, -daysBack)

		txn := domain.Transaction{
			CustomerID:          customerID,
			FromAccount:         string(customer.AccountNo),
			ToAccount:           fmt.Sprintf("SYN-%s", uuid.New().String()[:8]),
			Amount:              amount,
			Remark:              fmt.Sprintf("Synthetic %s", transactionType),
			TransactionType:     transactionType,
			RequestId:           uuid.New().String(),
			Reference:           uuid.New().String()[:8],
			ThirdPartyReference: "",
			InstitutionId:       "",
			ClearedBalance:      currentBalance,
			TransactionDate:     txnDate,
			BillerId:            "",
			Status:              "completed",
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}
		txns = append(txns, txn)
	}

	for _, txn := range txns {
		_, err := uc.transactionRepo.Create(ctx, &txn)
		if err != nil {
			return nil, fmt.Errorf("failed to create synthetic transaction: %v", err)
		}
	}

	
	customer.CustomerBalance = currentBalance
	_, err = uc.customerRepo.Create(ctx, customer)
	if err != nil {
		return nil, fmt.Errorf("failed to update customer balance: %v", err)
	}

	return txns, nil
}

func (uc *TransactionUseCase) CalculateCustomerRating(ctx context.Context, customerID string) (float64, map[string]float64, error) {
	if customerID == "" {
		return 0, nil, errors.New("customer ID is required")
	}

	transactions, err := uc.transactionRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to fetch transactions: %v", err)
	}
	if len(transactions) == 0 {
		return 0, nil, errors.New("no transactions found for customer")
	}

	
	count := float64(len(transactions))
	totalVolume := 0.0
	var earliest, latest time.Time
	balances := make([]float64, 0, len(transactions))

	for i, tx := range transactions {
		totalVolume += tx.Amount
		if i == 0 {
			earliest = tx.TransactionDate
			latest = tx.TransactionDate
		} else {
			if tx.TransactionDate.Before(earliest) {
				earliest = tx.TransactionDate
			}
			if tx.TransactionDate.After(latest) {
				latest = tx.TransactionDate
			}
		}
		balances = append(balances, tx.ClearedBalance)
	}


	countScore := math.Min(count/10.0, 1.0)

	
	volumeScore := math.Min(totalVolume/100000.0, 1.0)

	
	durationDays := latest.Sub(earliest).Hours() / 24.0
	durationScore := math.Min(durationDays/365.0, 1.0)


	var meanBalance, variance float64
	if len(balances) > 0 {
		sum := 0.0
		for _, b := range balances {
			sum += b
		}
		meanBalance = sum / float64(len(balances))
		sumSquaredDiff := 0.0
		for _, b := range balances {
			sumSquaredDiff += math.Pow(b-meanBalance, 2)
		}
		variance = sumSquaredDiff / float64(len(balances))
	}
	stabilityScore := math.Max(1.0-variance/1000000.0, 0.0)


	totalRating := (0.3*countScore + 0.3*volumeScore + 0.2*durationScore + 0.2*stabilityScore) * 10.0
	totalRating = math.Round(math.Min(math.Max(totalRating, 0.0), 10.0)*100) / 100

	breakdown := map[string]float64{
		"count_score":     math.Round(countScore*100) / 100,
		"volume_score":    math.Round(volumeScore*100) / 100,
		"duration_score":  math.Round(durationScore*100) / 100,
		"stability_score": math.Round(stabilityScore*100) / 100,
	}

	return totalRating, breakdown, nil
}