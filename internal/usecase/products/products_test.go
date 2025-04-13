package products_test

import (
	"AvitoPVZ/internal/models"
	"AvitoPVZ/internal/usecase/products"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// --- Мок репозитория ---
type mockProductRepo struct {
	mock.Mock
}

func (m *mockProductRepo) CreateProductTransactional(ctx context.Context, pvzID uuid.UUID, productType models.TypeProduct) (models.Product, error) {
	args := m.Called(ctx, pvzID, productType)
	return args.Get(0).(models.Product), args.Error(1)
}

func (m *mockProductRepo) DeleteLastProductTransactional(ctx context.Context, pvzID string) error {
	args := m.Called(ctx, pvzID)
	return args.Error(0)
}

// --- Тестовая структура ---
type ProductUseCaseSuite struct {
	suite.Suite
	repo *mockProductRepo
	uc   *products.ProductUseCase
}

// --- Setup ---
func (s *ProductUseCaseSuite) SetupTest() {
	s.repo = new(mockProductRepo)
	s.uc = products.NewProductUseCase(s.repo)
}

// --- Успешное создание товара ---
func (s *ProductUseCaseSuite) Test_CreateProduct_Success() {
	// Arrange
	pvzID := uuid.New()
	productType := models.TypeElectronics
	expectedProduct := models.Product{
		ID:     uuid.New(),
		PVZ:    pvzID,
		Type:   productType,
		Status: models.ProductPending,
	}

	s.repo.On("CreateProductTransactional", mock.Anything, pvzID, productType).Return(expectedProduct, nil)

	// Act
	result, err := s.uc.CreateProduct(context.Background(), pvzID, productType)

	// Assert
	s.Require().NoError(err)
	s.Equal(expectedProduct, result)
	s.repo.AssertExpectations(s.T())
}

// --- Ошибка при создании товара ---
func (s *ProductUseCaseSuite) Test_CreateProduct_Error() {
	// Arrange
	pvzID := uuid.New()
	productType := models.TypeDocuments
	expectedErr := errors.New("database failure")

	s.repo.On("CreateProductTransactional", mock.Anything, pvzID, productType).Return(models.Product{}, expectedErr)

	// Act
	result, err := s.uc.CreateProduct(context.Background(), pvzID, productType)

	// Assert
	s.Require().Error(err)
	s.Equal(expectedErr, err)
	s.Equal(models.Product{}, result)
	s.repo.AssertExpectations(s.T())
}

// --- Успешное удаление последнего товара ---
func (s *ProductUseCaseSuite) Test_DeleteLastProduct_Success() {
	// Arrange
	pvzID := uuid.NewString()
	s.repo.On("DeleteLastProductTransactional", mock.Anything, pvzID).Return(nil)

	// Act
	err := s.uc.DeleteLastProduct(context.Background(), pvzID)

	// Assert
	s.Require().NoError(err)
	s.repo.AssertExpectations(s.T())
}

// --- Ошибка при удалении последнего товара ---
func (s *ProductUseCaseSuite) Test_DeleteLastProduct_Error() {
	// Arrange
	pvzID := uuid.NewString()
	expectedErr := errors.New("delete failed")

	s.repo.On("DeleteLastProductTransactional", mock.Anything, pvzID).Return(expectedErr)

	// Act
	err := s.uc.DeleteLastProduct(context.Background(), pvzID)

	// Assert
	s.Require().Error(err)
	s.Equal(expectedErr, err)
	s.repo.AssertExpectations(s.T())
}

// --- Запуск ---
func TestProductUseCaseSuite(t *testing.T) {
	suite.Run(t, new(ProductUseCaseSuite))
}
