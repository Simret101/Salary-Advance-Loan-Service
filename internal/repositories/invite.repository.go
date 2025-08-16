package repositories

import (
	"context"
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/pkg/config"

	"gorm.io/gorm"
)

type InviteRepositoryImpl struct {
	DB *gorm.DB
}

func NewInviteRepository(db *gorm.DB) *InviteRepositoryImpl {
	return &InviteRepositoryImpl{DB: db}
}

func (r *InviteRepositoryImpl) Create(ctx context.Context, invite *domain.Invite) (*domain.Invite, error) {
	result := r.DB.WithContext(ctx).Create(invite)
	if result.Error != nil {
		return nil, result.Error
	}
	return invite, nil
}

func (r *InviteRepositoryImpl) FindByToken(ctx context.Context, token string) (*domain.Invite, error) {
	var invite domain.Invite
	if err := r.DB.WithContext(ctx).Where("token = ?", token).First(&invite).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, config.ErrNotFound
		}
		return nil, config.ErrInternalServer
	}
	return &invite, nil
}

func (r *InviteRepositoryImpl) Update(ctx context.Context, invite *domain.Invite) (*domain.Invite, error) {
	result := r.DB.WithContext(ctx).Save(invite)
	if result.Error != nil {
		return nil, result.Error
	}
	return invite, nil
}