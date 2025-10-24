package foundation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
)

type ProviderState struct {
	instance   foundation.ServiceProvider
	registered bool
	booted     bool
}

var _ foundation.ProviderRepository = (*ProviderRepository)(nil)

type ProviderRepository struct {
	allProviders        map[string]*ProviderState
	configuredProviders []foundation.ServiceProvider
	configuredSet       bool
}

func NewProviderRepository() *ProviderRepository {
	return &ProviderRepository{
		allProviders: make(map[string]*ProviderState),
	}
}

func (r *ProviderRepository) Boot(app foundation.Application, providers []foundation.ServiceProvider) {
	for _, provider := range providers {
		name := r.getProviderName(provider)
		state, exists := r.allProviders[name]

		if exists && state.registered && !state.booted {
			state.instance.Boot(app)
			state.booted = true
		}
	}
}

func (r *ProviderRepository) GetBooted() []foundation.ServiceProvider {
	booted := make([]foundation.ServiceProvider, 0, len(r.allProviders))
	for _, state := range r.allProviders {
		if state.booted {
			booted = append(booted, state.instance)
		}
	}
	return booted
}

func (r *ProviderRepository) LoadConfigured(app foundation.Application) []foundation.ServiceProvider {
	if r.configuredSet {
		return r.configuredProviders
	}
	if r.configuredProviders != nil {
		return r.configuredProviders
	}

	configFacade := app.MakeConfig()
	if configFacade == nil {
		color.Warningln(errors.ConfigFacadeNotSet.Error())
		r.configuredProviders = []foundation.ServiceProvider{}
		return r.configuredProviders
	}

	providers, ok := configFacade.Get("app.providers").([]foundation.ServiceProvider)
	if !ok {
		color.Warningln(errors.ConsoleProvidersNotArray.Error())
		r.configuredProviders = []foundation.ServiceProvider{}
		return r.configuredProviders
	}

	r.configuredProviders = r.sort(providers)

	return r.configuredProviders
}

func (r *ProviderRepository) Register(app foundation.Application, providers []foundation.ServiceProvider) []foundation.ServiceProvider {
	sortedProviders := r.sort(providers)

	for _, provider := range sortedProviders {
		name := r.getProviderName(provider)
		state, exists := r.allProviders[name]

		if !exists {
			state = &ProviderState{instance: provider}
			r.allProviders[name] = state
		}

		if !state.registered {
			state.instance.Register(app)
			state.registered = true
		}
	}

	return sortedProviders
}

func (r *ProviderRepository) ResetConfiguredCache() {
	r.configuredProviders = nil
	r.configuredSet = false
}

func (r *ProviderRepository) SetConfigured(providers []foundation.ServiceProvider) {
	r.configuredProviders = r.sort(providers)
	r.configuredSet = true
}

func (r *ProviderRepository) getRelationship(provider foundation.ServiceProvider) binding.Relationship {
	if p, ok := provider.(foundation.ServiceProviderWithRelations); ok {
		return p.Relationship()
	}
	return binding.Relationship{}
}

func (r *ProviderRepository) getProviderName(provider foundation.ServiceProvider) string {
	return fmt.Sprintf("%T", provider)
}

