package inflector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchCase(t *testing.T) {
	tests := []struct {
		name        string
		word        string
		comparison  string
		expected    string
		description string
	}{
		{"empty word", "", "ANYTHING", "", "empty word should remain empty"},
		{"empty comparison", "word", "", "word", "empty comparison should not change word"},
		{"lowercase to uppercase", "word", "COMP", "WORD", "should convert to uppercase"},
		{"uppercase to lowercase", "WORD", "comp", "word", "should convert to lowercase"},
		{"mixed to uppercase", "WoRd", "COMP", "WORD", "should convert to uppercase"},
		{"mixed to lowercase", "WoRd", "comp", "word", "should convert to lowercase"},
		{"lowercase to title case", "word", "Comp", "Word", "should convert to title case"},
		{"uppercase to title case", "WORD", "Comp", "Word", "should convert to title case"},
		{"title case to lowercase", "Word", "comp", "word", "should convert to lowercase"},
		{"title case to uppercase", "Word", "COMP", "WORD", "should convert to uppercase"},
		{"title case to title case", "Word", "Comp", "Word", "should remain title case"},
		{"longer word to title case", "complicated", "Title", "Complicated", "should convert longer word to title case"},
		{"with numbers and symbols", "word123!@#", "COMP", "WORD123!@#", "should preserve numbers and symbols"},
		{"with leading numbers", "123word", "COMP", "123WORD", "should preserve leading numbers"},
		{"with leading symbols", "!@#word", "Comp", "!@#Word", "should preserve leading symbols"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := MatchCase(test.word, test.comparison)
			assert.Equal(t, test.expected, result, "Input: %s, Comparison: %s - %s", test.word, test.comparison, test.description)
		})
	}
}

func TestMatchCaseWithSpecialPatterns(t *testing.T) {
	tests := []struct {
		name        string
		word        string
		comparison  string
		expected    string
		description string
	}{
		{"unicode characters", "café", "UPPER", "CAFÉ", "should handle unicode characters"},
		{"mixed unicode", "café", "Title", "Café", "should handle mixed unicode in title case"},
		{"empty word with title case", "", "Title", "", "empty word should remain empty regardless of pattern"},
		{"single character to title", "x", "Title", "X", "single character should be capitalized for title case"},
		{"single character to upper", "x", "UPPER", "X", "single character should be capitalized for upper case"},
		{"single character to lower", "X", "lower", "x", "single character should be lowercase for lower case"},
		{"non-letter characters only", "123!@#", "UPPER", "123!@#", "non-letter characters should remain unchanged"},
		{"non-letter characters only to title", "123!@#", "Title", "123!@#", "non-letter characters should remain unchanged for title case"},
		{"ucwords pattern", "hello world", "Hello World", "Hello World", "should convert to ucwords pattern"},
		{"ucwords with symbols", "hello-world", "Hello-World", "Hello-World", "should handle ucwords with symbols"},
		{"ucwords with multiple words", "hello brave new world", "This Is A Test", "Hello Brave New World", "should handle multiple words in ucwords pattern"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := MatchCase(test.word, test.comparison)
			assert.Equal(t, test.expected, result, "Input: %s, Comparison: %s - %s", test.word, test.comparison, test.description)
		})
	}
}

func TestUcFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"word", "Word"},
		{"Word", "Word"},
		{"wORD", "Word"},
		{"123word", "123Word"},
		{"!@#word", "!@#Word"},
		{"", ""},
		{"a", "A"},
		{"A", "A"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := UcFirst(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestUcWords(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "Hello World"},
		{"HELLO WORLD", "Hello World"},
		{"hElLo WoRLd", "Hello World"},
		{"hello-world", "Hello-World"},
		{"hello_world", "Hello_World"},
		{"hello123 world456", "Hello123 World456"},
		{"", ""},
		{"a b c", "A B C"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := UcWords(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}
