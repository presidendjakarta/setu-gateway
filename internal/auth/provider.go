package auth

import (
	"context"
	"net/http"

	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

// AuthProvider is the interface all authentication providers must implement
type AuthProvider interface {
	// Name returns the provider name
	Name() string
	
	// Type returns the provider type
	Type() types.AuthProviderType
	
	// Authenticate authenticates the request
	Authenticate(ctx context.Context, r *http.Request, config types.AuthConfig) (*types.AuthResult, error)
	
	// Validate validates the provider configuration
	Validate(config types.AuthConfig) error
}
