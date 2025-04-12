package receptions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"AvitoPVZ/internal/models"
)

type ReceptionRepositoryPg struct {
	pool *pgxpool.Pool
}

func NewReceptionRepositoryPg(pool *pgxpool.Pool) *ReceptionRepositoryPg {
	return &ReceptionRepositoryPg{pool: pool}
}

func (r *ReceptionRepositoryPg) CreateReceptionTransactional(ctx context.Context, pvzID uuid.UUID) (models.Reception, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return models.Reception{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	var active models.Reception
	querySelect := `
  SELECT id, receiving_datetime, pickup_point_id, status
  FROM receiving
  WHERE pickup_point_id = $1 AND status = 'in_progress'
  FOR UPDATE
 `
	err = tx.QueryRow(ctx, querySelect, pvzID).Scan(&active.ID, &active.DateTime, &active.PvzID, &active.Status)
	if err == nil {
		return models.Reception{}, fmt.Errorf("active reception already exists")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return models.Reception{}, fmt.Errorf("query active reception: %w", err)
	}

	id := uuid.NewString()
	now := time.Now()
	insertQuery := `
  INSERT INTO receiving (id, receiving_datetime, pickup_point_id, status)
  VALUES ($1, $2, $3, $4)
  RETURNING id, receiving_datetime, pickup_point_id, status
 `
	var newRec models.Reception
	err = tx.QueryRow(ctx, insertQuery, id, now, pvzID, "in_progress").
		Scan(&newRec.ID, &newRec.DateTime, &newRec.PvzID, &newRec.Status)
	if err != nil {
		return models.Reception{}, fmt.Errorf("insert reception: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return models.Reception{}, fmt.Errorf("commit transaction: %w", err)
	}

	return newRec, nil
}

func (r *ReceptionRepositoryPg) CloseLastReceptionTransactional(ctx context.Context, pvzID string) (models.Reception, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return models.Reception{}, fmt.Errorf("begin transaction: %w", err)
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
	var rec models.Reception
	err = tx.QueryRow(ctx, queryReception, pvzID).Scan(&rec.ID, &rec.DateTime, &rec.PvzID, &rec.Status)
	if err != nil {
		return models.Reception{}, fmt.Errorf("активная приемка не найдена для pvzID=%s: %w", pvzID, err)
	}

	updateQuery := `
		UPDATE receiving
		SET status = 'close'
		WHERE id = $1
		RETURNING id, receiving_datetime, pickup_point_id, status
	`
	var updatedRec models.Reception
	err = tx.QueryRow(ctx, updateQuery, rec.ID).
		Scan(&updatedRec.ID, &updatedRec.DateTime, &updatedRec.PvzID, &updatedRec.Status)
	if err != nil {
		return models.Reception{}, fmt.Errorf("невозможно закрыть приемку: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return models.Reception{}, fmt.Errorf("commit transaction: %w", err)
	}

	return updatedRec, nil
}
