package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

// routeRepo implements repository.RouteRepository
type routeRepo struct {
	pool *pgxpool.Pool
}

// NewRouteRepository creates a new route repository
func NewRouteRepository(pool *pgxpool.Pool) *routeRepo {
	return &routeRepo{pool: pool}
}

// Create creates a new route
func (r *routeRepo) Create(ctx context.Context, route *types.Route) error {
	query := `
		INSERT INTO routes (
			id, name, description, path, path_type, methods, strip_path, 
			preserve_host, enabled, priority, upstream_id, auth_chain, 
			plugins, rate_limit_id, transform_config, timeout_interval, 
			retry_enabled, circuit_breaker, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, 
			$15, $16, $17, $18, $19, $20
		)
	`

	now := time.Now()
	_, err := r.pool.Exec(ctx, query,
		route.ID,
		route.Name,
		route.Description,
		route.Path,
		route.PathType,
		route.Methods,
		route.StripPath,
		route.PreserveHost,
		route.Enabled,
		route.Priority,
		route.UpstreamID,
		route.AuthChain,
		route.Plugins,
		route.RateLimitID,
		route.Transform,
		route.Timeout,
		route.RetryEnabled,
		route.CircuitBreaker,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create route: %w", err)
	}

	return nil
}

// GetByID retrieves a route by ID
func (r *routeRepo) GetByID(ctx context.Context, id string) (*types.Route, error) {
	query := `
		SELECT id, name, description, path, path_type, methods, strip_path,
			preserve_host, enabled, priority, upstream_id, auth_chain,
			plugins, rate_limit_id, transform_config, timeout_interval,
			retry_enabled, circuit_breaker, created_at, updated_at
		FROM routes
		WHERE id = $1
	`

	route := &types.Route{}
	var transformBytes []byte
	
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&route.ID,
		&route.Name,
		&route.Description,
		&route.Path,
		&route.PathType,
		&route.Methods,
		&route.StripPath,
		&route.PreserveHost,
		&route.Enabled,
		&route.Priority,
		&route.UpstreamID,
		&route.AuthChain,
		&route.Plugins,
		&route.RateLimitID,
		&transformBytes,
		&route.Timeout,
		&route.RetryEnabled,
		&route.CircuitBreaker,
		&route.CreatedAt,
		&route.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("route not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get route: %w", err)
	}
	
	// Unmarshal transform_config if not null
	if transformBytes != nil && len(transformBytes) > 0 {
		if err := json.Unmarshal(transformBytes, &route.Transform); err != nil {
			return nil, fmt.Errorf("failed to unmarshal transform_config: %w", err)
		}
	}

	return route, nil
}

