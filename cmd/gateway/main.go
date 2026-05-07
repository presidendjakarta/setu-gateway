package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/presidendjakarta/setu-gateway/internal/config"
	"github.com/presidendjakarta/setu-gateway/internal/database"
	"github.com/presidendjakarta/setu-gateway/internal/gateway"
	"github.com/presidendjakarta/setu-gateway/internal/logger"
	"github.com/presidendjakarta/setu-gateway/internal/observability"
	"github.com/presidendjakarta/setu-gateway/internal/repository/postgres"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	cfg := config.New()
	configPath := "configs/gateway.yaml"
	
	if path := os.Getenv("SETU_CONFIG"); path != "" {
		configPath = path
	}

	if err := cfg.Load(configPath); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	rawConfig := cfg.Get()

	// Initialize logger
	log, err := logger.New(
		rawConfig.Logging.Level,
		rawConfig.Logging.Format,
		rawConfig.Logging.Output,
	)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Infow("Starting Setu API Gateway",
		"name", rawConfig.Gateway.Name,
		"version", rawConfig.Gateway.Version,
	)

	// Initialize database
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.NewPostgreSQL(ctx, &rawConfig.Database.Postgres)
	if err != nil {
		log.Fatalw("Failed to connect to database", "error", err)
	}
	defer db.Close()

	log.Infow("Database connection established")

	// Initialize repositories
	routeRepo := postgres.NewRouteRepository(db.Pool())

	// Load routes from database
	routes, err := routeRepo.List(ctx)
	if err != nil {
		log.Fatalw("Failed to load routes", "error", err)
	}

	log.Infow("Routes loaded from database", "count", len(routes))

	// Initialize observability
	metrics := observability.NewMetrics()
	healthChecker := observability.NewHealthChecker()

	// Register health checks
	healthChecker.RegisterCheck("database", func(ctx context.Context) observability.HealthStatus {
		if db.Health(ctx) {
			stats := db.Stats()
			return observability.HealthStatus{
				Status:  "healthy",
				Message: "Database connection OK",
				Details: map[string]interface{}{
					"total_conns":  stats["total_conns"],
					"idle_conns":   stats["idle_conns"],
					"acquired_conns": stats["acquired_conns"],
				},
			}
		}
		return observability.HealthStatus{
			Status:  "unhealthy",
			Message: "Database connection failed",
		}
	})

	// Update goroutine metrics periodically
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			metrics.UpdateGoroutines(0) // Will be implemented
		}
	}()

	// Initialize gateway
	gw, err := gateway.New(rawConfig, log, metrics, healthChecker)
	if err != nil {
		log.Fatalw("Failed to initialize gateway", "error", err)
	}
	defer gw.Close()

	// Reload routes into router
	if err := gw.ReloadRoutes(ctx, routes); err != nil {
		log.Fatalw("Failed to reload routes", "error", err)
	}

	// Start config watcher for hot-reload
	if err := cfg.Watch(configPath); err != nil {
		log.Warnw("Failed to start config watcher", "error", err)
	}

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", rawConfig.Server.Host, rawConfig.Server.Port)
	
	srv := &http.Server{
		Addr:         addr,
		Handler:      gw,
		ReadTimeout:  rawConfig.Server.ReadTimeout,
		WriteTimeout: rawConfig.Server.WriteTimeout,
		IdleTimeout:  rawConfig.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Infow("Gateway server starting", "address", addr)
		
		var err error
		if rawConfig.Server.TLS.Enabled {
			err = srv.ListenAndServeTLS(
				rawConfig.Server.TLS.CertFile,
				rawConfig.Server.TLS.KeyFile,
			)
		} else {
			err = srv.ListenAndServe()
		}
		
		if err != nil && err != http.ErrServerClosed {
			log.Fatalw("Server failed to start", "error", err)
		}
	}()

	// Start admin server if enabled
	if rawConfig.Admin.Enabled {
		go startAdminServer(rawConfig, log)
	}

	// Start metrics server if enabled
	if rawConfig.Metrics.Enabled {
		go startMetricsServer(rawConfig, log, metrics, healthChecker)
	}

	log.Infow("Gateway is ready to accept requests")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Infow("Shutting down gateway...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorw("Server forced to shutdown", "error", err)
	}

	log.Infow("Gateway stopped")
}

// startAdminServer starts the admin API server
func startAdminServer(cfg *config.RawConfig, log *logger.Logger) {
	addr := fmt.Sprintf("%s:%d", cfg.Admin.Host, cfg.Admin.Port)
	
	mux := http.NewServeMux()
	
	// Admin routes will be added here
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Infow("Admin server starting", "address", addr)
	
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Errorw("Admin server failed", "error", err)
	}
}

// startMetricsServer starts the Prometheus metrics server
func startMetricsServer(cfg *config.RawConfig, log *logger.Logger, metrics *observability.Metrics, healthChecker *observability.HealthChecker) {
	addr := fmt.Sprintf(":%d", cfg.Metrics.Port)
	
	mux := http.NewServeMux()
	
	// Prometheus metrics endpoint
	mux.Handle(cfg.Metrics.Path, promhttp.Handler())
	
	// Health check endpoints
	mux.HandleFunc("/health", healthChecker.WriteHealthCheck)
	mux.HandleFunc("/ready", healthChecker.WriteReady)
	mux.HandleFunc("/live", healthChecker.WriteLive)
	
	// Root endpoint with server info
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"service": "setu-gateway",
			"version": "%s",
			"metrics_path": "%s",
			"health_path": "/health",
			"ready_path": "/ready",
			"live_path": "/live"
		}`, cfg.Gateway.Version, cfg.Metrics.Path)
	})

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Infow("Metrics server starting", "address", addr)
	
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Errorw("Metrics server failed", "error", err)
	}
}
