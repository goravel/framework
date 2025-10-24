package foundation

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

type ProviderRepositoryTestSuite struct {
	suite.Suite
	repository *ProviderRepository
	mockApp    *mocksfoundation.Application
}

func TestProviderRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ProviderRepositoryTestSuite))
}

func (s *ProviderRepositoryTestSuite) SetupTest() {
	s.mockApp = mocksfoundation.NewApplication(s.T())
	s.repository = NewProviderRepository()
}

func (s *ProviderRepositoryTestSuite) TestBoot() {
	mockProvider := mocksfoundation.NewServiceProvider(s.T())
	providers := []foundation.ServiceProvider{mockProvider}

	// Boot should not be called if not registered
	s.repository.Boot(s.mockApp, providers)
	mockProvider.AssertNotCalled(s.T(), "Boot", s.mockApp)

	mockProvider.EXPECT().Register(s.mockApp).Return().Once()
	s.repository.Register(s.mockApp, providers)

	mockProvider.EXPECT().Boot(s.mockApp).Return().Once()
	s.repository.Boot(s.mockApp, providers)

	s.True(s.repository.allProviders[s.repository.getProviderName(mockProvider)].booted)
}

func (s *ProviderRepositoryTestSuite) TestBoot_Idempotency() {
	mockProvider := mocksfoundation.NewServiceProvider(s.T())
	providers := []foundation.ServiceProvider{mockProvider}

	mockProvider.EXPECT().Register(s.mockApp).Return().Once()
	s.repository.Register(s.mockApp, providers)

	mockProvider.EXPECT().Boot(s.mockApp).Return().Once()

	s.repository.Boot(s.mockApp, providers)
	s.repository.Boot(s.mockApp, providers)

	s.True(s.repository.allProviders[s.repository.getProviderName(mockProvider)].booted)
}

func (s *ProviderRepositoryTestSuite) TestGetBooted() {
	providerA := &AServiceProvider{}
	providerB := &BServiceProvider{}
	providers := []foundation.ServiceProvider{providerA, providerB}

	s.repository.Register(s.mockApp, providers)

	// Boot only A.
	s.repository.Boot(s.mockApp, []foundation.ServiceProvider{providerA})

	// GetBooted should only return A
	booted := s.repository.GetBooted()
	s.Len(booted, 1, "Expected only one booted provider")
	s.Equal(providerA, booted[0], "The booted provider should be providerA")
}

