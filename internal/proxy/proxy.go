package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/presidendjakarta/setu-gateway/internal/config"
	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

// Proxy implements reverse proxy with connection pooling
type Proxy struct {
	transport *http.Transport
	config    *config.ProxyConfig
	bufferPool sync.Pool
}

// New creates a new proxy instance
func New(cfg *config.ProxyConfig) *Proxy {
	// Create optimized HTTP transport
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          cfg.Transport.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.Transport.MaxIdleConnsPerHost,
		MaxConnsPerHost:       cfg.Transport.MaxConnsPerHost,
		IdleConnTimeout:       cfg.Transport.IdleConnTimeout,
		TLSHandshakeTimeout:   cfg.Transport.TLSHandshakeTimeout,
		ResponseHeaderTimeout: cfg.Transport.ResponseHeaderTimeout,
		ExpectContinueTimeout: cfg.Transport.ExpectContinueTimeout,
		DisableKeepAlives:     cfg.Transport.DisableKeepAlives,
		DisableCompression:    cfg.Transport.DisableCompression,
		ForceAttemptHTTP2:     true,
	}

	return &Proxy{
		transport: transport,
		config:    cfg,
		bufferPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 32*1024) // 32KB buffer
			},
		},
	}
}

// ServeHTTP proxies the request to upstream
func (p *Proxy) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request, route *types.Route, target *types.Target) error {
	// Build upstream URL
	upstreamURL := fmt.Sprintf("http://%s:%d", target.Host, target.Port)
	
	// Handle path stripping if configured
	targetPath := r.URL.Path
	if route.StripPath && route.Path != "/" {
		targetPath = strings.TrimPrefix(targetPath, route.Path)
		if !strings.HasPrefix(targetPath, "/") {
			targetPath = "/" + targetPath
		}
	}

	// Parse upstream URL
	upstream, err := url.Parse(upstreamURL)
	if err != nil {
		return types.WrapError(types.ErrCodeBadGateway, "Invalid upstream URL", http.StatusBadGateway, err)
	}

	// Modify request for upstream
	upstreamReq := r.Clone(ctx)
	upstreamReq.URL.Scheme = upstream.Scheme
	upstreamReq.URL.Host = upstream.Host
	upstreamReq.URL.Path = targetPath
	upstreamReq.Host = upstream.Host

	// Preserve original host if configured
	if route.PreserveHost {
		upstreamReq.Host = r.Host
	}

	// Add standard proxy headers
	upstreamReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
	upstreamReq.Header.Set("X-Forwarded-Proto", r.URL.Scheme)
	upstreamReq.Header.Set("X-Forwarded-Host", r.Host)
	upstreamReq.Header.Set("X-Real-IP", r.RemoteAddr)

	// Apply header transformations
	if route.Transform.HeaderRewrite != nil {
		p.applyHeaderRewrite(upstreamReq, route.Transform.HeaderRewrite)
	}

	// Create reverse proxy
	reverseProxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			*req = *upstreamReq
		},
		Transport:     p.transport,
		ErrorHandler:  p.errorHandler,
		BufferPool:    p,
		FlushInterval: -1, // Enable streaming
	}

	// Serve the request
	reverseProxy.ServeHTTP(w, upstreamReq)

	return nil
}

// Get implements sync.Pool interface
func (p *Proxy) Get() []byte {
	return p.bufferPool.Get().([]byte)
}

// Put implements sync.Pool interface
func (p *Proxy) Put(buf []byte) {
	p.bufferPool.Put(buf)
}

// applyHeaderRewrite applies header transformations
func (p *Proxy) applyHeaderRewrite(r *http.Request, rewrite *types.HeaderRewrite) {
	// Add headers
	for key, value := range rewrite.Add {
		r.Header.Set(key, value)
	}

	// Remove headers
	for _, key := range rewrite.Remove {
		r.Header.Del(key)
	}

	// Rename headers
	for oldKey, newKey := range rewrite.Rename {
		if values := r.Header[oldKey]; len(values) > 0 {
			r.Header.Del(oldKey)
			r.Header[newKey] = values
		}
	}
}

// errorHandler handles proxy errors
func (p *Proxy) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	// Log the error
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte(`{"error":"bad_gateway","message":"Upstream service unavailable"}`))
}

// Close closes the proxy and releases resources
func (p *Proxy) Close() error {
	p.transport.CloseIdleConnections()
	return nil
}

// Transport returns the underlying transport
func (p *Proxy) Transport() *http.Transport {
	return p.transport
}
