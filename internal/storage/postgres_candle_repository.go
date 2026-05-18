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

var ErrCandlesNotFound = errors.New("candles rows not found")

type CandleRepository struct {
	db *pgxpool.Pool
}

func NewCandleRepository(pool *pgxpool.Pool) *CandleRepository {
	return &CandleRepository{db: pool}
}

func (cr *CandleRepository) GetCandles(ctx context.Context, figi string, interval string, from time.Time, to time.Time) ([]domain.Candle, error) {
	query := `
		SELECT time, open_units, open_nano, close_units, close_nano 
		FROM candles 
		WHERE figi=$1 AND interval=$2 AND time>=$3 AND time<=$4 
		ORDER BY time ASC`

	rows, err := cr.db.Query(ctx, query, figi, interval, from, to)
	if err != nil {
		return []domain.Candle{}, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var candles []domain.Candle

	for rows.Next() {
		var c domain.Candle
		err := rows.Scan(
			&c.Time,
			&c.Open.Units,
			&c.Open.Nano,
			&c.Close.Units,
			&c.Close.Nano,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		candles = append(candles, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration error: %w", err)
	}

	if len(candles) == 0 {
		return []domain.Candle{}, ErrCandlesNotFound
	}

	return candles, nil
}

func (cr *CandleRepository) PutCandles(ctx context.Context, figi string, interval string, candles []domain.Candle) error {
	query := `
		INSERT INTO candles (figi, interval, time, open_units, open_nano, close_units, close_nano)
		VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING
	`

	batch := &pgx.Batch{}
	for _, c := range candles {
		batch.Queue(
			query,
			figi,
			interval,
			c.Time,
			c.Open.Units,
			c.Open.Nano,
			c.Close.Units,
			c.Close.Nano,
		)
	}

	results := cr.db.SendBatch(ctx, batch)
	defer results.Close()

	for range candles {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("batch execution failed: %w", err)
		}
	}

	return nil
}