func (s *ProviderRepositoryTestSuite) TestLoadConfigured() {
	providers := []foundation.ServiceProvider{&BServiceProvider{}, &AServiceProvider{}}
	expectedSorted := []foundation.ServiceProvider{&AServiceProvider{}, &BServiceProvider{}}

	mockConfig := mocksconfig.NewConfig(s.T())

	testCases := []struct {
		name              string
		setup             func()
		expectedProviders []foundation.ServiceProvider
		expectEmpty       bool
	}{
		{
			name: "Success: Load from config",
			setup: func() {
				s.repository.ResetConfiguredCache()

				s.mockApp.EXPECT().MakeConfig().Return(mockConfig).Once()
				mockConfig.EXPECT().Get("app.providers").Return(providers).Once()
			},
			expectedProviders: expectedSorted,
		},
		{
			name: "Success: Load from manual set",
			setup: func() {
				s.repository.SetConfigured(providers)
			},
			expectedProviders: expectedSorted,
		},
		{
			name: "Success: Load from cache",
			setup: func() {
				// Pre-populate the cache (e.g., from a previous call)
				s.repository.configuredProviders = expectedSorted
				s.repository.configuredSet = false // Ensure it's not a manual set
			},
			expectedProviders: expectedSorted,
		},
		{
			name: "Failure: Config facade is nil",
			setup: func() {
				s.repository.ResetConfiguredCache()
				s.mockApp.EXPECT().MakeConfig().Return(nil).Once()
			},
			expectEmpty: true,
		},
		{
			name: "Failure: Config Get fails",
			setup: func() {
				s.repository.ResetConfiguredCache()
				s.mockApp.EXPECT().MakeConfig().Return(mockConfig).Once()
				mockConfig.EXPECT().Get("app.providers").Return("not a slice").Once()
			},
			expectEmpty: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.mockApp.Mock = mock.Mock{}
			mockConfig.Mock = mock.Mock{}

			tc.setup()

			result := s.repository.LoadConfigured(s.mockApp)

			if tc.expectEmpty {
				s.Empty(result)
				s.NotNil(result, "Should return empty slice, not nil")
			} else {
				s.Equal(tc.expectedProviders, result)
			}

			s.mockApp.AssertExpectations(s.T())
			mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *ProviderRepositoryTestSuite) TestRegister() {
	mockProvider := mocksfoundation.NewServiceProvider(s.T())
	providers := []foundation.ServiceProvider{mockProvider}

	mockProvider.EXPECT().Register(s.mockApp).Return().Once()

	sorted := s.repository.Register(s.mockApp, providers)

	s.Equal(providers, sorted) // Only one provider, so sorted == original
	s.True(s.repository.allProviders[s.repository.getProviderName(mockProvider)].registered)
	s.False(s.repository.allProviders[s.repository.getProviderName(mockProvider)].booted)
}

func (s *ProviderRepositoryTestSuite) TestRegister_Idempotency() {
	mockProvider := mocksfoundation.NewServiceProvider(s.T())
	providers := []foundation.ServiceProvider{mockProvider}

	mockProvider.EXPECT().Register(s.mockApp).Return().Once()

	s.repository.Register(s.mockApp, providers)
	s.repository.Register(s.mockApp, providers)

	s.True(s.repository.allProviders[s.repository.getProviderName(mockProvider)].registered)
}

func (s *ProviderRepositoryTestSuite) TestResetConfiguredCache() {
	s.repository.configuredProviders = []foundation.ServiceProvider{}
	s.repository.configuredSet = true

	s.repository.ResetConfiguredCache()

	s.False(s.repository.configuredSet)
	s.Nil(s.repository.configuredProviders)
}

func (s *ProviderRepositoryTestSuite) TestSetConfigured() {
	providers := []foundation.ServiceProvider{&BServiceProvider{}, &AServiceProvider{}}
	expectedSorted := []foundation.ServiceProvider{&AServiceProvider{}, &BServiceProvider{}}

	s.repository.SetConfigured(providers)

	s.True(s.repository.configuredSet)
	s.Equal(expectedSorted, s.repository.configuredProviders)
}

func (s *ProviderRepositoryTestSuite) TestSort() {
	testCases := []struct {
		name      string
		providers []foundation.ServiceProvider
		expected  []foundation.ServiceProvider
		checkTopo bool
	}{
		{
			name: "not found basic dependency, should be sorted correctly",
			providers: []foundation.ServiceProvider{
				&CServiceProvider{},
				&BServiceProvider{},
			},
			expected: []foundation.ServiceProvider{
				&BServiceProvider{},
				&CServiceProvider{},
			},
		},
		{
			name: "BasicSorting",
			providers: []foundation.ServiceProvider{
				&BServiceProvider{},
				&CServiceProvider{},
				&AServiceProvider{},
			},
			expected: []foundation.ServiceProvider{
				&AServiceProvider{},
				&BServiceProvider{},
				&CServiceProvider{},
			},
			checkTopo: true,
		},
		{
			name: "SingleProvider",
			providers: []foundation.ServiceProvider{
				&BasicServiceProvider{},
			},
			expected: []foundation.ServiceProvider{
				&BasicServiceProvider{},
			},
			checkTopo: true,
		},
		{
			name:      "EmptyProviders",
			providers: []foundation.ServiceProvider{},
			expected:  []foundation.ServiceProvider{},
			checkTopo: true,
		},
		{
			name: "ProvideForRelationship",
			providers: []foundation.ServiceProvider{
				&ProvideForBServiceProvider{},
				&ProvideForAServiceProvider{},
			},
			expected: []foundation.ServiceProvider{
				&ProvideForAServiceProvider{},
				&ProvideForBServiceProvider{},
			},
			checkTopo: true,
		},
		{
			name: "SingleProviderWithMock",
			providers: []foundation.ServiceProvider{
				&MockProviderE{},
			},
			expected: []foundation.ServiceProvider{
				&MockProviderE{},
			},
		},
		{
			name: "EmptyDependencies",
			providers: []foundation.ServiceProvider{
				&MockProviderC{},
				&EmptyDependenciesProvider{},
			},
			expected: []foundation.ServiceProvider{
				&EmptyDependenciesProvider{},
				&MockProviderC{},
			},
			checkTopo: true,
		},
		{
			name: "EmptyProvideFor",
			providers: []foundation.ServiceProvider{
				&EmptyProvideForProvider{},
				&MockProviderA{},
			},
			expected: []foundation.ServiceProvider{
				&MockProviderA{},
				&EmptyProvideForProvider{},
			},
			checkTopo: true,
		},
		{
			name: "AllEmptyMethods",
			providers: []foundation.ServiceProvider{
				&AllEmptyProvider{},
				&MockProviderE{},
			},
			expected: []foundation.ServiceProvider{
				&MockProviderE{},
				&AllEmptyProvider{},
			},
			checkTopo: true,
		},
		{
			name: "MixedEmptyAndNonEmpty",
			providers: []foundation.ServiceProvider{
				&AllEmptyProvider{},
				&MockProviderC{},
				&EmptyDependenciesProvider{},
			},
			expected: []foundation.ServiceProvider{
				&EmptyDependenciesProvider{},
				&MockProviderC{},
				&AllEmptyProvider{},
			},
			checkTopo: true,
		},
		{
			name: "EmptyBindingsWithDependencies",
			providers: []foundation.ServiceProvider{
				&EmptyBindingsWithDependenciesProvider{},
				&MockProviderC{},
			},
			expected: []foundation.ServiceProvider{
				&MockProviderC{},
				&EmptyBindingsWithDependenciesProvider{},
			},
			checkTopo: true,
		},
		{
			name: "EmptyBindingsWithProvideFor",
			providers: []foundation.ServiceProvider{
				&MockProviderA{},
				&EmptyBindingsWithProvideForProvider{},
			},
			expected: []foundation.ServiceProvider{
				&EmptyBindingsWithProvideForProvider{},
				&MockProviderA{},
			},
			checkTopo: true,
		},
		{
			name: "EmptyBindingsWithBothDependenciesAndProvideFor",
			providers: []foundation.ServiceProvider{
				&EmptyBindingsWithBothProvider{},
				&MockProviderA{},
			},
			expected: []foundation.ServiceProvider{
				&EmptyBindingsWithBothProvider{},
				&MockProviderA{},
			},
		},
		{
			name: "MultipleEmptyBindingsProviders",
			providers: []foundation.ServiceProvider{
				&EmptyBindingsWithDependenciesProvider{},
				&EmptyBindingsWithProvideForProvider{},
				&MockProviderE{},
			},
			expected: []foundation.ServiceProvider{
				&EmptyBindingsWithDependenciesProvider{},
				&EmptyBindingsWithProvideForProvider{},
				&MockProviderE{},
			},
		},
		{
			name: "ComplexEmptyBindingsScenario",
			providers: []foundation.ServiceProvider{
				&EmptyBindingsWithDependenciesProvider{},
				&EmptyBindingsWithProvideForProvider{},
				&EmptyBindingsWithBothProvider{},
				&MockProviderE{},
				&AllEmptyProvider{},
			},
			expected: []foundation.ServiceProvider{
				&AllEmptyProvider{},
				&EmptyBindingsWithDependenciesProvider{},
				&EmptyBindingsWithProvideForProvider{},
				&MockProviderE{},
				&EmptyBindingsWithBothProvider{},
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			result := s.repository.sort(tt.providers)

			if tt.checkTopo {
				s.Assert().Equal(tt.expected, result)
				s.Assert().True(isTopologicalOrder(tt.providers, result), "Result is not a valid topological order")
			} else {
				s.Assert().ElementsMatch(tt.expected, result)
			}
		})
	}
}

func (s *ProviderRepositoryTestSuite) TestSort_PanicsOnCycle() {
	testCases := []struct {
		name                     string
		providers                []foundation.ServiceProvider
		expectedProvidersInCycle []string
	}{
		{
			name: "ComplexCycle A->B->C->A",
			providers: []foundation.ServiceProvider{
				&ComplexProviderA{},
				&ComplexProviderB{},
				&ComplexProviderC{},
			},
			expectedProvidersInCycle: []string{
				"*foundation.ComplexProviderA",
				"*foundation.ComplexProviderB",
				"*foundation.ComplexProviderC",
			},
		},
		{
			name: "CircularBinding A->B, B->A",
			providers: []foundation.ServiceProvider{
				&CircularBindingAProvider{},
				&CircularBindingBProvider{},
			},
			expectedProvidersInCycle: []string{
				"*foundation.CircularBindingAProvider",
				"*foundation.CircularBindingBProvider",
			},
		},
		{
			name: "EmptyBindingsCircular A->B, B->A",
			providers: []foundation.ServiceProvider{
				&EmptyBindingsCircularAProvider{},
				&EmptyBindingsCircularBProvider{},
			},
			// This test will fail if virtual bindings aren't named deterministically,
			// but the panic should still occur.
			expectedProvidersInCycle: []string{
				"*foundation.EmptyBindingsCircularAProvider",
				"*foundation.EmptyBindingsCircularBProvider",
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			var recovered any
			defer func() {
				recovered = recover()
				s.Assert().NotNil(recovered, "Expected panic but none occurred")

				err, ok := recovered.(error)
				s.Assert().True(ok, "Expected panic to be an error")
				s.Assert().ErrorContains(err, "circular dependency detected")

				for _, providerName := range tt.expectedProvidersInCycle {
					s.Assert().ErrorContains(err, providerName)
				}
			}()

			s.repository.sort(tt.providers)
		})
	}
}

func (s *ProviderRepositoryTestSuite) TestDetectCycle() {
	testCases := []struct {
		name              string
		graph             map[string][]string
		bindingToProvider map[string]foundation.ServiceProvider
		expected          []string
	}{
		{
			name: "SimpleCycle",
			graph: map[string][]string{
				"A": {"B"},
				"B": {"A"},
			},
			bindingToProvider: map[string]foundation.ServiceProvider{
				"A": &MockProviderA{},
				"B": &MockProviderB{},
			},
			expected: []string{"*foundation.MockProviderA", "*foundation.MockProviderB", "*foundation.MockProviderA"},
		},
		{
			name: "ComplexCycle",
			graph: map[string][]string{
				"A": {"B"},
				"B": {"C"},
				"C": {"A"},
			},
			bindingToProvider: map[string]foundation.ServiceProvider{
				"A": &MockProviderA{},
				"B": &MockProviderB{},
				"C": &MockProviderC{},
			},
			expected: []string{"*foundation.MockProviderA", "*foundation.MockProviderB", "*foundation.MockProviderC", "*foundation.MockProviderA"},
		},
		{
			name:              "SelfLoop",
			graph:             map[string][]string{"A": {"A"}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}},
			expected:          []string{"*foundation.MockProviderA", "*foundation.MockProviderA"},
		},
		{
			name:              "NoCycle",
			graph:             map[string][]string{"A": {"B"}, "B": {"C"}, "C": {}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}, "B": &MockProviderB{}, "C": &MockProviderC{}},
			expected:          nil,
		},
		{
			name:              "DisconnectedComponents",
			graph:             map[string][]string{"A": {"B"}, "B": {"A"}, "C": {"D"}, "D": {}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}, "B": &MockProviderB{}, "C": &MockProviderC{}, "D": &MockProviderD{}},
			expected:          []string{"*foundation.MockProviderA", "*foundation.MockProviderB", "*foundation.MockProviderA"},
		},
		{
			name:              "EmptyGraph",
			graph:             map[string][]string{},
			bindingToProvider: map[string]foundation.ServiceProvider{},
			expected:          nil,
		},
		{
			name:              "SingleNode",
			graph:             map[string][]string{"A": {}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}},
			expected:          nil,
		},
		{
			name:              "MultipleCycles",
			graph:             map[string][]string{"A": {"B"}, "B": {"A"}, "C": {"D"}, "D": {"C"}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}, "B": &MockProviderB{}, "C": &MockProviderC{}, "D": &MockProviderD{}},
			expected:          []string{"*foundation.MockProviderA", "*foundation.MockProviderB", "*foundation.MockProviderA"},
		},
		{
			name:              "ComplexPath",
			graph:             map[string][]string{"A": {"B"}, "B": {"C"}, "C": {"D"}, "D": {"B"}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}, "B": &MockProviderB{}, "C": &MockProviderC{}, "D": &MockProviderD{}},
			expected:          []string{"*foundation.MockProviderB", "*foundation.MockProviderC", "*foundation.MockProviderD", "*foundation.MockProviderB"},
		},
		{
			name:              "IsolatedNodes",
			graph:             map[string][]string{"A": {"B"}, "B": {"A"}, "C": {}, "D": {}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}, "B": &MockProviderB{}, "C": &MockProviderC{}, "D": &MockProviderD{}},
			expected:          []string{"*foundation.MockProviderA", "*foundation.MockProviderB", "*foundation.MockProviderA"},
		},
		{
			name:              "LongCycle",
			graph:             map[string][]string{"A": {"B"}, "B": {"C"}, "C": {"D"}, "D": {"E"}, "E": {"A"}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}, "B": &MockProviderB{}, "C": &MockProviderC{}, "D": &MockProviderD{}, "E": &MockProviderE{}},
			expected:          []string{"*foundation.MockProviderA", "*foundation.MockProviderB", "*foundation.MockProviderC", "*foundation.MockProviderD", "*foundation.MockProviderE", "*foundation.MockProviderA"},
		},
		{
			name:              "MissingProviderMapping",
			graph:             map[string][]string{"A": {"B"}, "B": {"A"}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}}, // B missing
			expected:          []string{"*foundation.MockProviderA"},
		},
		{
			name:              "DuplicateProviderNames",
			graph:             map[string][]string{"A1": {"B"}, "A2": {"C"}, "B": {"A1"}, "C": {"A2"}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A1": &MockProviderA{}, "A2": &MockProviderA{}, "B": &MockProviderB{}, "C": &MockProviderC{}},
			expected:          []string{"*foundation.MockProviderA", "*foundation.MockProviderB", "*foundation.MockProviderA"},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result := s.repository.detectCycle(tc.graph, tc.bindingToProvider)
			s.Assert().Equal(tc.expected, result)
		})
	}
}

