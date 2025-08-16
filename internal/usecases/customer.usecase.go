package usecases

import (
	"SalaryAdvance/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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
