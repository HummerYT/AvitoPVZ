package pvz

import (
	"context"
	"fmt"
	"time"

	"AvitoPVZ/internal/models"
)

type repository interface {
	Create(ctx context.Context, city models.PVZCity) (models.PVZ, error)
	GetPVZData(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]models.PVZData, error)
}

type UseCase struct {
	pvzRepo repository
}

func NewPVZUseCase(repo repository) *UseCase {
	return &UseCase{
		pvzRepo: repo,
	}
}

func (uc *UseCase) CreatePVZ(ctx context.Context, city models.PVZCity) (models.PVZ, error) {
	newPVZ, err := uc.pvzRepo.Create(ctx, city)
	if err != nil {
		return models.PVZ{}, fmt.Errorf("failed to create PVZ: %w", err)
	}
	return newPVZ, nil
}

func (uc *UseCase) GetPVZData(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]models.PVZData, error) {
	return uc.pvzRepo.GetPVZData(ctx, startDate, endDate, page, limit)
}