func isTopologicalOrder(providers []foundation.ServiceProvider, sorted []foundation.ServiceProvider) bool {
	providerIndex := make(map[foundation.ServiceProvider]int)
	for i, p := range sorted {
		providerIndex[p] = i
	}

	getDependencies := func(provider foundation.ServiceProvider) []string {
		if p, ok := provider.(interface{ Dependencies() []string }); ok {
			return p.Dependencies()
		}
		return []string{}
	}

	getBindings := func(provider foundation.ServiceProvider) []string {
		if p, ok := provider.(interface{ Relationship() binding.Relationship }); ok {
			return p.Relationship().Bindings
		}
		return []string{}
	}

	getProvideFor := func(provider foundation.ServiceProvider) []string {
		if p, ok := provider.(interface{ Relationship() binding.Relationship }); ok {
			return p.Relationship().ProvideFor
		}
		return []string{}
	}

	// Build binding to provider mapping
	bindingToProvider := make(map[string]foundation.ServiceProvider)
	for _, p := range providers {
		for _, b := range getBindings(p) {
			bindingToProvider[b] = p
		}
	}

	// Build provideFor to provider mapping (reverse relationship)
	provideForToProvider := make(map[string]foundation.ServiceProvider)
	for _, p := range providers {
		for _, pf := range getProvideFor(p) {
			provideForToProvider[pf] = p
		}
	}

	// Check all dependency relationships
	for _, p := range providers {
		// Check explicit dependencies (this provider depends on others)
		for _, dep := range getDependencies(p) {
			if depProvider, ok := bindingToProvider[dep]; ok {
				if providerIndex[depProvider] > providerIndex[p] {
					return false
				}
			}
		}

		// Check implicit dependencies through ProvideFor (others depend on this provider)
		for _, pf := range getProvideFor(p) {
			if dependentProvider, ok := bindingToProvider[pf]; ok {
				if providerIndex[p] > providerIndex[dependentProvider] {
					return false
				}
			}
		}
	}

	return true
}