// sort performs a topological sort on a list of providers to ensure
// providers with dependencies are registered and booted *after*
// the providers they depend on.
func (r *ProviderRepository) sort(providers []foundation.ServiceProvider) []foundation.ServiceProvider {
	if len(providers) == 0 {
		return providers
	}

	// These maps build a directed graph of dependencies.
	bindingToProvider := make(map[string]foundation.ServiceProvider)
	providerToVirtualBinding := make(map[foundation.ServiceProvider]string)
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	virtualBindingCounter := 0

	// --- Build Graph Nodes ---
	// Identify all "nodes" (bindings) in the graph.
	for _, provider := range providers {
		relationship := r.getRelationship(provider)
		bindings := relationship.Bindings
		dependencies := relationship.Dependencies
		provideFor := relationship.ProvideFor

		if len(bindings) > 0 {
			for _, b := range bindings {
				bindingToProvider[b] = provider
				inDegree[b] = 0
			}
		} else if len(dependencies) > 0 || len(provideFor) > 0 {
			// This provider has no bindings but has relationships.
			// We create a "virtual" node to represent it in the graph
			// so its dependencies can be sorted.
			virtualBinding := fmt.Sprintf("__virtual_%d", virtualBindingCounter)
			virtualBindingCounter++
			bindingToProvider[virtualBinding] = provider
			providerToVirtualBinding[provider] = virtualBinding
			inDegree[virtualBinding] = 0
		}
	}

	// --- Build Graph Edges ---
	// Connect the nodes based on 'Dependencies' and 'ProvideFor'.
	for _, provider := range providers {
		relationship := r.getRelationship(provider)
		bindings := relationship.Bindings
		dependencies := relationship.Dependencies
		provideFor := relationship.ProvideFor

		var providerBindings []string
		if len(bindings) > 0 {
			providerBindings = bindings
		} else if virtualBinding, exists := providerToVirtualBinding[provider]; exists {
			providerBindings = []string{virtualBinding}
		}

		if len(providerBindings) == 0 {
			// Provider is independent and not part of the sort.
			continue
		}

		for _, b := range providerBindings {
			// Edge: dep -> binding
			for _, dep := range dependencies {
				if _, exists := bindingToProvider[dep]; exists {
					graph[dep] = append(graph[dep], b)
					inDegree[b]++
				}
			}

			// Edge: binding -> provideForBinding
			for _, provideForBinding := range provideFor {
				if _, exists := bindingToProvider[provideForBinding]; exists {
					graph[b] = append(graph[b], provideForBinding)
					inDegree[provideForBinding]++
				}
			}
		}
	}

	// --- Topological Sort (Kahn's Algorithm) ---
	queue := make([]string, 0, len(inDegree))
	for b, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, b)
		}
	}

	result := make([]string, 0, len(inDegree))
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// --- Cycle Detection & Result Reconstruction ---
	if len(result) != len(inDegree) {
		// A cycle exists, the dependency order is impossible.
		cycle := r.detectCycle(graph, bindingToProvider)
		if len(cycle) > 0 {
			panic(errors.ServiceProviderCycle.Args(strings.Join(cycle, " -> ")))
		}
		panic(errors.ServiceProviderCycle.Args("unknown cycle detected"))
	}

	sortedProviders := make([]foundation.ServiceProvider, 0, len(providers))
	used := make(map[foundation.ServiceProvider]bool)

	for _, b := range result {
		provider := bindingToProvider[b]
		// Use a map to prevent adding the same provider multiple
		// times if it had more than one binding (e.g., "log" and "logger").
		if !used[provider] {
			sortedProviders = append(sortedProviders, provider)
			used[provider] = true
		}
	}

	// Add any remaining providers that were not part of the graph.
	for _, provider := range providers {
		if !used[provider] {
			sortedProviders = append(sortedProviders, provider)
		}
	}

	return sortedProviders
}

// detectCycle uses a Depth-First Search (DFS) to find and report a
// cycle in the provider dependency graph.
func (r *ProviderRepository) detectCycle(graph map[string][]string, bindingToProvider map[string]foundation.ServiceProvider) []string {
	// visited: Nodes already processed and known to be safe.
	// recStack: Nodes currently in our DFS recursion stack.
	// If we hit a node already in recStack, we've found a cycle.
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)
	cycle := make([]string, 0)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				// Cycle detected! Reconstruct the path for the error message.
				cycleStart := -1
				for i, p := range path {
					if p == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart != -1 {
					cycle = append(cycle, path[cycleStart:]...)
					// Append the start node to the end to show the full loop.
					cycle = append(cycle, neighbor)
				}
				return true
			}
		}

		// Backtrack
		recStack[node] = false
		path = path[:len(path)-1]
		return false
	}

	// Sort nodes to make cycle detection deterministic.
	var nodes []string
	for node := range graph {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	for _, node := range nodes {
		if !visited[node] {
			if dfs(node) {
				break
			}
		}
	}

	if len(cycle) == 0 {
		return nil
	}

	// Convert the list of *bindings* into user-friendly *provider names*.
	var cycleProviders []string
	providerSet := make(map[string]struct{})

	for _, b := range cycle {
		if provider, exists := bindingToProvider[b]; exists {
			providerName := r.getProviderName(provider)
			cycleProviders = append(cycleProviders, providerName)
			providerSet[providerName] = struct{}{}
		}
	}

	// Handle a specific edge case for self-loops.
	if len(cycleProviders) == 2 && cycleProviders[0] == cycleProviders[1] {
		if len(providerSet) == 1 && len(cycle) > 2 {
			return cycleProviders[0:1]
		}
	}

	return cycleProviders
}
