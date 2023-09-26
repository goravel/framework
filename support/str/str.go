package str

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/exp/constraints"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type String struct {
	value string
}

// ExcerptOption is the option for Excerpt method
type ExcerptOption struct {
	Radius   int
	Omission string
}

func Of(value string) *String {
	return &String{value: value}
}

func (s *String) After(search string) *String {
	index := strings.Index(s.value, search)
	if index != -1 {
		s.value = s.value[index+len(search):]
	}

	return s
}

func (s *String) AfterLast(search string) *String {
	index := strings.LastIndex(s.value, search)
	if index != -1 {
		s.value = s.value[index+len(search):]
	}

	return s
}

func (s *String) Append(values ...string) *String {
	s.value += strings.Join(values, "")
	return s
}

func (s *String) Basename(suffix ...string) *String {
	s.value = filepath.Base(s.value)
	if len(suffix) > 0 && suffix[0] != "" {
		s.value = strings.TrimSuffix(s.value, suffix[0])
	}
	return s
}

func (s *String) Before(search string) *String {
	index := strings.Index(s.value, search)
	if index != -1 {
		s.value = s.value[:index]
	}

	return s
}

func (s *String) BeforeLast(search string) *String {
	index := strings.LastIndex(s.value, search)
	if index != -1 {
		s.value = s.value[:index]
	}

	return s
}

func (s *String) Between(start, end string) *String {
	return s.After(start).Before(end)
}

func (s *String) BetweenFirst(start, end string) *String {
	return s.Before(end).After(start)
}

func (s *String) Camel() *String {
	return s.Studly().LcFirst()
}

func (s *String) CharAt(index int) string {
	if index < 0 || index >= len(s.value) {
		return ""
	}
	return string(s.value[index])
}

func (s *String) Contains(values ...string) bool {
	for _, value := range values {
		if strings.Contains(s.value, value) {
			return true
		}
	}

	return false
}

func (s *String) ContainsAll(values ...string) bool {
	for _, value := range values {
		if !strings.Contains(s.value, value) {
			return false
		}
	}

	return true
}

func (s *String) Dirname(levels ...int) *String {
	defaultLevels := 1
	if len(levels) > 0 {
		defaultLevels = levels[0]
	}

	dir := s.value
	for i := 0; i < defaultLevels; i++ {
		dir = filepath.Dir(dir)
	}

	s.value = dir
	return s
}

func (s *String) EndsWith(values ...string) bool {
	for _, value := range values {
		if strings.HasSuffix(s.value, value) {
			return true
		}
	}

	return false
}

func (s *String) Exactly(value string) bool {
	return s.value == value
}

func (s *String) Excerpt(phrase string, options ...ExcerptOption) *String {
	defaultOptions := ExcerptOption{
		Radius:   100,
		Omission: "...",
	}

	if len(options) > 0 {
		if options[0].Radius != 0 {
			defaultOptions.Radius = options[0].Radius
		}
		if options[0].Omission != "" {
			defaultOptions.Omission = options[0].Omission
		}
	}

	radius := Max(0, defaultOptions.Radius)
	omission := defaultOptions.Omission

	regex := regexp.MustCompile(`(.*?)(` + regexp.QuoteMeta(phrase) + `)(.*)`)
	matches := regex.FindStringSubmatch(s.value)

	if len(matches) == 0 {
		return s
	}

	start := strings.TrimRight(matches[1], "")
	end := strings.TrimLeft(matches[3], "")

	end = Of(Substr(end, 0, radius)).LTrim("").
		Unless(func(s *String) bool {
			return s.Exactly(end)
		}, func(s *String) *String {
			return s.Append(omission)
		}).String()

	s.value = Of(Substr(start, Max(len(start)-radius, 0), radius)).LTrim("").
		Unless(func(s *String) bool {
			return s.Exactly(start)
		}, func(s *String) *String {
			return s.Prepend(omission)
		}).Append(matches[2], end).String()

	return s
}

func (s *String) Explode(delimiter string) []string {
	return strings.Split(s.value, delimiter)
}

