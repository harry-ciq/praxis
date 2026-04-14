package provider

// Registry holds all registered achievement providers.
type Registry struct {
	providers map[string]AchievementProvider
}

// NewRegistry creates a new empty provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]AchievementProvider),
	}
}

// Register adds a provider to the registry.
func (r *Registry) Register(p AchievementProvider) {
	r.providers[p.ID()] = p
}

// Get retrieves a provider by its ID.
func (r *Registry) Get(id string) (AchievementProvider, bool) {
	p, ok := r.providers[id]
	return p, ok
}

// All returns all registered providers.
func (r *Registry) All() []AchievementProvider {
	result := make([]AchievementProvider, 0, len(r.providers))
	for _, p := range r.providers {
		result = append(result, p)
	}
	return result
}