// GetByPath retrieves routes by path pattern
func (r *routeRepo) GetByPath(ctx context.Context, path string) ([]*types.Route, error) {
	query := `
		SELECT id, name, description, path, path_type, methods, strip_path,
			preserve_host, enabled, priority, upstream_id, auth_chain,
			plugins, rate_limit_id, transform_config, timeout_interval,
			retry_enabled, circuit_breaker, created_at, updated_at
		FROM routes
		WHERE path = $1 OR path LIKE $2
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, path, path+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query routes: %w", err)
	}
	defer rows.Close()

	var routes []*types.Route
	for rows.Next() {
		route := &types.Route{}
		var transformBytes []byte
		
		err := rows.Scan(
			&route.ID,
			&route.Name,
			&route.Description,
			&route.Path,
			&route.PathType,
			&route.Methods,
			&route.StripPath,
			&route.PreserveHost,
			&route.Enabled,
			&route.Priority,
			&route.UpstreamID,
			&route.AuthChain,
			&route.Plugins,
			&route.RateLimitID,
			&transformBytes,
			&route.Timeout,
			&route.RetryEnabled,
			&route.CircuitBreaker,
			&route.CreatedAt,
			&route.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		
		// Unmarshal transform_config if not null
		if transformBytes != nil && len(transformBytes) > 0 {
			if err := json.Unmarshal(transformBytes, &route.Transform); err != nil {
				return nil, fmt.Errorf("failed to unmarshal transform_config: %w", err)
			}
		}
		
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating routes: %w", err)
	}

	return routes, nil
}

// Update updates an existing route
func (r *routeRepo) Update(ctx context.Context, route *types.Route) error {
	query := `
		UPDATE routes SET
			name = $2, description = $3, path = $4, path_type = $5,
			methods = $6, strip_path = $7, preserve_host = $8,
			enabled = $9, priority = $10, upstream_id = $11,
			auth_chain = $12, plugins = $13, rate_limit_id = $14,
			transform_config = $15, timeout_interval = $16,
			retry_enabled = $17, circuit_breaker = $18, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		route.ID,
		route.Name,
		route.Description,
		route.Path,
		route.PathType,
		route.Methods,
		route.StripPath,
		route.PreserveHost,
		route.Enabled,
		route.Priority,
		route.UpstreamID,
		route.AuthChain,
		route.Plugins,
		route.RateLimitID,
		route.Transform,
		route.Timeout,
		route.RetryEnabled,
		route.CircuitBreaker,
	)

	if err != nil {
		return fmt.Errorf("failed to update route: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("route not found: %s", route.ID)
	}

	return nil
}

// Delete deletes a route by ID
func (r *routeRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM routes WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete route: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("route not found: %s", id)
	}

	return nil
}

// List retrieves all enabled routes
func (r *routeRepo) List(ctx context.Context) ([]*types.Route, error) {
	query := `
		SELECT id, name, description, path, path_type, methods, strip_path,
			preserve_host, enabled, priority, upstream_id, auth_chain,
			plugins, rate_limit_id, transform_config, timeout_interval,
			retry_enabled, circuit_breaker, created_at, updated_at
		FROM routes
		WHERE enabled = true
		ORDER BY priority DESC, created_at DESC
	`

	return r.queryRoutes(ctx, query)
}

// ListAll retrieves all routes (including disabled)
func (r *routeRepo) ListAll(ctx context.Context) ([]*types.Route, error) {
	query := `
		SELECT id, name, description, path, path_type, methods, strip_path,
			preserve_host, enabled, priority, upstream_id, auth_chain,
			plugins, rate_limit_id, transform_config, timeout_interval,
			retry_enabled, circuit_breaker, created_at, updated_at
		FROM routes
		ORDER BY priority DESC, created_at DESC
	`

	return r.queryRoutes(ctx, query)
}

// Count returns the total number of routes
func (r *routeRepo) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM routes`

	var count int64
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count routes: %w", err)
	}

	return count, nil
}

// queryRoutes is a helper function to execute route queries
func (r *routeRepo) queryRoutes(ctx context.Context, query string, args ...interface{}) ([]*types.Route, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query routes: %w", err)
	}
	defer rows.Close()

	var routes []*types.Route
	for rows.Next() {
		route := &types.Route{}
		// Use pointer to []byte for nullable JSONB fields
		var transformBytes *[]byte
		
		err := rows.Scan(
			&route.ID,
			&route.Name,
			&route.Description,
			&route.Path,
			&route.PathType,
			&route.Methods,
			&route.StripPath,
			&route.PreserveHost,
			&route.Enabled,
			&route.Priority,
			&route.UpstreamID,
			&route.AuthChain,
			&route.Plugins,
			&route.RateLimitID,
			&transformBytes,
			&route.Timeout,
			&route.RetryEnabled,
			&route.CircuitBreaker,
			&route.CreatedAt,
			&route.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		
		// Unmarshal transform_config if not null
		if transformBytes != nil && len(*transformBytes) > 0 {
			if err := json.Unmarshal(*transformBytes, &route.Transform); err != nil {
				return nil, fmt.Errorf("failed to unmarshal transform_config: %w", err)
			}
		}
		
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating routes: %w", err)
	}

	return routes, nil
}