func (s *String) Finish(value string) *String {
	quoted := regexp.QuoteMeta(value)
	reg := regexp.MustCompile("(?:" + quoted + ")+$")
	s.value = reg.ReplaceAllString(s.value, "") + value
	return s
}

func (s *String) Headline() *String {
	parts := s.Explode(" ")

	if len(parts) > 1 {
		return s.Studly()
	}

	parts = Of(strings.Join(parts, "_")).Studly().UcSplit()
	collapsed := Of(strings.Join(parts, "_")).
		Replace("-", "_").
		Replace(" ", "_").
		Replace("_", "_").Explode("_")

	s.value = strings.Join(collapsed, " ")
	return s
}

func (s *String) Is(patterns ...string) bool {
	for _, pattern := range patterns {
		if pattern == s.value {
			return true
		}

		// Escape special characters in the pattern
		pattern = regexp.QuoteMeta(pattern)

		// Replace asterisks with regular expression wildcards
		pattern = strings.ReplaceAll(pattern, `\*`, ".*")

		// Create a regular expression pattern for matching
		regexPattern := "^" + pattern + "$"

		// Compile the regular expression
		regex := regexp.MustCompile(regexPattern)

		// Check if the value matches the pattern
		if regex.MatchString(s.value) {
			return true
		}
	}

	return false
}

func (s *String) IsEmpty() bool {
	return s.value == ""
}

func (s *String) IsNotEmpty() bool {
	return !s.IsEmpty()
}

func (s *String) IsJson() bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s.value), &js) == nil
}

func (s *String) IsUlid() bool {
	return s.IsMatch(`^[0-9A-Z]{26}$`)
}

// func (s *String) IsUrl() bool

func (s *String) IsUuid() bool {
	return s.IsMatch(`^[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$`)
}

func (s *String) Kebab() *String {
	return s.Snake("-")
}

func (s *String) LcFirst() *String {
	if s.Length() == 0 {
		return s
	}
	s.value = strings.ToLower(Substr(s.value, 0, 1)) + Substr(s.value, 1)
	return s
}

func (s *String) Length() int {
	return len([]rune(s.value))
}

func (s *String) Limit(limit int, end ...string) *String {
	defaultEnd := "..."
	if len(end) > 0 {
		defaultEnd = end[0]
	}

	if s.Length() <= limit {
		return s
	}
	s.value = s.value[:limit] + defaultEnd
	return s
}

func (s *String) Lower() *String {
	s.value = strings.ToLower(s.value)
	return s
}

func (s *String) LTrim(characters ...string) *String {
	if len(characters) == 0 {
		s.value = strings.TrimLeft(s.value, " ")
		return s
	}

	s.value = strings.TrimLeft(s.value, characters[0])
	return s
}

func (s *String) Mask(character string, index int, length ...int) *String {
	// Check if the character is empty, if so, return the original string.
	if character == "" {
		return s
	}

	segment := Substr(s.value, index, length...)

	// Check if the segment is empty, if so, return the original string.
	if segment == "" {
		return s
	}

	strLen := utf8.RuneCountInString(s.value)
	startIndex := index

	// Check if the start index is out of bounds.
	if index < 0 {
		if index < -strLen {
			startIndex = 0
		} else {
			startIndex = strLen + index
		}
	}

	start := Substr(s.value, 0, startIndex)
	segmentLen := utf8.RuneCountInString(segment)
	end := Substr(s.value, startIndex+segmentLen)

	s.value = start + strings.Repeat(Substr(character, 0, 1), segmentLen) + end
	return s
}

func (s *String) Match(pattern string) string {
	reg := regexp.MustCompile(pattern)
	return reg.FindString(s.value)
}

func (s *String) MatchAll(pattern string) []string {
	reg := regexp.MustCompile(pattern)
	return reg.FindAllString(s.value, -1)
}

func (s *String) IsMatch(pattern string) bool {
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(s.value)
}

func (s *String) NewLine(count ...int) *String {
	if len(count) == 0 {
		s.value += "\n"
		return s
	}

	s.value += strings.Repeat("\n", count[0])
	return s
}

