package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/presidendjakarta/setu-gateway/internal/logger"
	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

// APIKeyProvider implements API Key authentication
type APIKeyProvider struct {
	name   string
	config types.AuthConfig
	logger *logger.Logger
}

// NewAPIKeyProvider creates a new API Key authentication provider
func NewAPIKeyProvider(name string, config types.AuthConfig, log *logger.Logger) *APIKeyProvider {
	return &APIKeyProvider{
		name:   name,
		config: config,
		logger: log,
	}
}

// Name returns the provider name
func (p *APIKeyProvider) Name() string {
	return p.name
}

// Type returns the provider type
func (p *APIKeyProvider) Type() types.AuthProviderType {
	return types.AuthProviderAPIKey
}

// Authenticate validates API key from request
func (p *APIKeyProvider) Authenticate(ctx context.Context, r *http.Request, config types.AuthConfig) (*types.AuthResult, error) {
	// Extract API key from request
	apiKey, err := p.extractAPIKey(r, config)
	if err != nil {
		return nil, err
	}

	// Validate API key (in production, check against database)
	valid, keyMetadata, err := p.validateAPIKey(ctx, apiKey)
	if err != nil {
		p.logger.Warnw("API key validation failed",
			"provider", p.name,
			"error", err,
		)
		return nil, types.WrapError(types.ErrCodeAuthFailed, "Invalid API key", http.StatusUnauthorized, err)
	}

	if !valid {
		return nil, types.NewGatewayError(types.ErrCodeAuthFailed, "Invalid API key", http.StatusUnauthorized)
	}

	// Build auth result
	result := &types.AuthResult{
		Authenticated: true,
		ProviderID:    p.name,
		Subject:       apiKey[:8] + "...", // Masked key
		Claims: map[string]interface{}{
			"api_key_prefix": apiKey[:8],
		},
		Metadata: keyMetadata,
	}

	p.logger.Debugw("API key authentication successful",
		"provider", p.name,
		"key_prefix", apiKey[:8],
	)

	return result, nil
}

// Validate validates the provider configuration
func (p *APIKeyProvider) Validate(config types.AuthConfig) error {
	if config.HeaderName == "" {
		config.HeaderName = "X-API-Key"
	}

	return nil
}

// extractAPIKey extracts API key from request
func (p *APIKeyProvider) extractAPIKey(r *http.Request, config types.AuthConfig) (string, error) {
	headerName := config.HeaderName
	if headerName == "" {
		headerName = "X-API-Key"
	}

	// Try header first
	if apiKey := r.Header.Get(headerName); apiKey != "" {
		// Remove prefix if configured
		if config.Prefix != "" {
			if strings.HasPrefix(apiKey, config.Prefix) {
				return apiKey[len(config.Prefix):], nil
			}
			return "", types.NewGatewayError(types.ErrCodeAuthFailed, "Invalid API key format", http.StatusUnauthorized)
		}
		return apiKey, nil
	}

	// Try query parameter
	if config.QueryParam != "" {
		if apiKey := r.URL.Query().Get(config.QueryParam); apiKey != "" {
			return apiKey, nil
		}
	}

	return "", types.NewGatewayError(types.ErrCodeAuthFailed, "No API key provided", http.StatusUnauthorized)
}

// validateAPIKey validates API key against database
func (p *APIKeyProvider) validateAPIKey(ctx context.Context, apiKey string) (bool, map[string]string, error) {
	// TODO: Implement actual API key validation
	// - Query database for API key
	// - Check if key is active
	// - Check rate limits
	// - Get key metadata (owner, permissions, etc.)
	
	// Placeholder: accept any key with length >= 32
	if len(apiKey) < 32 {
		return false, nil, nil
	}

	// Mock metadata
	metadata := map[string]string{
		"key_type":    "api_key",
		"owner":       "user@example.com",
		"permissions": "read,write",
		"rate_limit":  "1000/hour",
	}

	return true, metadata, nil
}
