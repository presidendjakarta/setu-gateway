package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

type targetRepo struct {
	pool *pgxpool.Pool
}

func NewTargetRepository(pool *pgxpool.Pool) *targetRepo {
	return &targetRepo{pool: pool}
}

func (r *targetRepo) Create(ctx context.Context, target *types.Target) error {
	query := `
		INSERT INTO targets (id, upstream_id, host, port, weight, enabled, healthy, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err := r.pool.Exec(ctx, query, target.ID, target.UpstreamID, target.Host, target.Port, target.Weight, target.Enabled, target.Healthy)
	if err != nil {
		return fmt.Errorf("failed to create target: %w", err)
	}
	return nil
}

func (r *targetRepo) GetByID(ctx context.Context, id string) (*types.Target, error) {
	query := `
		SELECT id, upstream_id, host, port, weight, enabled, healthy
		FROM targets
		WHERE id = $1
	`
	target := &types.Target{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&target.ID, &target.UpstreamID, &target.Host, &target.Port, &target.Weight,
		&target.Enabled, &target.Healthy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("target not found")
		}
		return nil, fmt.Errorf("failed to get target: %w", err)
	}
	return target, nil
}

func (r *targetRepo) GetByUpstreamID(ctx context.Context, upstreamID string) ([]*types.Target, error) {
	query := `
		SELECT id, upstream_id, host, port, weight, enabled, healthy
		FROM targets
		WHERE upstream_id = $1
		ORDER BY weight DESC
	`
	rows, err := r.pool.Query(ctx, query, upstreamID)
	if err != nil {
		return nil, fmt.Errorf("failed to query targets: %w", err)
	}
	defer rows.Close()

	var targets []*types.Target
	for rows.Next() {
		target := &types.Target{}
		err := rows.Scan(
			&target.ID, &target.UpstreamID, &target.Host, &target.Port, &target.Weight,
			&target.Enabled, &target.Healthy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan target: %w", err)
		}
		targets = append(targets, target)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating targets: %w", err)
	}

	return targets, nil
}

func (r *targetRepo) Update(ctx context.Context, target *types.Target) error {
	query := `
		UPDATE targets
		SET host = $2, port = $3, weight = $4, enabled = $5, healthy = $6, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, target.ID, target.Host, target.Port, target.Weight, target.Enabled, target.Healthy)
	if err != nil {
		return fmt.Errorf("failed to update target: %w", err)
	}
	return nil
}

func (r *targetRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM targets WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete target: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("target not found")
	}
	return nil
}

func (r *targetRepo) UpdateHealth(ctx context.Context, id string, healthy bool) error {
	query := `UPDATE targets SET healthy = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, healthy, id)
	if err != nil {
		return fmt.Errorf("failed to update target health: %w", err)
	}
	return nil
}
