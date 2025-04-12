package products

import (
	"context"
	"github.com/google/uuid"

	"AvitoPVZ/internal/models"
)

type ProductRepository interface {
	CreateProductTransactional(ctx context.Context, pvzID uuid.UUID, productType models.TypeProduct) (models.Product, error)
	DeleteLastProductTransactional(ctx context.Context, pvzID string) error
}

type ProductUseCase struct {
	repo ProductRepository
}

func NewProductUseCase(repo ProductRepository) *ProductUseCase {
	return &ProductUseCase{repo: repo}
}

func (uc *ProductUseCase) CreateProduct(ctx context.Context, pvzID uuid.UUID, productType models.TypeProduct) (models.Product, error) {
	return uc.repo.CreateProductTransactional(ctx, pvzID, productType)
}

func (uc *ProductUseCase) DeleteLastProduct(ctx context.Context, pvzID string) error {
	return uc.repo.DeleteLastProductTransactional(ctx, pvzID)
}
