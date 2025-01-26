package providers

import "fmt"

// SecretProvider defines what a provider must implement
type SecretProvider interface {
	MigrateAll(sourceVaultURL, destinationVaultURL string) error
}

// Constructor type for dynamic provider creation
type ProviderConstructor func() (SecretProvider, error)

// Global registry of providers
var registry = make(map[string]ProviderConstructor)

// Register adds a provider to the registry
func Register(name string, constructor ProviderConstructor) {
	registry[name] = constructor
}

// GetProvider returns a provider by name
func GetProvider(name string) (SecretProvider, error) {
	constructor, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return constructor()
}