type AServiceProvider struct{}

func (r *AServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"A"},
		Dependencies: []string{},
		ProvideFor:   []string{},
	}
}
func (r *AServiceProvider) Register(_ foundation.Application) {}
func (r *AServiceProvider) Boot(_ foundation.Application)     {}

type BServiceProvider struct{}

func (r *BServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"B"},
		Dependencies: []string{"A"},
		ProvideFor:   []string{"C"},
	}
}
func (r *BServiceProvider) Register(_ foundation.Application) {}
func (r *BServiceProvider) Boot(_ foundation.Application)     {}

type CServiceProvider struct{}

func (r *CServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"C"},
		Dependencies: []string{"A"},
		ProvideFor:   []string{},
	}
}
func (r *CServiceProvider) Register(_ foundation.Application) {}
func (r *CServiceProvider) Boot(_ foundation.Application)     {}

type BasicServiceProvider struct{}

func (r *BasicServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"Basic"},
		Dependencies: []string{},
		ProvideFor:   []string{},
	}
}
func (r *BasicServiceProvider) Register(_ foundation.Application) {}
func (r *BasicServiceProvider) Boot(_ foundation.Application)     {}

type ProvideForBServiceProvider struct{}

func (r *ProvideForBServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"ProvideForB"},
		Dependencies: []string{"ProvideForA"},
		ProvideFor:   []string{},
	}
}
func (r *ProvideForBServiceProvider) Register(_ foundation.Application) {}
func (r *ProvideForBServiceProvider) Boot(_ foundation.Application)     {}

