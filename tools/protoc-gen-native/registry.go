package main

import "sync"

// PackageFunctionRegistry tracks which functions have been generated for a specific Go package
// to prevent duplicate function declarations when multiple proto files generate code
// into the same Go package
type PackageFunctionRegistry struct {
	mu                 sync.RWMutex
	generatedFunctions map[string]map[string]bool // packageName -> functionName -> exists
}

// NewPackageFunctionRegistry creates a new registry for tracking generated functions
func NewPackageFunctionRegistry() *PackageFunctionRegistry {
	return &PackageFunctionRegistry{
		generatedFunctions: make(map[string]map[string]bool),
	}
}

// RegisterFunction marks a function as generated for a specific package
// Returns true if the function was newly registered, false if it was already registered
func (r *PackageFunctionRegistry) RegisterFunction(packageName, functionName string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// initialize the package map if it doesn't exist
	if _, exists := r.generatedFunctions[packageName]; !exists {
		r.generatedFunctions[packageName] = make(map[string]bool)
	}

	// check if the function already exists
	if r.generatedFunctions[packageName][functionName] {
		return false
	}

	// register the function
	r.generatedFunctions[packageName][functionName] = true
	return true
}

// IsFunctionRegistered checks if a function has already been registered for a specific package
func (r *PackageFunctionRegistry) IsFunctionRegistered(packageName, functionName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// check if the package exists in the registry
	if funcs, exists := r.generatedFunctions[packageName]; exists {
		// check if the function exists for this package
		return funcs[functionName]
	}

	return false
}

// GlobalRegistry is the singleton registry used across the plugin's execution
var GlobalRegistry = NewPackageFunctionRegistry()
