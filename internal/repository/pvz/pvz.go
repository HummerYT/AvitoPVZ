package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"

	"AvitoPVZ/internal/models"
	"github.com/google/uuid"
)

type PvzRepositoryPostgres struct {
	pool *pgxpool.Pool
}

func NewPVZRepositoryPostgres(db *pgxpool.Pool) *PvzRepositoryPostgres {
	return &PvzRepositoryPostgres{pool: db}
}

func (r *PvzRepositoryPostgres) Create(ctx context.Context, city models.PVZCity) (models.PVZ, error) {
	id := uuid.NewString()
	registrationDate := time.Now()

	query := `
        INSERT INTO pickup_point (id, registration_date, city)
        VALUES ($1, $2, $3)
        RETURNING id, registration_date, city
    `

	var pvz models.PVZ
	err := r.pool.QueryRow(ctx, query, id, registrationDate, city).
		Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
	if err != nil {
		return models.PVZ{}, fmt.Errorf("failed to insert PVZ: %w", err)
	}

	return pvz, nil
}
func (r *PvzRepositoryPostgres) GetPVZData(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]models.PVZData, error) {
	var pvzList []models.PVZ
	var args []interface{}
	query := ""
	argIdx := 1

	if startDate != nil || endDate != nil {
		query = `SELECT DISTINCT p.id, p.registration_date, p.city
			FROM pickup_point p
			JOIN receiving r ON p.id = r.pickup_point_id
			WHERE 1=1`
		if startDate != nil {
			query += fmt.Sprintf(" AND r.receiving_datetime >= $%d", argIdx)
			args = append(args, *startDate)
			argIdx++
		}
		if endDate != nil {
			query += fmt.Sprintf(" AND r.receiving_datetime <= $%d", argIdx)
			args = append(args, *endDate)
			argIdx++
		}
		query += fmt.Sprintf(" ORDER BY p.registration_date LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, limit, (page-1)*limit)
	} else {
		query = fmt.Sprintf(`SELECT id, registration_date, city FROM pickup_point ORDER BY registration_date LIMIT $1 OFFSET $2`)
		args = append(args, limit, (page-1)*limit)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query pvz: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p models.PVZ
		if err := rows.Scan(&p.ID, &p.RegistrationDate, &p.City); err != nil {
			return nil, fmt.Errorf("scan pvz: %w", err)
		}
		pvzList = append(pvzList, p)
	}

	if len(pvzList) == 0 {
		return []models.PVZData{}, nil
	}

	var results []models.PVZData
	for _, p := range pvzList {
		recvQuery := `SELECT id, receiving_datetime, pickup_point_id, status FROM receiving WHERE pickup_point_id = $1`
		recvArgs := []interface{}{p.ID}
		argPosition := 2
		if startDate != nil {
			recvQuery += fmt.Sprintf(" AND receiving_datetime >= $%d", argPosition)
			recvArgs = append(recvArgs, *startDate)
			argPosition++
		}
		if endDate != nil {
			recvQuery += fmt.Sprintf(" AND receiving_datetime <= $%d", argPosition)
			recvArgs = append(recvArgs, *endDate)
			argPosition++
		}
		recvQuery += " ORDER BY receiving_datetime"

		recvRows, err := r.pool.Query(ctx, recvQuery, recvArgs...)
		if err != nil {
			return nil, fmt.Errorf("query receptions: %w", err)
		}

		var recDataList []models.ReceptionData
		for recvRows.Next() {
			var rec models.Reception
			if err := recvRows.Scan(&rec.ID, &rec.DateTime, &rec.PvzID, &rec.Status); err != nil {
				recvRows.Close()
				return nil, fmt.Errorf("scan reception: %w", err)
			}

			prodQuery := `SELECT id, accepted_datetime, product_type, receiving_id FROM goods WHERE receiving_id = $1 ORDER BY accepted_datetime`
			prodRows, err := r.pool.Query(ctx, prodQuery, rec.ID)
			if err != nil {
				recvRows.Close()
				return nil, fmt.Errorf("query products: %w", err)
			}
			var products []models.Product
			for prodRows.Next() {
				var prod models.Product
				if err := prodRows.Scan(&prod.ID, &prod.DateTime, &prod.Type, &prod.ReceptionID); err != nil {
					prodRows.Close()
					recvRows.Close()
					return nil, fmt.Errorf("scan product: %w", err)
				}
				products = append(products, prod)
			}
			prodRows.Close()

			recData := models.ReceptionData{
				Reception: rec,
				Products:  products,
			}
			recDataList = append(recDataList, recData)
		}
		recvRows.Close()

		if len(recDataList) > 0 {
			pvzData := models.PVZData{
				PVZ:        p,
				Receptions: recDataList,
			}
			results = append(results, pvzData)
		}
	}

	return results, nil
}
