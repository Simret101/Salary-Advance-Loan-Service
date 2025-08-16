package domain

import (
	"context"
	"io"
)

type RatingRepository interface {
	Create(ctx context.Context, rating Rating) (Rating, error)
	GetByCustomerID(ctx context.Context, customerID string) (Rating, error)
	GetAll(ctx context.Context) ([]Rating, error)
}

type RatingUseCase interface {
	CalculateRating(ctx context.Context, customerID string) (Rating, error)
	GetAllRatings(ctx context.Context) ([]Rating, error)
	ProcessTransactionsAndRatings(ctx context.Context, file io.Reader) ([]*Transaction, []Rating, []map[string]interface{}, error)
}
