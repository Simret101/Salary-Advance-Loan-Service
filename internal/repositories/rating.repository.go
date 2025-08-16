package repositories

import (
	"SalaryAdvance/internal/domain"
	"context"

	"gorm.io/gorm"
)


type RatingRepositoryImpl struct {
	DB *gorm.DB
}

func NewRatingRepository(db *gorm.DB) *RatingRepositoryImpl {
	return &RatingRepositoryImpl{
		DB: db}
	
}

func (r *RatingRepositoryImpl) Create(ctx context.Context, rating domain.Rating) (domain.Rating, error) {
	
	ratingCopy := rating
	result := r.DB.WithContext(ctx).Create(&ratingCopy)
	if result.Error != nil {
		return domain.Rating{}, result.Error
	}
	return ratingCopy, nil 
}

func (r *RatingRepositoryImpl) GetByCustomerID(ctx context.Context, customerID string) (domain.Rating, error) {
	var rating domain.Rating
	if err := r.DB.WithContext(ctx).Where("customer_id = ?", customerID).First(&rating).Error; err != nil {
		return domain.Rating{}, err
	}
	return rating, nil
}

func (r *RatingRepositoryImpl) GetAll(ctx context.Context) ([]domain.Rating, error) {
	var ratings []domain.Rating
	if err := r.DB.WithContext(ctx).Find(&ratings).Error; err != nil {
		return nil, err
	}
	return ratings, nil
}
