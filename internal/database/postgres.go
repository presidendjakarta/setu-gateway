package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/presidendjakarta/setu-gateway/internal/config"
)

// PostgreSQL represents a PostgreSQL connection pool
type PostgreSQL struct {
	pool *pgxpool.Pool
	cfg  *config.PostgresConfig
}

// NewPostgreSQL creates a new PostgreSQL connection pool
func NewPostgreSQL(ctx context.Context, cfg *config.PostgresConfig) (*PostgreSQL, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime
	poolConfig.HealthCheckPeriod = 30 * time.Second

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgreSQL{
		pool: pool,
		cfg:  cfg,
	}, nil
}

// Pool returns the underlying connection pool
func (p *PostgreSQL) Pool() *pgxpool.Pool {
	return p.pool
}

// Close closes the connection pool
func (p *PostgreSQL) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}

// Health checks if the database connection is healthy
func (p *PostgreSQL) Health(ctx context.Context) bool {
	if p.pool == nil {
		return false
	}
	return p.pool.Ping(ctx) == nil
}

// Stats returns connection pool statistics
func (p *PostgreSQL) Stats() map[string]int32 {
	if p.pool == nil {
		return nil
	}
	
	stat := p.pool.Stat()
	return map[string]int32{
		"acquire_count":            int32(stat.AcquireCount()),
		"acquired_conns":           stat.AcquiredConns(),
		"cancelled_acquire_count":  int32(stat.CanceledAcquireCount()),
		"constructing_conns":       stat.ConstructingConns(),
		"empty_acquire_count":      int32(stat.EmptyAcquireCount()),
		"idle_conns":               stat.IdleConns(),
		"max_conns":                stat.MaxConns(),
		"total_conns":              stat.TotalConns(),
	}
}
