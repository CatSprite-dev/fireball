package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrOperationsNotFound = errors.New("operations rows not found")

type OperationRepository struct {
	db *pgxpool.Pool
}

func NewOperationsRepository(pool *pgxpool.Pool) *OperationRepository {
	return &OperationRepository{db: pool}
}

func (r *OperationRepository) GetOperations(ctx context.Context, accountID string, from, to time.Time, types []string) (domain.UserOperations, error) {
	query := `
		SELECT date, type, instrument_type, figi, ticker, quantity, payment_currency, payment_units, payment_nano 
		FROM operations 
		WHERE account_id=$1 AND date>=$2 AND date<=$3 AND type = ANY($4)
		ORDER BY date ASC`

	rows, err := r.db.Query(ctx, query, accountID, from, to, types)
	if err != nil {
		return domain.UserOperations{}, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var operations domain.UserOperations
	operations.Items = []domain.Item{}

	for rows.Next() {
		var o domain.Item
		err := rows.Scan(
			&o.Date,
			&o.Type,
			&o.InstrumentType,
			&o.Figi,
			&o.Ticker,
			&o.Quantity,
			&o.Payment.Currency,
			&o.Payment.Units,
			&o.Payment.Nano,
		)
		if err != nil {
			return domain.UserOperations{}, fmt.Errorf("scan failed: %w", err)
		}
		operations.Items = append(operations.Items, o)
	}

	if err := rows.Err(); err != nil {
		return domain.UserOperations{}, fmt.Errorf("iteration error: %w", err)
	}

	if len(operations.Items) == 0 {
		return domain.UserOperations{}, ErrOperationsNotFound
	}

	return operations, nil
}

func (r *OperationRepository) PutOperations(ctx context.Context, accountID string, operations domain.UserOperations) error {
	query := `
		INSERT INTO operations (id, account_id, date, type, instrument_type, figi, ticker, quantity, payment_currency, payment_units, payment_nano)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) ON CONFLICT DO NOTHING
	`

	batch := &pgx.Batch{}
	for _, o := range operations.Items {
		batch.Queue(
			query,
			o.ID,
			accountID,
			o.Date,
			o.Type,
			o.InstrumentType,
			o.Figi,
			o.Ticker,
			o.Quantity,
			o.Payment.Currency,
			o.Payment.Units,
			o.Payment.Nano)
	}

	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	for range operations.Items {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("batch execution failed: %w", err)
		}
	}

	return nil
}