func (s *String) PadBoth(length int, pad ...string) *String {
	defaultPad := " "
	if len(pad) > 0 {
		defaultPad = pad[0]
	}
	short := Max(0, length-s.Length())
	left := short / 2
	right := short/2 + short%2

	s.value = Substr(strings.Repeat(defaultPad, left), 0, left) + s.value + Substr(strings.Repeat(defaultPad, right), 0, right)

	return s
}

func (s *String) PadLeft(length int, pad ...string) *String {
	defaultPad := " "
	if len(pad) > 0 {
		defaultPad = pad[0]
	}
	short := Max(0, length-s.Length())

	s.value = Substr(strings.Repeat(defaultPad, short), 0, short) + s.value
	return s
}

func (s *String) PadRight(length int, pad ...string) *String {
	defaultPad := " "
	if len(pad) > 0 {
		defaultPad = pad[0]
	}
	short := Max(0, length-s.Length())

	s.value = s.value + Substr(strings.Repeat(defaultPad, short), 0, short)
	return s
}

func (s *String) Pipe(callback func(s string) string) *String {
	s.value = callback(s.value)
	return s
}

func (s *String) Prepend(values ...string) *String {
	s.value = strings.Join(values, "") + s.value
	return s
}

func (s *String) Remove(values ...string) *String {
	for _, value := range values {
		s.value = strings.ReplaceAll(s.value, value, "")
	}

	return s
}

func (s *String) Repeat(times int) *String {
	s.value = strings.Repeat(s.value, times)
	return s
}

func (s *String) Replace(search string, replace string, caseSensitive ...bool) *String {
	caseSensitive = append(caseSensitive, true)
	if len(caseSensitive) > 0 && !caseSensitive[0] {
		s.value = regexp.MustCompile("(?i)"+search).ReplaceAllString(s.value, replace)
		return s
	}
	s.value = strings.ReplaceAll(s.value, search, replace)
	return s
}

func (s *String) ReplaceEnd(search string, replace string) *String {
	if search == "" {
		return s
	}

	if s.EndsWith(search) {
		return s.ReplaceLast(search, replace)
	}

	return s
}

func (s *String) ReplaceFirst(search string, replace string) *String {
	s.value = strings.Replace(s.value, search, replace, 1)
	return s
}

func (s *String) ReplaceLast(search string, replace string) *String {
	index := strings.LastIndex(s.value, search)
	if index != -1 {
		s.value = s.value[:index] + replace + s.value[index+len(search):]
		return s
	}

	return s
}

func (s *String) ReplaceMatches(pattern string, replace string) *String {
	s.value = regexp.MustCompile(pattern).ReplaceAllString(s.value, replace)
	return s
}

func (s *String) ReplaceStart(search string, replace string) *String {
	if search == "" {
		return s
	}

	if s.StartsWith(search) {
		return s.ReplaceFirst(search, replace)
	}

	return s
}

func (s *String) RTrim(characters ...string) *String {
	if len(characters) == 0 {
		s.value = strings.TrimRight(s.value, " ")
		return s
	}

	s.value = strings.TrimRight(s.value, characters[0])
	return s
}

func (s *String) Snake(delimiter ...string) *String {
	defaultDelimiter := '_'
	if len(delimiter) > 0 {
		defaultDelimiter = []rune(delimiter[0])[0]
	}
	value := s.Studly().String()
	var result []rune
	for i, r := range value {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, defaultDelimiter)
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	s.value = string(result)
	return s
}

func (s *String) Split(pattern string, limit ...int) []string {
	r := regexp.MustCompile(pattern)
	defaultLimit := -1
	if len(limit) != 0 {
		defaultLimit = limit[0]
	}

	return r.Split(s.value, defaultLimit)
}

func (s *String) Squish() *String {
	leadWhitespace := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	insideWhitespace := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	s.value = leadWhitespace.ReplaceAllString(s.value, "")
	s.value = insideWhitespace.ReplaceAllString(s.value, " ")
	return s
}

func (s *String) Start(prefix string) *String {
	quoted := regexp.QuoteMeta(prefix)
	re := regexp.MustCompile(`^(` + quoted + `)+`)
	s.value = prefix + re.ReplaceAllString(s.value, "")
	return s
}

