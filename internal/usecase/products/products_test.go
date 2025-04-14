package products_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/models"
	"AvitoPVZ/internal/usecase/products"
)

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

type ProductUseCaseSuite struct {
	suite.Suite
	repo *mockProductRepo
	uc   *products.ProductUseCase
}

func (s *ProductUseCaseSuite) SetupTest() {
	s.repo = new(mockProductRepo)
	s.uc = products.NewProductUseCase(s.repo)
}

func (s *ProductUseCaseSuite) Test_CreateProduct_Success() {
	pvzID := uuid.New()
	productType := models.TypeElectronic
	expectedProduct := models.Product{
		ID:          uuid.New(),
		ReceptionID: pvzID,
		Type:        productType,
		DateTime:    time.Now(),
	}

	s.repo.On("CreateProductTransactional", mock.Anything, pvzID, productType).Return(expectedProduct, nil)

	result, err := s.uc.CreateProduct(context.Background(), pvzID, productType)

	s.Require().NoError(err)
	s.Equal(expectedProduct, result)
	s.repo.AssertExpectations(s.T())
}

func (s *ProductUseCaseSuite) Test_CreateProduct_Error() {
	pvzID := uuid.New()
	productType := models.TypeClothes
	expectedErr := errors.New("database failure")

	s.repo.On("CreateProductTransactional", mock.Anything, pvzID, productType).Return(models.Product{}, expectedErr)

	result, err := s.uc.CreateProduct(context.Background(), pvzID, productType)

	s.Require().Error(err)
	s.Equal(expectedErr, err)
	s.Equal(models.Product{}, result)
	s.repo.AssertExpectations(s.T())
}

func (s *ProductUseCaseSuite) Test_DeleteLastProduct_Success() {
	pvzID := uuid.NewString()
	s.repo.On("DeleteLastProductTransactional", mock.Anything, pvzID).Return(nil)

	err := s.uc.DeleteLastProduct(context.Background(), pvzID)

	s.Require().NoError(err)
	s.repo.AssertExpectations(s.T())
}

func (s *ProductUseCaseSuite) Test_DeleteLastProduct_Error() {
	pvzID := uuid.NewString()
	expectedErr := errors.New("delete failed")

	s.repo.On("DeleteLastProductTransactional", mock.Anything, pvzID).Return(expectedErr)

	err := s.uc.DeleteLastProduct(context.Background(), pvzID)

	s.Require().Error(err)
	s.Equal(expectedErr, err)
	s.repo.AssertExpectations(s.T())
}

func TestProductUseCaseSuite(t *testing.T) {
	suite.Run(t, new(ProductUseCaseSuite))
}
