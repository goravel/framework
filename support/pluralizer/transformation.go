package pluralizer

import (
	"github.com/goravel/framework/contracts/support/pluralizer"
	"regexp"
)

type Transformation struct {
	pattern     *regexp.Regexp
	replacement string
}

var _ pluralizer.Transformation = (*Transformation)(nil)

func NewTransformation(pattern, replacement string) *Transformation {
	return &Transformation{
		pattern:     regexp.MustCompile(pattern),
		replacement: replacement,
	}
}

func (r *Transformation) Apply(word string) string {
	if r.pattern.MatchString(word) {
		return r.pattern.ReplaceAllString(word, r.replacement)
	}
	return word
}
