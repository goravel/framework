package schema

type IndexDefinition interface {
	Algorithm(algorithm string) IndexDefinition
	Name(name string) IndexDefinition
}