func (s *String) StartsWith(values ...string) bool {
	for _, value := range values {
		if strings.HasPrefix(s.value, value) {
			return true
		}
	}

	return false
}

func (s *String) String() string {
	return s.value
}

func (s *String) Studly() *String {
	words := FieldsFunc(s.value, func(r rune) bool {
		return r == '_' || r == ' ' || r == '-' || r == ',' || r == '.'
	}, func(r rune) bool {
		return unicode.IsUpper(r)
	})

	casesTitle := cases.Title(language.Und)
	var studlyWords []string
	for _, word := range words {
		studlyWords = append(studlyWords, casesTitle.String(word))
	}

	s.value = strings.Join(studlyWords, "")
	return s
}

func (s *String) Substr(start int, length ...int) *String {
	s.value = Substr(s.value, start, length...)
	return s
}

func (s *String) Swap(replacements map[string]string) *String {
	if len(replacements) == 0 {
		return s
	}

	oldNewPairs := make([]string, 0, len(replacements)*2)
	for k, v := range replacements {
		if k == "" {
			return s
		}
		oldNewPairs = append(oldNewPairs, k, v)
	}

	s.value = strings.NewReplacer(oldNewPairs...).Replace(s.value)
	return s
}

func (s *String) Tap(callback func(String)) *String {
	callback(*s)
	return s
}

func (s *String) Test(pattern string) bool {
	return s.IsMatch(pattern)
}

func (s *String) Title() *String {
	casesTitle := cases.Title(language.Und)
	s.value = casesTitle.String(s.value)
	return s
}

func (s *String) Trim(characters ...string) *String {
	if len(characters) == 0 {
		s.value = strings.TrimSpace(s.value)
		return s
	}

	s.value = strings.Trim(s.value, characters[0])
	return s
}

func (s *String) UcFirst() *String {
	if s.Length() == 0 {
		return s
	}
	s.value = strings.ToUpper(Substr(s.value, 0, 1)) + Substr(s.value, 1)
	return s
}

func (s *String) UcSplit() []string {
	words := FieldsFunc(s.value, func(r rune) bool {
		return false
	}, func(r rune) bool {
		return unicode.IsUpper(r)
	})
	return words
}

func (s *String) Unless(callback func(*String) bool, fallback func(*String) *String) *String {
	if !callback(s) {
		fallback(s)
	}

	return s
}

func (s *String) Upper() *String {
	s.value = strings.ToUpper(s.value)
	return s
}

func (s *String) When(condition bool, callback ...func(*String) *String) *String {
	if condition {
		callback[0](s)
	} else {
		if len(callback) > 1 {
			callback[1](s)
		}
	}

	return s
}

func (s *String) WhenContains(value string, callback ...func(*String) *String) *String {
	return s.When(s.Contains(value), callback...)
}

func (s *String) WhenContainsAll(values []string, callback ...func(*String) *String) *String {
	return s.When(s.ContainsAll(values...), callback...)
}

func (s *String) WhenEmpty(callback ...func(*String) *String) *String {
	return s.When(s.IsEmpty(), callback...)
}

func (s *String) WhenNotEmpty(callback ...func(*String) *String) *String {
	return s.When(s.IsNotEmpty(), callback...)
}

func (s *String) WhenStartsWith(value string, callback ...func(*String) *String) *String {
	return s.When(s.StartsWith(value), callback...)
}

func (s *String) WhenEndsWith(value string, callback ...func(*String) *String) *String {
	return s.When(s.EndsWith(value), callback...)
}

func (s *String) WhenExactly(value string, callback ...func(*String) *String) *String {
	return s.When(s.Exactly(value), callback...)
}

func (s *String) WhenNotExactly(value string, callback ...func(*String) *String) *String {
	return s.When(!s.Exactly(value), callback...)
}

func (s *String) WhenIs(value string, callback ...func(*String) *String) *String {
	return s.When(s.Is(value), callback...)
}

func (s *String) WhenIsUlid(callback ...func(*String) *String) *String {
	return s.When(s.IsUlid(), callback...)
}

func (s *String) WhenIsUuid(callback ...func(*String) *String) *String {
	return s.When(s.IsUuid(), callback...)
}

