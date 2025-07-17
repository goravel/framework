package pluralizer

import "github.com/goravel/framework/contracts/support/pluralizer"

var _ pluralizer.Ruleset = (*Ruleset)(nil)

type Ruleset struct {
	regular     pluralizer.Transformations
	uninflected pluralizer.Patterns
	irregular   pluralizer.Substitutions
}

func NewRuleset(regular pluralizer.Transformations, uninflected pluralizer.Patterns, irregular pluralizer.Substitutions) *Ruleset {
	return &Ruleset{
		regular:     regular,
		uninflected: uninflected,
		irregular:   irregular,
	}
}

func (r *Ruleset) Regular() pluralizer.Transformations {
	return r.regular
}

func (r *Ruleset) Uninflected() pluralizer.Patterns {
	return r.uninflected
}

func (r *Ruleset) Irregular() pluralizer.Substitutions {
	return r.irregular
}

func (r *Ruleset) IsUncountable(word string) bool {
	for _, pattern := range r.uninflected {
		if pattern.Matches(word) {
			return true
		}
	}
	return false
}