type ProvideForAServiceProvider struct{}

func (r *ProvideForAServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"ProvideForA"},
		Dependencies: []string{},
		ProvideFor:   []string{"ProvideForB"},
	}
}
func (r *ProvideForAServiceProvider) Register(_ foundation.Application) {}
func (r *ProvideForAServiceProvider) Boot(_ foundation.Application)     {}

type MockProviderA struct{}

func (p *MockProviderA) Register(_ foundation.Application) {}
func (p *MockProviderA) Boot(_ foundation.Application)     {}
func (p *MockProviderA) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"provider_a"},
		Dependencies: []string{"provider_b"},
		ProvideFor:   []string{},
	}
}

type MockProviderB struct{}

func (p *MockProviderB) Register(_ foundation.Application) {}
func (p *MockProviderB) Boot(_ foundation.Application)     {}
func (p *MockProviderB) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"provider_b"},
		Dependencies: []string{"provider_a"},
		ProvideFor:   []string{},
	}
}

type MockProviderC struct{}

func (p *MockProviderC) Register(_ foundation.Application) {}
func (p *MockProviderC) Boot(_ foundation.Application)     {}
func (p *MockProviderC) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"provider_c"},
		Dependencies: []string{"provider_d"},
		ProvideFor:   []string{},
	}
}

