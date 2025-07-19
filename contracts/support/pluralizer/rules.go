package pluralizer

type Ruleset interface {
	Regular() Transformations
	Uninflected() Patterns
	Irregular() Substitutions
}
