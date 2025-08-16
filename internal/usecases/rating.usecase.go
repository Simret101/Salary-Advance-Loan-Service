package usecases

import (
	"SalaryAdvance/internal/domain"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"time"
)

type RatingUseCase struct {
	ratingRepo         domain.RatingRepository
	transactionRepo    domain.TransactionRepository
	customerRepo       domain.CustomerRepository
	transactionUseCase domain.TransactionUseCase
}

func NewRatingUseCase(ratingRepo domain.RatingRepository, transactionRepo domain.TransactionRepository, customerRepo domain.CustomerRepository, transactionUseCase domain.TransactionUseCase) *RatingUseCase {
	return &RatingUseCase{
		ratingRepo:         ratingRepo,
		transactionRepo:    transactionRepo,
		customerRepo:       customerRepo,
		transactionUseCase: transactionUseCase,
	}
}

func (uc *RatingUseCase) CalculateRating(ctx context.Context, customerID string) (domain.Rating, error) {
	if customerID == "" {
		return domain.Rating{}, errors.New("customer ID is required")
	}

	txns, err := uc.transactionRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return domain.Rating{}, err
	}
	if len(txns) == 0 {
		return domain.Rating{}, errors.New("no transactions found for customer")
	}

	var txnPointers []*domain.Transaction
	for i := range txns {
		txnPointers = append(txnPointers, &txns[i])
	}

	countScore := uc.computeCountScore(len(txns))
	volumeScore := uc.computeVolumeScore(txnPointers)
	durationScore := uc.computeDurationScore(txnPointers)
	stabilityScore := uc.computeStabilityScore(txnPointers)

	ratingValue := (0.3*countScore + 0.3*volumeScore + 0.2*durationScore + 0.2*stabilityScore) * 10

	rating := domain.Rating{
		CustomerID: customerID,
		Score:      math.Round(ratingValue*100) / 100,
		Breakdown: struct {
			CountScore     float64 `json:"count_score"`
			VolumeScore    float64 `json:"volume_score"`
			DurationScore  float64 `json:"duration_score"`
			StabilityScore float64 `json:"stability_score"`
		}{
			CountScore:     countScore,
			VolumeScore:    volumeScore,
			DurationScore:  durationScore,
			StabilityScore: stabilityScore,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	created, err := uc.ratingRepo.Create(ctx, rating)
	if err != nil {
		return domain.Rating{}, err
	}
	return created, nil
}

func (uc *RatingUseCase) GetAllRatings(ctx context.Context) ([]domain.Rating, error) {
	return uc.ratingRepo.GetAll(ctx)
}

func (uc *RatingUseCase) ProcessTransactionsAndRatings(ctx context.Context, file io.Reader) ([]*domain.Transaction, []domain.Rating, []map[string]interface{}, error) {

	transactions, logs, err := uc.transactionUseCase.ImportTransactions(ctx, file)
	if err != nil {
		return nil, nil, logs, err
	}

	
	customers, err := uc.customerRepo.FindAll(ctx)
	if err != nil {
		return nil, nil, logs, fmt.Errorf("failed to fetch customers: %v", err)
	}

	
	for _, customer := range customers {
		txns, err := uc.transactionRepo.GetByCustomerID(ctx, customer.CustomerId)
		if err != nil {
			logs = append(logs, map[string]interface{}{
				"customer_id": customer.CustomerId,
				"errors":      []string{fmt.Sprintf("failed to check transactions: %v", err)},
			})
			continue
		}
		if len(txns) == 0 {
			syntheticTxns, err := uc.transactionUseCase.GenerateSyntheticTransactions(ctx, customer.CustomerId)
			if err != nil {
				logs = append(logs, map[string]interface{}{
					"customer_id": customer.CustomerId,
					"errors":      []string{fmt.Sprintf("failed to generate synthetic transactions: %v", err)},
				})
				continue
			}
			
			for i := range syntheticTxns {
				transactions = append(transactions, &syntheticTxns[i])
			}
		}
	}


	var ratings []domain.Rating
	for _, customer := range customers {
		rating, err := uc.CalculateRating(ctx, customer.CustomerId)
		if err != nil {
			logs = append(logs, map[string]interface{}{
				"customer_id": customer.CustomerId,
				"errors":      []string{fmt.Sprintf("failed to calculate rating: %v", err)},
			})
			continue
		}
		ratings = append(ratings, rating)
	}

	if len(transactions) == 0 && len(ratings) == 0 {
		return nil, nil, logs, errors.New("no transactions or ratings processed")
	}

	return transactions, ratings, logs, nil
}

func (uc *RatingUseCase) computeCountScore(count int) float64 {
	score := float64(count) / 100
	if score > 1 {
		score = 1
	}
	return score
}

func (uc *RatingUseCase) computeVolumeScore(txns []*domain.Transaction) float64 {
	var total float64
	for _, t := range txns {
		total += math.Abs(t.Amount)
	}
	score := total / 100000
	if score > 1 {
		score = 1
	}
	return score
}

func (uc *RatingUseCase) computeDurationScore(txns []*domain.Transaction) float64 {
	if len(txns) < 2 {
		return 0
	}
	var first, last time.Time
	for i, t := range txns {
		if i == 0 || t.TransactionDate.Before(first) {
			first = t.TransactionDate
		}
		if i == 0 || t.TransactionDate.After(last) {
			last = t.TransactionDate
		}
	}
	if first.IsZero() || last.IsZero() {
		return 0
	}
	durationDays := last.Sub(first).Hours() / 24
	score := durationDays / (365 * 5)
	if score > 1 {
		score = 1
	}
	return score
}

func (uc *RatingUseCase) computeStabilityScore(txns []*domain.Transaction) float64 {
	if len(txns) < 2 {
		return 1
	}
	balances := make([]float64, 0)
	for _, t := range txns {
		balances = append(balances, t.ClearedBalance)
	}
	var sum float64
	for _, b := range balances {
		sum += b
	}
	mean := sum / float64(len(balances))
	variance := 0.0
	for _, b := range balances {
		variance += (b - mean) * (b - mean)
	}
	variance /= float64(len(balances))
	stdDev := math.Sqrt(variance)
	score := (2000 - stdDev) / 2000
	if score < 0 {
		score = 0
	} else if score > 1 {
		score = 1
	}
	return score
}
