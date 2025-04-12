package products

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"AvitoPVZ/internal/models"
)

type ProductRepositoryPg struct {
	pool *pgxpool.Pool
}

func NewProductRepositoryPg(pool *pgxpool.Pool) *ProductRepositoryPg {
	return &ProductRepositoryPg{pool: pool}
}

func (r *ProductRepositoryPg) CreateProductTransactional(ctx context.Context, pvzID uuid.UUID, productType models.TypeProduct) (models.Product, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return models.Product{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queryReception := `
		SELECT id, receiving_datetime, pickup_point_id, status
		FROM receiving
		WHERE pickup_point_id = $1 AND status = 'in_progress'
		ORDER BY receiving_datetime DESC
		LIMIT 1
		FOR UPDATE
	`
	var recID string
	var recTime time.Time
	var recPvzID string
	var recStatus string

	err = tx.QueryRow(ctx, queryReception, pvzID).
		Scan(&recID, &recTime, &recPvzID, &recStatus)
	if err != nil {
		return models.Product{}, fmt.Errorf("нет активной приёмки для pvzID=%s: %w", pvzID, err)
	}

	productID := uuid.NewString()
	acceptedTime := time.Now()

	queryInsert := `
		INSERT INTO goods (id, receiving_id, accepted_datetime, product_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, accepted_datetime, product_type, receiving_id
	`
	var prod models.Product
	err = tx.QueryRow(ctx, queryInsert, productID, recID, acceptedTime, productType).
		Scan(&prod.ID, &prod.DateTime, &prod.Type, &prod.ReceptionID)
	if err != nil {
		return models.Product{}, fmt.Errorf("невозможно добавить товар: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return models.Product{}, fmt.Errorf("commit transaction: %w", err)
	}

	return prod, nil
}

func (r *ProductRepositoryPg) DeleteLastProductTransactional(ctx context.Context, pvzID string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queryReception := `
		SELECT id 
		FROM receiving
		WHERE pickup_point_id = $1 AND status = 'in_progress'
		ORDER BY receiving_datetime DESC
		LIMIT 1
		FOR UPDATE
	`
	var recID string
	err = tx.QueryRow(ctx, queryReception, pvzID).Scan(&recID)
	if err != nil {
		return fmt.Errorf("нет активной приемки для pvzID=%s: %w", pvzID, err)
	}

	queryProduct := `
		SELECT id
		FROM goods
		WHERE receiving_id = $1
		ORDER BY accepted_datetime DESC
		LIMIT 1
		FOR UPDATE
	`
	var prodID string
	err = tx.QueryRow(ctx, queryProduct, recID).Scan(&prodID)
	if err != nil {
		return fmt.Errorf("нет товаров для удаления в приемке: %w", err)
	}

	deleteQuery := `DELETE FROM goods WHERE id = $1`
	_, err = tx.Exec(ctx, deleteQuery, prodID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении товара: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
