package rotation

import (
	"errors"
	"strings"
	"time"
)

// Strftime provides strftime-like formatting for filenames
type Strftime struct {
	pattern string
}

// NewStrftime creates a new Strftime formatter
func NewStrftime(pattern string) (*Strftime, error) {
	if pattern == "" {
		return nil, errors.New("pattern cannot be empty")
	}
	return &Strftime{pattern: pattern}, nil
}

// Format formats the time according to the pattern
func (s *Strftime) Format(t time.Time) string {
	result := s.pattern

	// Common strftime patterns used by the framework
	replacements := map[string]string{
		"%Y": t.Format("2006"),      // Year with century
		"%m": t.Format("01"),         // Month (01-12)
		"%d": t.Format("02"),         // Day of month (01-31)
		"%H": t.Format("15"),         // Hour (00-23)
		"%M": t.Format("04"),         // Minute (00-59)
		"%S": t.Format("05"),         // Second (00-59)
		"%%": "%",                    // Literal %
	}

	for pattern, replacement := range replacements {
		result = strings.ReplaceAll(result, pattern, replacement)
	}

	return result
}
