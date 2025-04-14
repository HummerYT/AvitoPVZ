package products

import (
	"AvitoPVZ/internal/repository/products/mocks"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"AvitoPVZ/internal/models"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ----------------------
// Пользовательский матчтер для проверки подстроки в строке.
// ----------------------
type containsMatcher struct {
	substr string
}

func (m *containsMatcher) Matches(x interface{}) bool {
	s, ok := x.(string)
	if !ok {
		return false
	}
	return strings.Contains(s, m.substr)
}

func (m *containsMatcher) String() string {
	return fmt.Sprintf("contains substring %q", m.substr)
}

func Contains(substr string) gomock.Matcher {
	return &containsMatcher{substr: substr}
}

// ----------------------
// fakeRow для эмуляции результатов вызовов QueryRow
// ----------------------
type fakeRow struct {
	values []interface{}
	err    error
}

func (r *fakeRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.values) {
		return fmt.Errorf("ожидалось %d аргументов для Scan, получено %d", len(r.values), len(dest))
	}
	for i, v := range r.values {
		switch d := dest[i].(type) {
		case *uuid.UUID:
			*d = v.(uuid.UUID)
		case *time.Time:
			*d = v.(time.Time)
		default:
			return errors.New("неподдерживаемый тип в fakeRow.Scan")
		}
	}
	return nil
}

// TestCreateProductTransactional_NoActiveReception проверяет ситуацию, когда активная приёмка не найдена.
func TestCreateProductTransactional_NoActiveReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockTx := mocks.NewMockTx(ctrl)

	ctx := context.Background()
	pvzID := uuid.New()
	productType := models.TypeProduct("example_type")

	mockDB.EXPECT().
		BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}).
		Return(mockTx, nil)

	mockTx.EXPECT().
		QueryRow(ctx, Contains("FROM receiving"), pvzID).
		Return(&fakeRow{
			err: errors.New("активная приёмка не найдена"),
		}).
		Times(1)

	mockTx.EXPECT().
		Rollback(ctx).
		Return(pgx.ErrTxClosed).
		AnyTimes()

	repo := NewProductRepositoryPg(mockDB)
	_, err := repo.CreateProductTransactional(ctx, pvzID, productType)
	if err == nil {
		t.Fatal("ожидалась ошибка при отсутствии активной приёмки")
	}
}

// TestDeleteLastProductTransactional_NoActiveReception проверяет отсутствие активной приёмки.
func TestDeleteLastProductTransactional_NoActiveReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockTx := mocks.NewMockTx(ctrl)

	ctx := context.Background()
	pvzID := "pvz-123"

	mockDB.EXPECT().
		BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}).
		Return(mockTx, nil)

	mockTx.EXPECT().
		QueryRow(ctx, Contains("FROM receiving"), pvzID).
		Return(&fakeRow{
			err: errors.New("активная приёмка не найдена"),
		}).
		Times(1)

	mockTx.EXPECT().
		Rollback(ctx).
		Return(pgx.ErrTxClosed).
		AnyTimes()

	repo := NewProductRepositoryPg(mockDB)
	err := repo.DeleteLastProductTransactional(ctx, pvzID)
	if err == nil {
		t.Fatal("ожидалась ошибка при отсутствии активной приёмки")
	}
}

// TestDeleteLastProductTransactional_NoProduct проверяет ситуацию, когда товар для удаления не найден.
func TestDeleteLastProductTransactional_NoProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockTx := mocks.NewMockTx(ctrl)

	ctx := context.Background()
	pvzID := "pvz-123"
	recID := "rec-uuid"

	mockDB.EXPECT().
		BeginTx(gomock.Any(), gomock.Any()).
		Return(mockTx, nil)

	// Возвращаем активную приёмку.
	mockTx.EXPECT().
		QueryRow(ctx, gomock.Any(), gomock.Any()).
		Return(&fakeRow{
			values: []interface{}{recID},
		}).
		Times(1)

	mockTx.EXPECT().
		Rollback(ctx).
		Return(pgx.ErrTxClosed).
		AnyTimes()

	repo := NewProductRepositoryPg(mockDB)
	err := repo.DeleteLastProductTransactional(ctx, pvzID)
	if err == nil {
		t.Fatal("ожидалась ошибка при отсутствии товара для удаления")
	}
}
