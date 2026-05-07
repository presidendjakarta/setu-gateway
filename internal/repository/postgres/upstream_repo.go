package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

type upstreamRepo struct {
	pool *pgxpool.Pool
}

func NewUpstreamRepository(pool *pgxpool.Pool) *upstreamRepo {
	return &upstreamRepo{pool: pool}
}

func (r *upstreamRepo) Create(ctx context.Context, upstream *types.Upstream) error {
	query := `
		INSERT INTO upstreams (id, name, description, algorithm, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`
	_, err := r.pool.Exec(ctx, query, upstream.ID, upstream.Name, upstream.Description, upstream.Algorithm, upstream.Enabled)
	if err != nil {
		return fmt.Errorf("failed to create upstream: %w", err)
	}
	return nil
}

func (r *upstreamRepo) GetByID(ctx context.Context, id string) (*types.Upstream, error) {
	query := `
		SELECT id, name, description, algorithm, enabled, created_at, updated_at
		FROM upstreams
		WHERE id = $1
	`
	upstream := &types.Upstream{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&upstream.ID, &upstream.Name, &upstream.Description, &upstream.Algorithm,
		&upstream.Enabled, &upstream.CreatedAt, &upstream.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("upstream not found")
		}
		return nil, fmt.Errorf("failed to get upstream: %w", err)
	}
	return upstream, nil
}

func (r *upstreamRepo) GetByName(ctx context.Context, name string) (*types.Upstream, error) {
	query := `
		SELECT id, name, description, algorithm, enabled, created_at, updated_at
		FROM upstreams
		WHERE name = $1
	`
	upstream := &types.Upstream{}
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&upstream.ID, &upstream.Name, &upstream.Description, &upstream.Algorithm,
		&upstream.Enabled, &upstream.CreatedAt, &upstream.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("upstream not found")
		}
		return nil, fmt.Errorf("failed to get upstream: %w", err)
	}
	return upstream, nil
}

func (r *upstreamRepo) Update(ctx context.Context, upstream *types.Upstream) error {
	query := `
		UPDATE upstreams
		SET name = $2, description = $3, algorithm = $4, enabled = $5, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, upstream.ID, upstream.Name, upstream.Description, upstream.Algorithm, upstream.Enabled)
	if err != nil {
		return fmt.Errorf("failed to update upstream: %w", err)
	}
	return nil
}

func (r *upstreamRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM upstreams WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete upstream: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("upstream not found")
	}
	return nil
}

func (r *upstreamRepo) List(ctx context.Context) ([]*types.Upstream, error) {
	query := `
		SELECT id, name, description, algorithm, enabled, created_at, updated_at
		FROM upstreams
		WHERE enabled = true
		ORDER BY name
	`
	return r.queryUpstreams(ctx, query)
}

func (r *upstreamRepo) ListAll(ctx context.Context) ([]*types.Upstream, error) {
	query := `
		SELECT id, name, description, algorithm, enabled, created_at, updated_at
		FROM upstreams
		ORDER BY name
	`
	return r.queryUpstreams(ctx, query)
}

func (r *upstreamRepo) queryUpstreams(ctx context.Context, query string, args ...interface{}) ([]*types.Upstream, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query upstreams: %w", err)
	}
	defer rows.Close()

	var upstreams []*types.Upstream
	for rows.Next() {
		upstream := &types.Upstream{}
		err := rows.Scan(
			&upstream.ID, &upstream.Name, &upstream.Description, &upstream.Algorithm,
			&upstream.Enabled, &upstream.CreatedAt, &upstream.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan upstream: %w", err)
		}
		upstreams = append(upstreams, upstream)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating upstreams: %w", err)
	}

	return upstreams, nil
}
