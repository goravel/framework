package english

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnglishRulesetStructure(t *testing.T) {
	lang := New()

	assert.Equal(t, "english", lang.Name())
	assert.NotNil(t, lang.PluralRuleset())
	assert.NotNil(t, lang.SingularRuleset())

	pluralRules := lang.PluralRuleset()
	singularRules := lang.SingularRuleset()

	assert.True(t, len(pluralRules.Irregular()) > 0)
	assert.True(t, len(singularRules.Irregular()) > 0)
	assert.True(t, len(pluralRules.Regular()) > 0)
	assert.True(t, len(singularRules.Regular()) > 0)
	assert.True(t, len(pluralRules.Uninflected()) > 0)
	assert.True(t, len(singularRules.Uninflected()) > 0)
}