type MockProviderD struct{}

func (p *MockProviderD) Register(_ foundation.Application) {}
func (p *MockProviderD) Boot(_ foundation.Application)     {}
func (p *MockProviderD) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"provider_d"},
		Dependencies: []string{"provider_c"},
		ProvideFor:   []string{},
	}
}

type MockProviderE struct{}

func (p *MockProviderE) Register(_ foundation.Application) {}
func (p *MockProviderE) Boot(_ foundation.Application)     {}
func (p *MockProviderE) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"provider_e"},
		Dependencies: []string{},
		ProvideFor:   []string{},
	}
}

type ComplexProviderA struct{}

func (p *ComplexProviderA) Register(_ foundation.Application) {}
func (p *ComplexProviderA) Boot(_ foundation.Application)     {}
func (p *ComplexProviderA) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"complex_a"},
		Dependencies: []string{"complex_b"},
		ProvideFor:   []string{},
	}
}

type ComplexProviderB struct{}

func (p *ComplexProviderB) Register(_ foundation.Application) {}
func (p *ComplexProviderB) Boot(_ foundation.Application)     {}
func (p *ComplexProviderB) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"complex_b"},
		Dependencies: []string{"complex_c"},
		ProvideFor:   []string{},
	}
}

type ComplexProviderC struct{}

