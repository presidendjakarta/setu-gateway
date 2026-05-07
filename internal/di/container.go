package di

import (
	"fmt"
	"sync"
)

// Container is a simple dependency injection container
type Container struct {
	mu        sync.RWMutex
	services  map[string]interface{}
	factories map[string]func(c *Container) (interface{}, error)
	singleton map[string]bool
	cache     map[string]interface{}
}

// New creates a new DI container
func New() *Container {
	return &Container{
		services:  make(map[string]interface{}),
		factories: make(map[string]func(c *Container) (interface{}, error)),
		singleton: make(map[string]bool),
		cache:     make(map[string]interface{}),
	}
}

// Register registers a service instance
func (c *Container) Register(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
}

// RegisterSingleton registers a singleton service
func (c *Container) RegisterSingleton(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
	c.singleton[name] = true
}

// RegisterFactory registers a factory function
func (c *Container) RegisterFactory(name string, factory func(c *Container) (interface{}, error), singleton bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[name] = factory
	c.singleton[name] = singleton
}

// Resolve resolves a service by name
func (c *Container) Resolve(name string) (interface{}, error) {
	c.mu.RLock()
	// Check cache first
	if cached, ok := c.cache[name]; ok {
		c.mu.RUnlock()
		return cached, nil
	}
	
	// Check direct registration
	if service, ok := c.services[name]; ok {
		c.mu.RUnlock()
		return service, nil
	}
	
	// Check factory
	factory, hasFactory := c.factories[name]
	isSingleton := c.singleton[name]
	c.mu.RUnlock()
	
	if !hasFactory {
		return nil, fmt.Errorf("service not found: %s", name)
	}
	
	// Create factory lock to avoid holding read lock during creation
	c.mu.Lock()
	service, err := factory(c)
	if err != nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("failed to create service %s: %w", name, err)
	}
	
	// Cache if singleton
	if isSingleton {
		c.cache[name] = service
	}
	c.mu.Unlock()
	
	return service, nil
}

// ResolveAs resolves a service and casts it to the target type
func (c *Container) ResolveAs(name string, target interface{}) error {
	service, err := c.Resolve(name)
	if err != nil {
		return err
	}
	
	// Type assertion
 targetType := fmt.Sprintf("%T", target)
 if service == nil {
  return fmt.Errorf("service %s is nil", name)
 }
 
 // Use type assertion with pointer
 switch t := target.(type) {
 case **struct{}:
  if s, ok := service.(*struct{}); ok {
   *t = s
   return nil
  }
 }
 
 // Generic type assertion attempt (caller should use correct type)
 return fmt.Errorf("type assertion must be done by caller for service: %s", name)
}

// Has checks if a service exists
func (c *Container) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	_, inServices := c.services[name]
	_, inFactories := c.factories[name]
	_, inCache := c.cache[name]
	
	return inServices || inFactories || inCache
}

// Clear clears all cached instances (useful for testing)
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]interface{})
}

// Close closes all closable services
func (c *Container) Close() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	type Closer interface {
		Close() error
	}
	
	for name, service := range c.services {
		if closer, ok := service.(Closer); ok {
			if err := closer.Close(); err != nil {
				fmt.Printf("failed to close service %s: %v\n", name, err)
			}
		}
	}
	
	for name, service := range c.cache {
		if closer, ok := service.(Closer); ok {
			if err := closer.Close(); err != nil {
				fmt.Printf("failed to close cached service %s: %v\n", name, err)
			}
		}
	}
}
