package foundation

import (
	"testing"

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

func (s *ProviderRepositoryTestSuite) TestAdd() {
	providerA := &AServiceProvider{}
	providerB := &BServiceProvider{}
	providers := []foundation.ServiceProvider{providerA, providerB}

	s.repository.Add(providers)

	s.Equal(providers, s.repository.providers)
	s.Len(s.repository.states, 2)
	s.Contains(s.repository.states, s.repository.getProviderName(providerA))
	s.Contains(s.repository.states, s.repository.getProviderName(providerB))
	s.False(s.repository.sortedValid)
}

func (s *ProviderRepositoryTestSuite) TestAdd_Duplicates() {
	providerA := &AServiceProvider{}

	s.repository.Add([]foundation.ServiceProvider{providerA})
	s.repository.Add([]foundation.ServiceProvider{providerA})

	s.Len(s.repository.providers, 1, "Provider should not be added twice")
	s.Len(s.repository.states, 1, "State should not be added twice")
}

func (s *ProviderRepositoryTestSuite) TestBoot() {
	mockProvider := mocksfoundation.NewServiceProvider(s.T())
	s.repository.Add([]foundation.ServiceProvider{mockProvider})

	s.repository.Boot(s.mockApp)
	mockProvider.AssertNotCalled(s.T(), "Boot", s.mockApp)

	mockProvider.EXPECT().Register(s.mockApp).Return().Once()
	s.repository.Register(s.mockApp)

	mockProvider.EXPECT().Boot(s.mockApp).Return().Once()
	s.repository.Boot(s.mockApp)

	s.True(s.repository.states[s.repository.getProviderName(mockProvider)].booted)
}

func (s *ProviderRepositoryTestSuite) TestBoot_Idempotency() {
	mockProvider := mocksfoundation.NewServiceProvider(s.T())
	s.repository.Add([]foundation.ServiceProvider{mockProvider})

	mockProvider.EXPECT().Register(s.mockApp).Return().Once()
	s.repository.Register(s.mockApp)

	mockProvider.EXPECT().Boot(s.mockApp).Return().Once()

	s.repository.Boot(s.mockApp)
	s.repository.Boot(s.mockApp)

	s.True(s.repository.states[s.repository.getProviderName(mockProvider)].booted)
	mockProvider.AssertExpectations(s.T())
}

func (s *ProviderRepositoryTestSuite) TestGetBooted() {
	providerA := &AServiceProvider{}
	providerB := &BServiceProvider{}

	s.repository.Add([]foundation.ServiceProvider{providerA, providerB})

	keyA := s.repository.getProviderName(providerA)
	keyB := s.repository.getProviderName(providerB)

	s.NotEqual(keyA, keyB)

	s.repository.states[keyA].registered = true
	s.repository.states[keyB].registered = false

	s.repository.Boot(s.mockApp)

	booted := s.repository.GetBooted()

	s.Len(booted, 1, "Expected only one booted provider")
	s.Equal(providerA, booted[0], "The booted provider should be providerA")
	s.NotContains(booted, providerB, "Booted list should not contain providerB")

	s.True(s.repository.states[keyA].booted)
	s.False(s.repository.states[keyB].booted)
}

func (s *ProviderRepositoryTestSuite) TestLoadFromConfig_Success() {
	providers := []foundation.ServiceProvider{&BServiceProvider{}, &AServiceProvider{}}
	mockConfig := mocksconfig.NewConfig(s.T())

	s.mockApp.EXPECT().MakeConfig().Return(mockConfig).Once()
	mockConfig.EXPECT().Get("app.providers").Return(providers).Once()

	result := s.repository.LoadFromConfig(s.mockApp)

	s.Equal(providers, result)
	s.Equal(providers, s.repository.providers)
	s.True(s.repository.loaded)
	s.False(s.repository.sortedValid)
}

func (s *ProviderRepositoryTestSuite) TestLoadFromConfig_AlreadyLoaded() {
	providers := []foundation.ServiceProvider{&AServiceProvider{}}
	s.repository.loaded = true
	s.repository.providers = providers

	result := s.repository.LoadFromConfig(s.mockApp)

	s.Equal(providers, result)
	s.mockApp.AssertNotCalled(s.T(), "MakeConfig")
}

func (s *ProviderRepositoryTestSuite) TestLoadFromConfig_NilConfig() {
	s.mockApp.EXPECT().MakeConfig().Return(nil).Once()

	result := s.repository.LoadFromConfig(s.mockApp)

	s.Empty(result)
	s.NotNil(result, "Should return empty slice, not nil")
	s.False(s.repository.loaded)
}

func (s *ProviderRepositoryTestSuite) TestLoadFromConfig_BadConfigType() {
	mockConfig := mocksconfig.NewConfig(s.T())
	s.mockApp.EXPECT().MakeConfig().Return(mockConfig).Once()
	mockConfig.EXPECT().Get("app.providers").Return("not a slice").Once()

	result := s.repository.LoadFromConfig(s.mockApp)

	s.Empty(result)
	s.NotNil(result, "Should return empty slice, not nil")
	s.False(s.repository.loaded)
}

func (s *ProviderRepositoryTestSuite) TestRegister() {
	mockProvider := mocksfoundation.NewServiceProvider(s.T())
	s.repository.Add([]foundation.ServiceProvider{mockProvider})

	mockProvider.EXPECT().Register(s.mockApp).Return().Once()

	processed := s.repository.Register(s.mockApp)

	s.Equal([]foundation.ServiceProvider{mockProvider}, processed)
	s.True(s.repository.states[s.repository.getProviderName(mockProvider)].registered)
	s.False(s.repository.states[s.repository.getProviderName(mockProvider)].booted)
}

func (s *ProviderRepositoryTestSuite) TestRegister_Idempotency() {
	mockProvider := mocksfoundation.NewServiceProvider(s.T())
	s.repository.Add([]foundation.ServiceProvider{mockProvider})

	mockProvider.EXPECT().Register(s.mockApp).Return().Once()

	s.repository.Register(s.mockApp)
	s.repository.Register(s.mockApp)

	s.True(s.repository.states[s.repository.getProviderName(mockProvider)].registered)
	mockProvider.AssertExpectations(s.T())
}

func (s *ProviderRepositoryTestSuite) TestReset() {
	s.repository.providers = []foundation.ServiceProvider{&AServiceProvider{}}
	s.repository.states["foo"] = &ProviderState{}
	s.repository.sorted = []foundation.ServiceProvider{}
	s.repository.sortedValid = true
	s.repository.loaded = true

	s.repository.Reset()

	s.Empty(s.repository.providers)
	s.Empty(s.repository.states)
	s.Empty(s.repository.sorted)
	s.False(s.repository.sortedValid)
	s.False(s.repository.loaded)
}

func (s *ProviderRepositoryTestSuite) TestGetProviders_SortingAndCaching() {
	providerA := &AServiceProvider{}
	providerB := &BServiceProvider{}
	providers := []foundation.ServiceProvider{providerB, providerA}
	expectedSorted := []foundation.ServiceProvider{providerA, providerB}

	s.repository.Add(providers)
	s.False(s.repository.sortedValid, "Cache should be invalid initially")

	sorted1 := s.repository.getProviders()
	s.Equal(expectedSorted, sorted1)
	s.True(s.repository.sortedValid, "Cache should be valid after first call")
	s.Same(&s.repository.sorted[0], &sorted1[0], "Internal cache should be set")

	sorted2 := s.repository.getProviders()
	s.Equal(expectedSorted, sorted2)
	s.Same(&sorted1[0], &sorted2[0], "Should return exact same slice from cache")
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
