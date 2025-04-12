package receptions

import (
	"context"
	"github.com/google/uuid"

	"AvitoPVZ/internal/models"
)

type ReceptionRepository interface {
	CreateReceptionTransactional(ctx context.Context, pvzID uuid.UUID) (models.Reception, error)
	CloseLastReceptionTransactional(ctx context.Context, pvzID string) (models.Reception, error)
}

type ReceptionUseCase struct {
	repo ReceptionRepository
}

func NewReceptionUseCase(repo ReceptionRepository) *ReceptionUseCase {
	return &ReceptionUseCase{repo: repo}
}

func (uc *ReceptionUseCase) CreateReception(ctx context.Context, pvzID uuid.UUID) (models.Reception, error) {
	return uc.repo.CreateReceptionTransactional(ctx, pvzID)
}

func (uc *ReceptionUseCase) CloseLastReception(ctx context.Context, pvzID string) (models.Reception, error) {
	return uc.repo.CloseLastReceptionTransactional(ctx, pvzID)
}