func (p *ComplexProviderC) Register(_ foundation.Application) {}
func (p *ComplexProviderC) Boot(_ foundation.Application)     {}
func (p *ComplexProviderC) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"complex_c"},
		Dependencies: []string{"complex_a"},
		ProvideFor:   []string{},
	}
}

type EmptyDependenciesProvider struct{}

func (p *EmptyDependenciesProvider) Register(_ foundation.Application) {}
func (p *EmptyDependenciesProvider) Boot(_ foundation.Application)     {}
func (p *EmptyDependenciesProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"empty_deps"},
		Dependencies: []string{},
		ProvideFor:   []string{"provider_c"},
	}
}

type EmptyProvideForProvider struct{}

func (p *EmptyProvideForProvider) Register(_ foundation.Application) {}
func (p *EmptyProvideForProvider) Boot(_ foundation.Application)     {}
func (p *EmptyProvideForProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"empty_provide"},
		Dependencies: []string{"provider_a"},
		ProvideFor:   []string{},
	}
}

type AllEmptyProvider struct{}

func (p *AllEmptyProvider) Register(_ foundation.Application) {}
func (p *AllEmptyProvider) Boot(_ foundation.Application)     {}
func (p *AllEmptyProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{},
		ProvideFor:   []string{},
	}
}

type EmptyBindingsWithDependenciesProvider struct{}

func (p *EmptyBindingsWithDependenciesProvider) Register(_ foundation.Application) {}
func (p *EmptyBindingsWithDependenciesProvider) Boot(_ foundation.Application)     {}
func (p *EmptyBindingsWithDependenciesProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{"provider_c"},
		ProvideFor:   []string{},
	}
}

type EmptyBindingsWithProvideForProvider struct{}

func (p *EmptyBindingsWithProvideForProvider) Register(_ foundation.Application) {}
func (p *EmptyBindingsWithProvideForProvider) Boot(_ foundation.Application)     {}
func (p *EmptyBindingsWithProvideForProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{},
		ProvideFor:   []string{"provider_a"},
	}
}

type EmptyBindingsWithBothProvider struct{}

func (p *EmptyBindingsWithBothProvider) Register(_ foundation.Application) {}
func (p *EmptyBindingsWithBothProvider) Boot(_ foundation.Application)     {}
func (p *EmptyBindingsWithBothProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{"provider_e"},
		ProvideFor:   []string{"provider_c"},
	}
}

type EmptyBindingsCircularAProvider struct{}

func (p *EmptyBindingsCircularAProvider) Register(_ foundation.Application) {}
func (p *EmptyBindingsCircularAProvider) Boot(_ foundation.Application)     {}
func (p *EmptyBindingsCircularAProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{"__virtual_1"},
		ProvideFor:   []string{},
	}
}

type EmptyBindingsCircularBProvider struct{}

func (p *EmptyBindingsCircularBProvider) Register(_ foundation.Application) {}
func (p *EmptyBindingsCircularBProvider) Boot(_ foundation.Application)     {}
func (p *EmptyBindingsCircularBProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{"__virtual_0"},
		ProvideFor:   []string{},
	}
}

type CircularBindingAProvider struct{}

func (p *CircularBindingAProvider) Register(_ foundation.Application) {}
func (p *CircularBindingAProvider) Boot(_ foundation.Application)     {}
func (p *CircularBindingAProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"circular_binding_a"},
		Dependencies: []string{"circular_binding_b"},
		ProvideFor:   []string{},
	}
}

type CircularBindingBProvider struct{}

func (p *CircularBindingBProvider) Register(_ foundation.Application) {}
func (p *CircularBindingBProvider) Boot(_ foundation.Application)     {}
func (p *CircularBindingBProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"circular_binding_b"},
		Dependencies: []string{"circular_binding_a"},
		ProvideFor:   []string{},
	}
}
