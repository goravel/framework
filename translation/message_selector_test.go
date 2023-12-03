package translation

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MessageSelectorTestSuite struct {
	suite.Suite
	selector *MessageSelector
}

func TestMessageSelectorTestSuite(t *testing.T) {
	suite.Run(t, new(MessageSelectorTestSuite))
}

func (m *MessageSelectorTestSuite) SetUpTest() {
	m.selector = NewMessageSelector()
}

func (m *MessageSelectorTestSuite) TestChoose() {
	tests := []struct {
		expected string
		message  string
		number   int
		locale   string
	}{
		{"first", "first", 0, "en"},
		{"first", "first", 10, "en"},
		{"first", "first|second", 1, "en"},
		{"second", "first|second", 10, "en"},
		{"second", "first|second", 0, "en"},

		{"first", "{0} first|{1}second", 0, "en"},
		{"first", "{1}first|{2}second", 1, "en"},
		{"second", "{1}first|{2}second", 2, "en"},
		{"first", "{2}first|{1}second", 2, "en"},
		{"second", "{9}first|{10}second", 0, "en"},
		{"first", "{9}first|{10}second", 1, "en"},
		{"", "{0}|{1}second", 0, "en"}, // there is some problem with it
		{"", "{0}first|{1}", 1, "en"},  // same
		{"first\nline", "{1}first\nline|{2}second", 1, "en"},
		{"first \nline", "{1}first \nline|{2}second", 1, "en"},
		{"first", "{0}first|[1,9]second", 0, "en"},
		{"second", "{0}first|[1,9]second", 1, "en"},
		{"second", "{0}first|[1,9]second", 10, "en"},
		{"first", "{0}first|[2,9]second", 1, "en"},
		{"second", "[4,*]first|[1,3]second", 1, "en"},
		{"first", "[4,*]first|[1,3]second", 100, "en"},
		{"second", "[1,5]first|[6,10]second", 7, "en"},
		{"first", "[*,4]first|[5,*]second", 1, "en"},
		{"second", "[5,*]first|[*,4]second", 1, "en"},
		{"second", "[5,*]first|[*,4]second", 0, "en"},

		{"first", "{0}first|[1,3]second|[4,*]third", 0, "en"},
		{"second", "{0}first|[1,3]second|[4,*]third", 1, "en"},
		{"third", "{0}first|[1,3]second|[4,*]third", 9, "en"},

		{"first", "first|second|third", 1, "en"},
		{"second", "first|second|third", 9, "en"},
		{"second", "first|second|third", 0, "en"},

		{"first", "{0} first | { 1 } second", 0, "en"},
		{"first", "[4,*]first | [1,3] second", 100, "en"},
	}

	for _, test := range tests {
		m.Equal(test.expected, m.selector.Choose(test.message, test.number, test.locale))
	}
}

func (m *MessageSelectorTestSuite) TestExtract() {
	tests := []struct {
		segments []string
		number   int
		expected *string
	}{
		{[]string{"{0} first", "{1}second"}, 0, stringPtr(" first")},
		{[]string{"{1}first", "{2}second"}, 0, nil},
		{[]string{"{0}first", "{1}second"}, 0, stringPtr("first")},
		{[]string{"[4,*]first", "[1,3]second"}, 100, stringPtr("first")},
	}
	for _, test := range tests {
		value := m.selector.extract(test.segments, test.number)
		if value == nil {
			m.Equal(test.expected, value)
			continue
		}
		m.Equal(*test.expected, *value)
	}

}

func (m *MessageSelectorTestSuite) TestExtractFromString() {
	var tests = []struct {
		segment  string
		number   int
		expected *string
	}{
		{"{0}first", 0, stringPtr("first")},
		{"[4,*]first", 5, stringPtr("first")},
		{"[1,3]second", 0, nil},
		{"[*,4]second", 3, stringPtr("second")},
		{"[*,*]second", 0, stringPtr("second")},
	}

	for _, test := range tests {
		value := m.selector.extractFromString(test.segment, test.number)
		if value == nil {
			m.Equal(test.expected, value)
			continue
		}
		m.Equal(*test.expected, *value)
	}
}

func (m *MessageSelectorTestSuite) TestStripConditions() {
	tests := []struct {
		segments []string
		expected []string
	}{
		{[]string{"{0}first", "[2,9]second"}, []string{"first", "second"}},
		{[]string{"[4,*]first", "[1,3]second"}, []string{"first", "second"}},
		{[]string{"first", "second"}, []string{"first", "second"}},
	}

	for _, test := range tests {
		m.Equal(test.expected, stripConditions(test.segments))
	}
}

func (m *MessageSelectorTestSuite) TestGetPluralIndex() {
	tests := []struct {
		locale   string
		number   int
		expected int
	}{
		{"az", 0, 0},
		{"af", 1, 0},
		{"af", 10, 1},
		{"am", 0, 0},
		{"am", 1, 0},
		{"am", 10, 1},
		{"be", 1, 0},
		{"be", 3, 1},
		{"be", 23, 1},
		{"be", 5, 2},
		{"cs", 1, 0},
		{"cs", 3, 1},
		{"cs", 10, 2},
		{"ga", 1, 0},
		{"ga", 2, 1},
		{"ga", 5, 2},
		{"lt", 1, 0},
		{"lt", 3, 1},
		{"lt", 3, 1},
		{"lt", 10, 2},
		{"sl", 1, 0},
		{"sl", 2, 1},
		{"sl", 3, 2},
		{"sl", 4, 2},
		{"sl", 5, 3},
		{"sl", 10, 3},
		{"mk", 1, 0},
		{"mk", 2, 1},
		{"mt", 1, 0},
		{"mt", 0, 1},
		{"mt", 2, 1},
		{"mt", 11, 2},
		{"mt", 20, 3},
		{"lv", 0, 0},
		{"lv", 21, 1},
		{"lv", 11, 2},
		{"lv", 2, 2},
		{"pl", 1, 0},
		{"pl", 2, 1},
		{"pl", 5, 2},
		{"pl", 10, 2},
		{"cy", 1, 0},
		{"cy", 2, 1},
		{"cy", 8, 2},
		{"cy", 11, 2},
		{"cy", 20, 3},
		{"ro", 1, 0},
		{"ro", 0, 1},
		{"ro", 112, 1},
		{"ro", 21, 2},
		{"ar", 0, 0},
		{"ar", 1, 1},
		{"ar", 2, 2},
		{"ar", 4, 3},
		{"ar", 112, 4},
		{"ar", 102, 5},
		{"else", 10, 0},
	}

	for _, test := range tests {
		m.Equal(test.expected, getPluralIndex(test.number, test.locale))
	}
}

func stringPtr(s string) *string {
	return &s
}
