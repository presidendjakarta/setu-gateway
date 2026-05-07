package router

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

// node represents a radix tree node
type node struct {
	path     string
	children map[string]*node
	route    *types.Route
	isWild   bool
}

// Router implements high-performance route matching using radix tree
type Router struct {
	mu       sync.RWMutex
	tree     *node
	routes   map[string]*types.Route
	priority []*types.Route
}

// New creates a new router instance
func New() *Router {
	return &Router{
		tree:     &node{children: make(map[string]*node)},
		routes:   make(map[string]*types.Route),
		priority: make([]*types.Route, 0),
	}
}

// Match finds the best matching route for the request
func (r *Router) Match(req *http.Request) (*types.Route, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path := req.URL.Path
	method := req.Method

	// First, try exact match from priority routes
	for _, route := range r.priority {
		if !route.Enabled {
			continue
		}

		// Check method
		if len(route.Methods) > 0 && !containsMethod(route.Methods, method) {
			continue
		}

		// Check path match based on type
		if r.matchPath(route, path) {
			return route, nil
		}
	}

	// Try tree-based matching
	if route := r.matchTree(path, method); route != nil {
		return route, nil
	}

	return nil, types.ErrRouteNotFound
}

// matchPath checks if a path matches the route based on path type
func (r *Router) matchPath(route *types.Route, path string) bool {
	switch route.PathType {
	case types.PathTypeExact:
		return route.Path == path
	case types.PathTypePrefix:
		return strings.HasPrefix(path, route.Path)
	case types.PathTypeWildcard:
		pattern := route.Path
		// Simple wildcard matching: * matches anything
		if pattern == "*" {
			return true
		}
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")
			return strings.HasPrefix(path, prefix)
		}
		return path == pattern
	case types.PathTypeRegex:
		// Regex matching would be implemented here
		// For now, fall back to prefix match
		return strings.HasPrefix(path, route.Path)
	default:
		return strings.HasPrefix(path, route.Path)
	}
}

// matchTree performs radix tree-based route matching
func (r *Router) matchTree(path string, method string) *types.Route {
	n := r.tree
	searchPath := path

	for len(searchPath) > 0 {
		found := false
		for prefix, child := range n.children {
			if strings.HasPrefix(searchPath, prefix) {
				n = child
				searchPath = searchPath[len(prefix):]
				found = true
				break
			}
		}

		if !found {
			// Try wildcard
			if n.isWild && n.route != nil {
				return n.route
			}
			break
		}
	}

	if n.route != nil && n.route.Enabled {
		if len(n.route.Methods) == 0 || containsMethod(n.route.Methods, method) {
			return n.route
		}
	}

	return nil
}

// AddRoute adds a new route to the router
func (r *Router) AddRoute(ctx context.Context, route *types.Route) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.routes[route.ID] = route
	r.rebuildPriority()

	return nil
}

// RemoveRoute removes a route from the router
func (r *Router) RemoveRoute(ctx context.Context, routeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.routes, routeID)
	r.rebuildPriority()

	return nil
}

// UpdateRoute updates an existing route
func (r *Router) UpdateRoute(ctx context.Context, route *types.Route) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.routes[route.ID] = route
	r.rebuildPriority()

	return nil
}

// GetRoute returns a route by ID
func (r *Router) GetRoute(routeID string) (*types.Route, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	route, exists := r.routes[routeID]
	if !exists {
		return nil, fmt.Errorf("route not found: %s", routeID)
	}

	return route, nil
}

// ListRoutes returns all routes
func (r *Router) ListRoutes() []*types.Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]*types.Route, 0, len(r.routes))
	for _, route := range r.routes {
		routes = append(routes, route)
	}

	return routes
}

// Reload reloads all routes
func (r *Router) Reload(ctx context.Context, routes []*types.Route) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Clear existing routes
	r.routes = make(map[string]*types.Route)
	r.priority = make([]*types.Route, 0)
	r.tree = &node{children: make(map[string]*node)}

	// Add new routes
	for _, route := range routes {
		r.routes[route.ID] = route
	}

	r.rebuildPriority()
	r.rebuildTree()

	return nil
}

// rebuildPriority rebuilds the priority-sorted route list
func (r *Router) rebuildPriority() {
	r.priority = make([]*types.Route, 0, len(r.routes))
	for _, route := range r.routes {
		if route.Enabled {
			r.priority = append(r.priority, route)
		}
	}

	// Sort by priority (higher first), then by path specificity
	sort.Slice(r.priority, func(i, j int) bool {
		if r.priority[i].Priority != r.priority[j].Priority {
			return r.priority[i].Priority > r.priority[j].Priority
		}
		// More specific paths first
		return len(r.priority[i].Path) > len(r.priority[j].Path)
	})
}

// rebuildTree rebuilds the radix tree
func (r *Router) rebuildTree() {
	r.tree = &node{children: make(map[string]*node)}

	for _, route := range r.priority {
		r.insertIntoTree(route)
	}
}

// insertIntoTree inserts a route into the radix tree
func (r *Router) insertIntoTree(route *types.Route) {
	n := r.tree
	path := route.Path

	for len(path) > 0 {
		found := false
		for prefix, child := range n.children {
			if hasCommonPrefix(prefix, path) {
				n = child
				path = path[len(prefix):]
				found = true
				break
			}
		}

		if !found {
			// Create new node
			newNode := &node{
				path:     path,
				children: make(map[string]*node),
				route:    route,
			}
			n.children[path] = newNode
			path = ""
		}
	}

	// Handle wildcard routes
	if route.PathType == types.PathTypeWildcard {
		n.isWild = true
		n.route = route
	}
}

// containsMethod checks if a method is in the list
func containsMethod(methods []string, method string) bool {
	for _, m := range methods {
		if m == method || m == "*" {
			return true
		}
	}
	return false
}

// hasCommonPrefix checks if two paths share a common prefix
func hasCommonPrefix(a, b string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	return a[0] == b[0]
}
