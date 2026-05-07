package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/presidendjakarta/setu-gateway/internal/logger"
	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

// JWTProvider implements JWT authentication
type JWTProvider struct {
	name   string
	config types.AuthConfig
	logger *logger.Logger
}

// NewJWTProvider creates a new JWT authentication provider
func NewJWTProvider(name string, config types.AuthConfig, log *logger.Logger) *JWTProvider {
	return &JWTProvider{
		name:   name,
		config: config,
		logger: log,
	}
}

// Name returns the provider name
func (p *JWTProvider) Name() string {
	return p.name
}

// Type returns the provider type
func (p *JWTProvider) Type() types.AuthProviderType {
	return types.AuthProviderJWT
}

// Authenticate validates JWT token from request
func (p *JWTProvider) Authenticate(ctx context.Context, r *http.Request, config types.AuthConfig) (*types.AuthResult, error) {
	// Extract token from request
	tokenString, err := p.extractToken(r, config)
	if err != nil {
		return nil, err
	}

	// Validate JWT token
	claims, err := p.validateToken(tokenString, config)
	if err != nil {
		p.logger.Warnw("JWT validation failed",
			"provider", p.name,
			"error", err,
		)
		return nil, types.WrapError(types.ErrCodeAuthFailed, "Invalid JWT token", http.StatusUnauthorized, err)
	}

	// Build auth result
	result := &types.AuthResult{
		Authenticated: true,
		ProviderID:    p.name,
		Subject:       claims.Subject,
		Claims:        claims.MapClaims,
		Metadata: map[string]string{
			"auth_type":   "jwt",
			"issuer":      claims.Issuer,
			"expires_at":  claims.ExpiresAt.Format(time.RFC3339),
		},
	}

	p.logger.Debugw("JWT authentication successful",
		"provider", p.name,
		"subject", claims.Subject,
		"issuer", claims.Issuer,
	)

	return result, nil
}

// Validate validates the provider configuration
func (p *JWTProvider) Validate(config types.AuthConfig) error {
	if config.Secret == "" && config.JWKSEndpoint == "" {
		return fmt.Errorf("JWT provider requires either secret or JWKS endpoint")
	}

	if len(config.Algorithms) == 0 {
		config.Algorithms = []string{"HS256"}
	}

	return nil
}

// extractToken extracts JWT token from request
func (p *JWTProvider) extractToken(r *http.Request, config types.AuthConfig) (string, error) {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Bearer token
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			return authHeader[7:], nil
		}
	}

	// Try query parameter
	if config.QueryParam != "" {
		if token := r.URL.Query().Get(config.QueryParam); token != "" {
			return token, nil
		}
	}

	// Try cookie
	if cookie, err := r.Cookie("token"); err == nil {
		return cookie.Value, nil
	}

	return "", types.NewGatewayError(types.ErrCodeAuthFailed, "No JWT token provided", http.StatusUnauthorized)
}

// validateToken validates JWT token and returns claims
func (p *JWTProvider) validateToken(tokenString string, config types.AuthConfig) (*JWTClaims, error) {
	// For now, return placeholder
	// In production, use github.com/golang-jwt/jwt/v5
	
	// TODO: Implement actual JWT validation
	// - Parse token
	// - Verify signature (HS256, RS256, ES256)
	// - Validate claims (exp, iss, aud, nbf)
	// - Support JWKS endpoint
	
	return &JWTClaims{
		Subject:   "user@example.com",
		Issuer:    "setu-gateway",
		ExpiresAt: time.Now().Add(time.Hour),
		MapClaims: map[string]interface{}{
			"sub":   "user@example.com",
			"iss":   "setu-gateway",
			"role":  "admin",
			"scope": "read write",
		},
	}, nil
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	Subject   string
	Issuer    string
	ExpiresAt time.Time
	MapClaims map[string]interface{}
}
