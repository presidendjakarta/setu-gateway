package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

type rateLimitRepo struct {
	pool *pgxpool.Pool
}

func NewRateLimitRepository(pool *pgxpool.Pool) *rateLimitRepo {
	return &rateLimitRepo{pool: pool}
}

func (r *rateLimitRepo) Create(ctx context.Context, limit *types.RateLimit) error {
	query := `
		INSERT INTO rate_limits (id, route_id, requests_per_second, burst_size, key_type, key_header, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err := r.pool.Exec(ctx, query, limit.ID, limit.RouteID, limit.RequestsPerSecond, limit.BurstSize, limit.KeyType, limit.KeyHeader, limit.Enabled)
	if err != nil {
		return fmt.Errorf("failed to create rate limit: %w", err)
	}
	return nil
}

func (r *rateLimitRepo) GetByID(ctx context.Context, id string) (*types.RateLimit, error) {
	query := `
		SELECT id, route_id, requests_per_second, burst_size, key_type, key_header, enabled, created_at, updated_at
		FROM rate_limits
		WHERE id = $1
	`
	limit := &types.RateLimit{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&limit.ID, &limit.RouteID, &limit.RequestsPerSecond, &limit.BurstSize, &limit.KeyType,
		&limit.KeyHeader, &limit.Enabled, &limit.CreatedAt, &limit.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("rate limit not found")
		}
		return nil, fmt.Errorf("failed to get rate limit: %w", err)
	}
	return limit, nil
}

func (r *rateLimitRepo) GetByRouteID(ctx context.Context, routeID string) (*types.RateLimit, error) {
	query := `
		SELECT id, route_id, requests_per_second, burst_size, key_type, key_header, enabled, created_at, updated_at
		FROM rate_limits
		WHERE route_id = $1
	`
	limit := &types.RateLimit{}
	err := r.pool.QueryRow(ctx, query, routeID).Scan(
		&limit.ID, &limit.RouteID, &limit.RequestsPerSecond, &limit.BurstSize, &limit.KeyType,
		&limit.KeyHeader, &limit.Enabled, &limit.CreatedAt, &limit.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("rate limit not found")
		}
		return nil, fmt.Errorf("failed to get rate limit: %w", err)
	}
	return limit, nil
}

func (r *rateLimitRepo) Update(ctx context.Context, limit *types.RateLimit) error {
	query := `
		UPDATE rate_limits
		SET route_id = $2, requests_per_second = $3, burst_size = $4, key_type = $5, key_header = $6, enabled = $7, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, limit.ID, limit.RouteID, limit.RequestsPerSecond, limit.BurstSize, limit.KeyType, limit.KeyHeader, limit.Enabled)
	if err != nil {
		return fmt.Errorf("failed to update rate limit: %w", err)
	}
	return nil
}

func (r *rateLimitRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM rate_limits WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete rate limit: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("rate limit not found")
	}
	return nil
}

func (r *rateLimitRepo) List(ctx context.Context) ([]*types.RateLimit, error) {
	query := `
		SELECT id, route_id, requests_per_second, burst_size, key_type, key_header, enabled, created_at, updated_at
		FROM rate_limits
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query rate limits: %w", err)
	}
	defer rows.Close()

	var limits []*types.RateLimit
	for rows.Next() {
		limit := &types.RateLimit{}
		err := rows.Scan(
			&limit.ID, &limit.RouteID, &limit.RequestsPerSecond, &limit.BurstSize, &limit.KeyType,
			&limit.KeyHeader, &limit.Enabled, &limit.CreatedAt, &limit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rate limit: %w", err)
		}
		limits = append(limits, limit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rate limits: %w", err)
	}

	return limits, nil
}
