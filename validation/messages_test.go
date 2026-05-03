package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMessage(t *testing.T) {
	t.Run("custom field+rule message has highest priority", func(t *testing.T) {
		custom := map[string]string{
			"name.required": "Name is absolutely required!",
			"required":      "Field is required.",
		}
		msg := getMessage("name", "required", custom, "string")
		assert.Equal(t, "Name is absolutely required!", msg)
	})

	t.Run("custom rule message", func(t *testing.T) {
		custom := map[string]string{
			"required": "This field cannot be empty.",
		}
		msg := getMessage("email", "required", custom, "string")
		assert.Equal(t, "This field cannot be empty.", msg)
	})

	t.Run("type-specific default for size rules", func(t *testing.T) {
		msg := getMessage("name", "min", nil, "string")
		assert.Equal(t, "The :attribute field must be at least :min characters.", msg)

		msg = getMessage("age", "min", nil, "numeric")
		assert.Equal(t, "The :attribute field must be at least :min.", msg)

		msg = getMessage("items", "min", nil, "array")
		assert.Equal(t, "The :attribute field must have at least :min items.", msg)

		msg = getMessage("doc", "min", nil, "file")
		assert.Equal(t, "The :attribute field must be at least :min kilobytes.", msg)
	})

	t.Run("generic default message", func(t *testing.T) {
		msg := getMessage("email", "email", nil, "string")
		assert.Equal(t, "The :attribute field must be a valid email address.", msg)
	})

	t.Run("unknown rule returns fallback", func(t *testing.T) {
		msg := getMessage("field", "unknown_rule_xyz", nil, "string")
		assert.Equal(t, "The :attribute field is invalid.", msg)
	})

	t.Run("all size rules have type variants", func(t *testing.T) {
		for rule := range sizeRules {
			for _, typ := range []string{"string", "numeric", "array", "file"} {
				msg := getMessage("field", rule, nil, typ)
				assert.NotEqual(t, "The :attribute field is invalid.", msg, "missing message for %s.%s", rule, typ)
			}
		}
	})
}

func TestFormatMessage(t *testing.T) {
	t.Run("no replacements", func(t *testing.T) {
		msg := formatMessage("Hello world", nil)
		assert.Equal(t, "Hello world", msg)
	})

	t.Run("single replacement", func(t *testing.T) {
		msg := formatMessage("The :attribute field is required.", map[string]string{
			":attribute": "name",
		})
		assert.Equal(t, "The name field is required.", msg)
	})

	t.Run("multiple replacements", func(t *testing.T) {
		msg := formatMessage("The :attribute field must be between :min and :max.", map[string]string{
			":attribute": "age",
			":min":       "1",
			":max":       "100",
		})
		assert.Equal(t, "The age field must be between 1 and 100.", msg)
	})

	t.Run("longer placeholders replaced before shorter ones", func(t *testing.T) {
		msg := formatMessage(":values and :value", map[string]string{
			":value":  "X",
			":values": "A, B",
		})
		assert.Equal(t, "A, B and X", msg)
	})

	t.Run("more than 4 replacements uses sort path", func(t *testing.T) {
		msg := formatMessage(":a :b :c :d :e", map[string]string{
			":a": "1", ":b": "2", ":c": "3", ":d": "4", ":e": "5",
		})
		assert.Equal(t, "1 2 3 4 5", msg)
	})
}

func TestGetDisplayableAttribute(t *testing.T) {
	t.Run("custom attribute name", func(t *testing.T) {
		custom := map[string]string{"first_name": "First Name"}
		assert.Equal(t, "First Name", getDisplayableAttribute("first_name", custom))
	})

	t.Run("replaces underscores with spaces", func(t *testing.T) {
		assert.Equal(t, "first name", getDisplayableAttribute("first_name", nil))
	})

	t.Run("no underscores returns as-is", func(t *testing.T) {
		assert.Equal(t, "email", getDisplayableAttribute("email", nil))
	})
}