func (s *String) WhenTest(pattern string, callback ...func(*String) *String) *String {
	return s.When(s.Test(pattern), callback...)
}

func (s *String) WordCount() int {
	return len(strings.Fields(s.value))
}

func (s *String) Words(limit int, end ...string) *String {
	defaultEnd := "..."
	if len(end) > 0 {
		defaultEnd = end[0]
	}

	words := strings.Fields(s.value)
	if len(words) <= limit {
		return s
	}

	s.value = strings.Join(words[:limit], " ") + defaultEnd
	return s
}

// FieldsFunc splits the input string into words with preservation, following the rules defined by
// the provided functions f and preserveFunc.
func FieldsFunc(s string, f func(rune) bool, preserveFunc ...func(rune) bool) []string {
	var fields []string
	var currentField strings.Builder

	shouldPreserve := func(r rune) bool {
		for _, preserveFn := range preserveFunc {
			if preserveFn(r) {
				return true
			}
		}
		return false
	}

	for _, r := range s {
		if f(r) {
			if currentField.Len() > 0 {
				fields = append(fields, currentField.String())
				currentField.Reset()
			}
		} else if shouldPreserve(r) {
			if currentField.Len() > 0 {
				fields = append(fields, currentField.String())
				currentField.Reset()
			}
			currentField.WriteRune(r)
		} else {
			currentField.WriteRune(r)
		}
	}

	if currentField.Len() > 0 {
		fields = append(fields, currentField.String())
	}

	return fields
}

// Substr returns a substring of a given string, starting at the specified index
// and with a specified length.
// It handles UTF-8 encoded strings.
func Substr(str string, start int, length ...int) string {
	// Convert the string to a rune slice for proper handling of UTF-8 encoding.
	runes := []rune(str)
	strLen := utf8.RuneCountInString(str)

	// Check if the start index is out of bounds.
	if start >= strLen {
		return ""
	}

	// If the start index is negative, count backwards from the end of the string.
	if start < 0 {
		start = strLen + start
		if start < 0 { // If start is still negative, set it to 0.
			start = 0
		}
	}

	// If the length is 0, return the substring from start to the end of the string.
	if len(length) == 0 {
		return string(runes[start:])
	}
	lenArg := length[0]
	// Calculate the end index based on the start and length.
	end := start + lenArg

	// Ensure the end index is within bounds.
	// Ensure the end index is within bounds.
	if end < 0 {
		end = start
	}
	if end > strLen {
		end = strLen
	}

	// Handle the case where lenArg is negative and less than start
	if lenArg < 0 && end < start {
		start, end = end, start
	}

	// Return the substring.
	return string(runes[start:end])
}

func Max[T constraints.Ordered](x T, y T) T {
	if x > y {
		return x
	}
	return y
}

func Random(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	letters := "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i, v := range b {
		b[i] = letters[v%byte(len(letters))]
	}

	return string(b)
}

func Case2Camel(name string) string {
	names := strings.Split(name, "_")

	var newName string
	for _, item := range names {
		buffer := NewBuffer()
		for i, r := range item {
			if i == 0 {
				buffer.Append(unicode.ToUpper(r))
			} else {
				buffer.Append(r)
			}
		}

		newName += buffer.String()
	}

	return newName
}

func Camel2Case(name string) string {
	buffer := NewBuffer()
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i != 0 {
				buffer.Append('_')
			}
			buffer.Append(unicode.ToLower(r))
		} else {
			buffer.Append(r)
		}
	}

	return buffer.String()
}

type Buffer struct {
	*bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{Buffer: new(bytes.Buffer)}
}

func (b *Buffer) Append(i any) *Buffer {
	switch val := i.(type) {
	case int:
		b.append(strconv.Itoa(val))
	case int64:
		b.append(strconv.FormatInt(val, 10))
	case uint:
		b.append(strconv.FormatUint(uint64(val), 10))
	case uint64:
		b.append(strconv.FormatUint(val, 10))
	case string:
		b.append(val)
	case []byte:
		b.Write(val)
	case rune:
		b.WriteRune(val)
	}
	return b
}

func (b *Buffer) append(s string) *Buffer {
	b.WriteString(s)

	return b
}
