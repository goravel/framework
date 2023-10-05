package str

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/support/env"
)

type StringTestSuite struct {
	suite.Suite
}

func TestStringTestSuite(t *testing.T) {
	suite.Run(t, &StringTestSuite{})
}

func (s *StringTestSuite) SetupTest() {
}

func (s *StringTestSuite) TestAfter() {
	s.Equal("Framework", Of("GoravelFramework").After("Goravel").String())
	s.Equal("lel", Of("parallel").After("l").String())
	s.Equal("3def", Of("abc123def").After("2").String())
	s.Equal("abc123def", Of("abc123def").After("4").String())
	s.Equal("GoravelFramework", Of("GoravelFramework").After("").String())
}

func (s *StringTestSuite) TestAfterLast() {
	s.Equal("Framework", Of("GoravelFramework").AfterLast("Goravel").String())
	s.Equal("", Of("parallel").AfterLast("l").String())
	s.Equal("3def", Of("abc123def").AfterLast("2").String())
	s.Equal("abc123def", Of("abc123def").AfterLast("4").String())
}

func (s *StringTestSuite) TestAppend() {
	s.Equal("foobar", Of("foo").Append("bar").String())
	s.Equal("foobar", Of("foo").Append("bar").Append("").String())
	s.Equal("foobar", Of("foo").Append("bar").Append().String())
}

func (s *StringTestSuite) TestBasename() {
	s.Equal("str", Of("/framework/support/str").Basename().String())
	s.Equal("str", Of("/framework/support/str/").Basename().String())
	s.Equal("str", Of("str").Basename().String())
	s.Equal("str", Of("/str").Basename().String())
	s.Equal("str", Of("/str/").Basename().String())
	s.Equal("str", Of("str/").Basename().String())

	str := Of("/").Basename().String()
	if env.IsWindows() {
		s.Equal("\\", str)
	} else {
		s.Equal("/", str)
	}

	s.Equal(".", Of("").Basename().String())
	s.Equal("str", Of("/framework/support/str/str.go").Basename(".go").String())
}

func (s *StringTestSuite) TestBefore() {
	s.Equal("Goravel", Of("GoravelFramework").Before("Framework").String())
	s.Equal("para", Of("parallel").Before("l").String())
	s.Equal("abc123", Of("abc123def").Before("def").String())
	s.Equal("abc", Of("abc123def").Before("123").String())
}

func (s *StringTestSuite) TestBeforeLast() {
	s.Equal("Goravel", Of("GoravelFramework").BeforeLast("Framework").String())
	s.Equal("paralle", Of("parallel").BeforeLast("l").String())
	s.Equal("abc123", Of("abc123def").BeforeLast("def").String())
	s.Equal("abc", Of("abc123def").BeforeLast("123").String())
}

func (s *StringTestSuite) TestBetween() {
	s.Equal("foobarbaz", Of("foobarbaz").Between("", "b").String())
	s.Equal("foobarbaz", Of("foobarbaz").Between("f", "").String())
	s.Equal("foobarbaz", Of("foobarbaz").Between("", "").String())
	s.Equal("obar", Of("foobarbaz").Between("o", "b").String())
	s.Equal("bar", Of("foobarbaz").Between("foo", "baz").String())
	s.Equal("foo][bar][baz", Of("[foo][bar][baz]").Between("[", "]").String())
}

func (s *StringTestSuite) TestBetweenFirst() {
	s.Equal("foobarbaz", Of("foobarbaz").BetweenFirst("", "b").String())
	s.Equal("foobarbaz", Of("foobarbaz").BetweenFirst("f", "").String())
	s.Equal("foobarbaz", Of("foobarbaz").BetweenFirst("", "").String())
	s.Equal("o", Of("foobarbaz").BetweenFirst("o", "b").String())
	s.Equal("foo", Of("[foo][bar][baz]").BetweenFirst("[", "]").String())
	s.Equal("foobar", Of("foofoobarbaz").BetweenFirst("foo", "baz").String())
}

func (s *StringTestSuite) TestCamel() {
	s.Equal("goravelGOFramework", Of("Goravel_g_o_framework").Camel().String())
	s.Equal("goravelGOFramework", Of("Goravel_gO_framework").Camel().String())
	s.Equal("goravelGoFramework", Of("Goravel -_- go -_-  framework  ").Camel().String())

	s.Equal("fooBar", Of("FooBar").Camel().String())
	s.Equal("fooBar", Of("foo_bar").Camel().String())
	s.Equal("fooBar", Of("foo-Bar").Camel().String())
	s.Equal("fooBar", Of("foo bar").Camel().String())
	s.Equal("fooBar", Of("foo.bar").Camel().String())
}

func (s *StringTestSuite) TestCharAt() {
	s.Equal("好", Of("你好，世界！").CharAt(1))
	s.Equal("त", Of("नमस्ते, दुनिया!").CharAt(4))
	s.Equal("w", Of("Привет, world!").CharAt(8))
	s.Equal("계", Of("안녕하세요, 세계!").CharAt(-2))
	s.Equal("", Of("こんにちは、世界！").CharAt(-200))
}

func (s *StringTestSuite) TestContains() {
	s.True(Of("kkumar").Contains("uma"))
	s.True(Of("kkumar").Contains("kumar"))
	s.True(Of("kkumar").Contains("uma", "xyz"))
	s.False(Of("kkumar").Contains("xyz"))
	s.False(Of("kkumar").Contains(""))
}

func (s *StringTestSuite) TestContainsAll() {
	s.True(Of("krishan kumar").ContainsAll("krishan", "kumar"))
	s.True(Of("krishan kumar").ContainsAll("kumar"))
	s.False(Of("krishan kumar").ContainsAll("kumar", "xyz"))
}

func (s *StringTestSuite) TestDirname() {
	str := Of("/framework/support/str").Dirname().String()
	if env.IsWindows() {
		s.Equal("\\framework\\support", str)
	} else {
		s.Equal("/framework/support", str)
	}

	str = Of("/framework/support/str").Dirname(2).String()
	if env.IsWindows() {
		s.Equal("\\framework", str)
	} else {
		s.Equal("/framework", str)
	}

	s.Equal(".", Of("framework").Dirname().String())
	s.Equal(".", Of(".").Dirname().String())

	str = Of("/").Dirname().String()
	if env.IsWindows() {
		s.Equal("\\", str)
	} else {
		s.Equal("/", str)
	}

	str = Of("/framework/").Dirname(2).String()
	if env.IsWindows() {
		s.Equal("\\", str)
	} else {
		s.Equal("/", str)
	}
}

func (s *StringTestSuite) TestEndsWith() {
	s.True(Of("bowen").EndsWith("wen"))
	s.True(Of("bowen").EndsWith("bowen"))
	s.True(Of("bowen").EndsWith("wen", "xyz"))
	s.False(Of("bowen").EndsWith("xyz"))
	s.False(Of("bowen").EndsWith(""))
	s.False(Of("bowen").EndsWith())
	s.False(Of("bowen").EndsWith("N"))
	s.True(Of("a7.12").EndsWith("7.12"))
	// Test for  muti-byte string
	s.True(Of("你好").EndsWith("好"))
	s.True(Of("你好").EndsWith("你好"))
	s.True(Of("你好").EndsWith("好", "xyz"))
	s.False(Of("你好").EndsWith("xyz"))
	s.False(Of("你好").EndsWith(""))
}

func (s *StringTestSuite) TestExactly() {
	s.True(Of("foo").Exactly("foo"))
	s.False(Of("foo").Exactly("Foo"))
}

func (s *StringTestSuite) TestExcerpt() {
	s.Equal("...is a beautiful morn...", Of("This is a beautiful morning").Excerpt("beautiful", ExcerptOption{
		Radius: 5,
	}).String())
	s.Equal("This is a beautiful morning", Of("This is a beautiful morning").Excerpt("foo", ExcerptOption{
		Radius: 5,
	}).String())
	s.Equal("(...)is a beautiful morn(...)", Of("This is a beautiful morning").Excerpt("beautiful", ExcerptOption{
		Omission: "(...)",
		Radius:   5,
	}).String())
}

func (s *StringTestSuite) TestExplode() {
	s.Equal([]string{"Foo", "Bar", "Baz"}, Of("Foo Bar Baz").Explode(" "))
	// with limit
	s.Equal([]string{"Foo", "Bar Baz"}, Of("Foo Bar Baz").Explode(" ", 2))
	s.Equal([]string{"Foo", "Bar"}, Of("Foo Bar Baz").Explode(" ", -1))
	s.Equal([]string{}, Of("Foo Bar Baz").Explode(" ", -10))
}

func (s *StringTestSuite) TestFinish() {
	s.Equal("abbc", Of("ab").Finish("bc").String())
	s.Equal("abbc", Of("abbcbc").Finish("bc").String())
	s.Equal("abcbbc", Of("abcbbcbc").Finish("bc").String())
}

func (s *StringTestSuite) TestHeadline() {
	s.Equal("Hello", Of("hello").Headline().String())
	s.Equal("This Is A Headline", Of("this is a headline").Headline().String())
	s.Equal("Camelcase Is A Headline", Of("CamelCase is a headline").Headline().String())
	s.Equal("Kebab-Case Is A Headline", Of("kebab-case is a headline").Headline().String())
}

func (s *StringTestSuite) TestIs() {
	s.True(Of("foo").Is("foo", "bar", "baz"))
	s.True(Of("foo123").Is("bar*", "baz*", "foo*"))
	s.False(Of("foo").Is("bar", "baz"))
	s.True(Of("a.b").Is("a.b", "c.*"))
	s.False(Of("abc*").Is("abc\\*", "xyz*"))
	s.False(Of("").Is("foo"))
	s.True(Of("foo/bar/baz").Is("foo/*", "bar/*", "baz*"))
	// Is case-sensitive
	s.False(Of("foo/bar/baz").Is("*BAZ*"))
}

func (s *StringTestSuite) TestIsEmpty() {
	s.True(Of("").IsEmpty())
	s.False(Of("F").IsEmpty())
}

func (s *StringTestSuite) TestIsNotEmpty() {
	s.False(Of("").IsNotEmpty())
	s.True(Of("F").IsNotEmpty())
}

func (s *StringTestSuite) TestIsAscii() {
	s.True(Of("abc").IsAscii())
	s.False(Of("你好").IsAscii())
}

func (s *StringTestSuite) TestIsSlice() {
	// Test when the string represents a valid JSON array
	s.True(Of(`["apple", "banana", "cherry"]`).IsSlice())

	// Test when the string represents a valid JSON array with objects
	s.True(Of(`[{"name": "John"}, {"name": "Alice"}]`).IsSlice())

	// Test when the string represents an empty JSON array
	s.True(Of(`[]`).IsSlice())

	// Test when the string represents an invalid JSON object
	s.False(Of(`{"name": "John"}`).IsSlice())

	// Test when the string is not valid JSON
	s.False(Of(`Not a JSON array`).IsSlice())

	// Test when the string is empty
	s.False(Of("").IsSlice())
}

func (s *StringTestSuite) TestIsMap() {
	// Test when the string represents a valid JSON object
	s.True(Of(`{"name": "John", "age": 30}`).IsMap())

	// Test when the string represents a valid JSON object with nested objects
	s.True(Of(`{"person": {"name": "Alice", "age": 25}}`).IsMap())

	// Test when the string represents an empty JSON object
	s.True(Of(`{}`).IsMap())

	// Test when the string represents an invalid JSON array
	s.False(Of(`["apple", "banana", "cherry"]`).IsMap())

	// Test when the string is not valid JSON
	s.False(Of(`Not a JSON object`).IsMap())

	// Test when the string is empty
	s.False(Of("").IsMap())
}

func (s *StringTestSuite) TestIsUlid() {
	s.True(Of("01E65Z7XCHCR7X1P2MKF78ENRP").IsUlid())
	// lowercase characters are not allowed
	s.False(Of("01e65z7xchcr7x1p2mkf78enrp").IsUlid())
	// too short (ULIDS must be 26 characters long)
	s.False(Of("01E65Z7XCHCR7X1P2MKF78E").IsUlid())
	// contains invalid characters
	s.False(Of("01E65Z7XCHCR7X1P2MKF78ENR!").IsUlid())
}

func (s *StringTestSuite) TestIsUuid() {
	s.True(Of("3f2504e0-4f89-41d3-9a0c-0305e82c3301").IsUuid())
	s.False(Of("3f2504e0-4f89-41d3-9a0c-0305e82c3301-extra").IsUuid())
}

func (s *StringTestSuite) TestKebab() {
	s.Equal("goravel-framework", Of("GoravelFramework").Kebab().String())
}

func (s *StringTestSuite) TestLcFirst() {
	s.Equal("framework", Of("Framework").LcFirst().String())
	s.Equal("framework", Of("framework").LcFirst().String())
}

func (s *StringTestSuite) TestLength() {
	s.Equal(11, Of("foo bar baz").Length())
	s.Equal(0, Of("").Length())
}

func (s *StringTestSuite) TestLimit() {
	s.Equal("This is...", Of("This is a beautiful morning").Limit(7).String())
	s.Equal("This is****", Of("This is a beautiful morning").Limit(7, "****").String())
	s.Equal("这是一...", Of("这是一段中文").Limit(3).String())
	s.Equal("这是一段中文", Of("这是一段中文").Limit(9).String())
}

func (s *StringTestSuite) TestLower() {
	s.Equal("foo bar baz", Of("FOO BAR BAZ").Lower().String())
	s.Equal("foo bar baz", Of("fOo Bar bAz").Lower().String())
}

func (s *StringTestSuite) TestLTrim() {
	s.Equal("foo ", Of(" foo ").LTrim().String())
}

func (s *StringTestSuite) TestMask() {
	s.Equal("kri**************", Of("krishan@email.com").Mask("*", 3).String())
	s.Equal("*******@email.com", Of("krishan@email.com").Mask("*", 0, 7).String())
	s.Equal("kris*************", Of("krishan@email.com").Mask("*", -13).String())
	s.Equal("kris***@email.com", Of("krishan@email.com").Mask("*", -13, 3).String())

	s.Equal("*****************", Of("krishan@email.com").Mask("*", -17).String())
	s.Equal("*****an@email.com", Of("krishan@email.com").Mask("*", -99, 5).String())

	s.Equal("krishan@email.com", Of("krishan@email.com").Mask("*", 17).String())
	s.Equal("krishan@email.com", Of("krishan@email.com").Mask("*", 17, 99).String())

	s.Equal("krishan@email.com", Of("krishan@email.com").Mask("", 3).String())

	s.Equal("krissssssssssssss", Of("krishan@email.com").Mask("something", 3).String())

	s.Equal("这是一***", Of("这是一段中文").Mask("*", 3).String())
	s.Equal("**一段中文", Of("这是一段中文").Mask("*", 0, 2).String())
}

func (s *StringTestSuite) TestMatch() {
	s.Equal("World", Of("Hello, World!").Match("World").String())
	s.Equal("(test)", Of("This is a (test) string").Match(`\([^)]+\)`).String())
	s.Equal("123", Of("abc123def456def").Match(`\d+`).String())
	s.Equal("", Of("No match here").Match(`\d+`).String())
	s.Equal("Hello, World!", Of("Hello, World!").Match("").String())
	s.Equal("[456]", Of("123 [456]").Match(`\[456\]`).String())
}

func (s *StringTestSuite) TestMatchAll() {
	s.Equal([]string{"World"}, Of("Hello, World!").MatchAll("World"))
	s.Equal([]string{"(test)"}, Of("This is a (test) string").MatchAll(`\([^)]+\)`))
	s.Equal([]string{"123", "456"}, Of("abc123def456def").MatchAll(`\d+`))
	s.Equal([]string(nil), Of("No match here").MatchAll(`\d+`))
	s.Equal([]string{"Hello, World!"}, Of("Hello, World!").MatchAll(""))
	s.Equal([]string{"[456]"}, Of("123 [456]").MatchAll(`\[456\]`))
}

func (s *StringTestSuite) TestIsMatch() {
	// Test matching with a single pattern
	s.True(Of("Hello, Goravel!").IsMatch(`.*,.*!`))
	s.True(Of("Hello, Goravel!").IsMatch(`^.*$(.*)`))
	s.True(Of("Hello, Goravel!").IsMatch(`(?i)goravel`))
	s.True(Of("Hello, GOravel!").IsMatch(`^(.*(.*(.*)))`))

	// Test non-matching with a single pattern
	s.False(Of("Hello, Goravel!").IsMatch(`H.o`))
	s.False(Of("Hello, Goravel!").IsMatch(`^goravel!`))
	s.False(Of("Hello, Goravel!").IsMatch(`goravel!(.*)`))
	s.False(Of("Hello, Goravel!").IsMatch(`^[a-zA-Z,!]+$`))

	// Test with multiple patterns
	s.True(Of("Hello, Goravel!").IsMatch(`.*,.*!`, `H.o`))
	s.True(Of("Hello, Goravel!").IsMatch(`(?i)goravel`, `^.*$(.*)`))
	s.True(Of("Hello, Goravel!").IsMatch(`(?i)goravel`, `goravel!(.*)`))
	s.True(Of("Hello, Goravel!").IsMatch(`^[a-zA-Z,!]+$`, `^(.*(.*(.*)))`))
}

func (s *StringTestSuite) TestNewLine() {
	s.Equal("Goravel\n", Of("Goravel").NewLine().String())
	s.Equal("Goravel\n\nbar", Of("Goravel").NewLine(2).Append("bar").String())
}

func (s *StringTestSuite) TestPadBoth() {
	// Test padding with spaces
	s.Equal("   Hello   ", Of("Hello").PadBoth(11, " ").String())
	s.Equal("  World!  ", Of("World!").PadBoth(10, " ").String())
	s.Equal("==Hello===", Of("Hello").PadBoth(10, "=").String())
	s.Equal("Hello", Of("Hello").PadBoth(3, " ").String())
	s.Equal("      ", Of("").PadBoth(6, " ").String())
}

func (s *StringTestSuite) TestPadLeft() {
	s.Equal("   Goravel", Of("Goravel").PadLeft(10, " ").String())
	s.Equal("==Goravel", Of("Goravel").PadLeft(9, "=").String())
	s.Equal("Goravel", Of("Goravel").PadLeft(3, " ").String())
}

func (s *StringTestSuite) TestPadRight() {
	s.Equal("Goravel   ", Of("Goravel").PadRight(10, " ").String())
	s.Equal("Goravel==", Of("Goravel").PadRight(9, "=").String())
	s.Equal("Goravel", Of("Goravel").PadRight(3, " ").String())
}

func (s *StringTestSuite) TestPipe() {
	callback := func(str string) string {
		return Of(str).Append("bar").String()
	}
	s.Equal("foobar", Of("foo").Pipe(callback).String())
}

func (s *StringTestSuite) TestPrepend() {
	s.Equal("foobar", Of("bar").Prepend("foo").String())
	s.Equal("foobar", Of("bar").Prepend("foo").Prepend("").String())
	s.Equal("foobar", Of("bar").Prepend("foo").Prepend().String())
}

func (s *StringTestSuite) TestRemove() {
	s.Equal("Fbar", Of("Foobar").Remove("o").String())
	s.Equal("Foo", Of("Foobar").Remove("bar").String())
	s.Equal("oobar", Of("Foobar").Remove("F").String())
	s.Equal("Foobar", Of("Foobar").Remove("f").String())

	s.Equal("Fbr", Of("Foobar").Remove("o", "a").String())
	s.Equal("Fooar", Of("Foobar").Remove("f", "b").String())
	s.Equal("Foobar", Of("Foo|bar").Remove("f", "|").String())
}

func (s *StringTestSuite) TestRepeat() {
	s.Equal("aaaaa", Of("a").Repeat(5).String())
	s.Equal("", Of("").Repeat(5).String())
}

func (s *StringTestSuite) TestReplace() {
	s.Equal("foo/foo/foo", Of("?/?/?").Replace("?", "foo").String())
	s.Equal("foo/foo/foo", Of("x/x/x").Replace("X", "foo", false).String())
	s.Equal("bar/bar", Of("?/?").Replace("?", "bar").String())
	s.Equal("?/?/?", Of("? ? ?").Replace(" ", "/").String())
}

func (s *StringTestSuite) TestReplaceEnd() {
	s.Equal("Golang is great!", Of("Golang is good!").ReplaceEnd("good!", "great!").String())
	s.Equal("Hello, World!", Of("Hello, Earth!").ReplaceEnd("Earth!", "World!").String())
	s.Equal("München Berlin", Of("München Frankfurt").ReplaceEnd("Frankfurt", "Berlin").String())
	s.Equal("Café Latte", Of("Café Americano").ReplaceEnd("Americano", "Latte").String())
	s.Equal("Golang is good!", Of("Golang is good!").ReplaceEnd("", "great!").String())
	s.Equal("Golang is good!", Of("Golang is good!").ReplaceEnd("excellent!", "great!").String())
}

func (s *StringTestSuite) TestReplaceFirst() {
	s.Equal("fooqux foobar", Of("foobar foobar").ReplaceFirst("bar", "qux").String())
	s.Equal("foo/qux? foo/bar?", Of("foo/bar? foo/bar?").ReplaceFirst("bar?", "qux?").String())
	s.Equal("foo foobar", Of("foobar foobar").ReplaceFirst("bar", "").String())
	s.Equal("foobar foobar", Of("foobar foobar").ReplaceFirst("xxx", "yyy").String())
	s.Equal("foobar foobar", Of("foobar foobar").ReplaceFirst("", "yyy").String())
	// Test for multibyte string support
	s.Equal("Jxxxnköping Malmö", Of("Jönköping Malmö").ReplaceFirst("ö", "xxx").String())
	s.Equal("Jönköping Malmö", Of("Jönköping Malmö").ReplaceFirst("", "yyy").String())
}

func (s *StringTestSuite) TestReplaceLast() {
	s.Equal("foobar fooqux", Of("foobar foobar").ReplaceLast("bar", "qux").String())
	s.Equal("foo/bar? foo/qux?", Of("foo/bar? foo/bar?").ReplaceLast("bar?", "qux?").String())
	s.Equal("foobar foo", Of("foobar foobar").ReplaceLast("bar", "").String())
	s.Equal("foobar foobar", Of("foobar foobar").ReplaceLast("xxx", "yyy").String())
	s.Equal("foobar foobar", Of("foobar foobar").ReplaceLast("", "yyy").String())
	// Test for multibyte string support
	s.Equal("Malmö Jönkxxxping", Of("Malmö Jönköping").ReplaceLast("ö", "xxx").String())
	s.Equal("Malmö Jönköping", Of("Malmö Jönköping").ReplaceLast("", "yyy").String())
}

func (s *StringTestSuite) TestReplaceMatches() {
	s.Equal("Golang is great!", Of("Golang is good!").ReplaceMatches("good", "great").String())
	s.Equal("Hello, World!", Of("Hello, Earth!").ReplaceMatches("Earth", "World").String())
	s.Equal("Apples, Apples, Apples", Of("Oranges, Oranges, Oranges").ReplaceMatches("Oranges", "Apples").String())
	s.Equal("1, 2, 3, 4, 5", Of("10, 20, 30, 40, 50").ReplaceMatches("0", "").String())
	s.Equal("München Berlin", Of("München Frankfurt").ReplaceMatches("Frankfurt", "Berlin").String())
	s.Equal("Café Latte", Of("Café Americano").ReplaceMatches("Americano", "Latte").String())
	s.Equal("The quick brown fox", Of("The quick brown fox").ReplaceMatches(`\b([a-z])`, `$1`).String())
	s.Equal("One, One, One", Of("1, 2, 3").ReplaceMatches(`\d`, "One").String())
	s.Equal("Hello, World!", Of("Hello, World!").ReplaceMatches("Earth", "").String())
	s.Equal("Hello, World!", Of("Hello, World!").ReplaceMatches("Golang", "Great").String())
}

func (s *StringTestSuite) TestReplaceStart() {
	s.Equal("foobar foobar", Of("foobar foobar").ReplaceStart("bar", "qux").String())
	s.Equal("foo/bar? foo/bar?", Of("foo/bar? foo/bar?").ReplaceStart("bar?", "qux?").String())
	s.Equal("quxbar foobar", Of("foobar foobar").ReplaceStart("foo", "qux").String())
	s.Equal("qux? foo/bar?", Of("foo/bar? foo/bar?").ReplaceStart("foo/bar?", "qux?").String())
	s.Equal("bar foobar", Of("foobar foobar").ReplaceStart("foo", "").String())
	s.Equal("1", Of("0").ReplaceStart("0", "1").String())
	// Test for multibyte string support
	s.Equal("xxxnköping Malmö", Of("Jönköping Malmö").ReplaceStart("Jö", "xxx").String())
	s.Equal("Jönköping Malmö", Of("Jönköping Malmö").ReplaceStart("", "yyy").String())
}

func (s *StringTestSuite) TestRTrim() {
	s.Equal(" foo", Of(" foo ").RTrim().String())
	s.Equal(" foo", Of(" foo__").RTrim("_").String())
}

func (s *StringTestSuite) TestSnake() {
	s.Equal("goravel_g_o_framework", Of("GoravelGOFramework").Snake().String())
	s.Equal("goravel_go_framework", Of("GoravelGoFramework").Snake().String())
	s.Equal("goravel go framework", Of("GoravelGoFramework").Snake(" ").String())
	s.Equal("goravel_go_framework", Of("Goravel Go Framework").Snake().String())
	s.Equal("goravel_go_framework", Of("Goravel    Go      Framework   ").Snake().String())
	s.Equal("goravel__go__framework", Of("GoravelGoFramework").Snake("__").String())
	s.Equal("żółta_łódka", Of("ŻółtaŁódka").Snake().String())
}

func (s *StringTestSuite) TestSplit() {
	s.Equal([]string{"one", "two", "three", "four"}, Of("one-two-three-four").Split("-"))
	s.Equal([]string{"", "", "D", "E", "", ""}, Of(",,D,E,,").Split(","))
	s.Equal([]string{"one", "two", "three,four"}, Of("one,two,three,four").Split(",", 3))
}

func (s *StringTestSuite) TestSquish() {
	s.Equal("Hello World", Of("  Hello   World  ").Squish().String())
	s.Equal("A B C", Of("A  B  C").Squish().String())
	s.Equal("Lorem ipsum dolor sit amet", Of(" Lorem   ipsum \n  dolor  sit \t amet ").Squish().String())
	s.Equal("Leading and trailing spaces", Of("  Leading  "+
		"and trailing "+
		" spaces  ").Squish().String())
	s.Equal("", Of("").Squish().String())
}

func (s *StringTestSuite) TestStart() {
	s.Equal("/test/string", Of("test/string").Start("/").String())
	s.Equal("/test/string", Of("/test/string").Start("/").String())
	s.Equal("/test/string", Of("//test/string").Start("/").String())
}

func (s *StringTestSuite) TestStartsWith() {
	s.True(Of("Wenbo Han").StartsWith("Wen"))
	s.True(Of("Wenbo Han").StartsWith("Wenbo"))
	s.True(Of("Wenbo Han").StartsWith("Han", "Wen"))
	s.False(Of("Wenbo Han").StartsWith())
	s.False(Of("Wenbo Han").StartsWith("we"))
	s.True(Of("Jönköping").StartsWith("Jö"))
	s.False(Of("Jönköping").StartsWith("Jonko"))
}

func (s *StringTestSuite) TestStudly() {
	s.Equal("GoravelGOFramework", Of("Goravel_g_o_framework").Studly().String())
	s.Equal("GoravelGOFramework", Of("Goravel_gO_framework").Studly().String())
	s.Equal("GoravelGoFramework", Of("Goravel -_- go -_-  framework  ").Studly().String())

	s.Equal("FooBar", Of("FooBar").Studly().String())
	s.Equal("FooBar", Of("foo_bar").Studly().String())
	s.Equal("FooBar", Of("foo-Bar").Studly().String())
	s.Equal("FooBar", Of("foo bar").Studly().String())
	s.Equal("FooBar", Of("foo.bar").Studly().String())
}

func (s *StringTestSuite) TestSubstr() {
	s.Equal("Ё", Of("БГДЖИЛЁ").Substr(-1).String())
	s.Equal("ЛЁ", Of("БГДЖИЛЁ").Substr(-2).String())
	s.Equal("И", Of("БГДЖИЛЁ").Substr(-3, 1).String())
	s.Equal("ДЖИЛ", Of("БГДЖИЛЁ").Substr(2, -1).String())
	s.Equal("", Of("БГДЖИЛЁ").Substr(4, -4).String())
	s.Equal("ИЛ", Of("БГДЖИЛЁ").Substr(-3, -1).String())
	s.Equal("ГДЖИЛЁ", Of("БГДЖИЛЁ").Substr(1).String())
	s.Equal("ГДЖ", Of("БГДЖИЛЁ").Substr(1, 3).String())
	s.Equal("БГДЖ", Of("БГДЖИЛЁ").Substr(0, 4).String())
	s.Equal("Ё", Of("БГДЖИЛЁ").Substr(-1, 1).String())
	s.Equal("", Of("Б").Substr(2).String())
}

func (s *StringTestSuite) TestSwap() {
	s.Equal("Go is excellent", Of("Golang is awesome").Swap(map[string]string{
		"Golang":  "Go",
		"awesome": "excellent",
	}).String())
	s.Equal("Golang is awesome", Of("Golang is awesome").Swap(map[string]string{}).String())
	s.Equal("Golang is awesome", Of("Golang is awesome").Swap(map[string]string{
		"":        "Go",
		"awesome": "excellent",
	}).String())
}

func (s *StringTestSuite) TestTap() {
	tap := Of("foobarbaz")
	fromTehTap := ""
	tap = tap.Tap(func(s String) {
		fromTehTap = s.Substr(0, 3).String()
	})
	s.Equal("foo", fromTehTap)
	s.Equal("foobarbaz", tap.String())
}

func (s *StringTestSuite) TestTitle() {
	s.Equal("Krishan Kumar", Of("krishan kumar").Title().String())
	s.Equal("Krishan Kumar", Of("kriSHan kuMAr").Title().String())
}

func (s *StringTestSuite) TestTrim() {
	s.Equal("foo", Of(" foo ").Trim().String())
	s.Equal("foo", Of("_foo_").Trim("_").String())
}

func (s *StringTestSuite) TestUcFirst() {
	s.Equal("", Of("").UcFirst().String())
	s.Equal("Framework", Of("framework").UcFirst().String())
	s.Equal("Framework", Of("Framework").UcFirst().String())
	s.Equal(" framework", Of(" framework").UcFirst().String())
	s.Equal("Goravel framework", Of("goravel framework").UcFirst().String())
}

func (s *StringTestSuite) TestUcSplit() {
	s.Equal([]string{"Krishan", "Kumar"}, Of("KrishanKumar").UcSplit())
	s.Equal([]string{"Hello", "From", "Goravel"}, Of("HelloFromGoravel").UcSplit())
	s.Equal([]string{"He_llo_", "World"}, Of("He_llo_World").UcSplit())
}

func (s *StringTestSuite) TestUnless() {
	str := Of("Hello, World!")

	// Test case 1: The callback returns true, so the fallback should not be applied
	s.Equal("Hello, World!", str.Unless(func(s *String) bool {
		return true
	}, func(s *String) *String {
		return Of("This should not be applied")
	}).String())

	// Test case 2: The callback returns false, so the fallback should be applied
	s.Equal("Fallback Applied", str.Unless(func(s *String) bool {
		return false
	}, func(s *String) *String {
		return Of("Fallback Applied")
	}).String())

	// Test case 3: Testing with an empty string
	s.Equal("Fallback Applied", Of("").Unless(func(s *String) bool {
		return false
	}, func(s *String) *String {
		return Of("Fallback Applied")
	}).String())
}

func (s *StringTestSuite) TestUpper() {
	s.Equal("FOO BAR BAZ", Of("foo bar baz").Upper().String())
	s.Equal("FOO BAR BAZ", Of("foO bAr BaZ").Upper().String())
}

func (s *StringTestSuite) TestWhen() {
	// true
	s.Equal("when true", Of("when ").When(true, func(s *String) *String {
		return s.Append("true")
	}).String())
	s.Equal("gets a value from if", Of("gets a value ").When(true, func(s *String) *String {
		return s.Append("from if")
	}).String())

	// false
	s.Equal("when", Of("when").When(false, func(s *String) *String {
		return s.Append("true")
	}).String())

	s.Equal("when false fallbacks to default", Of("when false ").When(false, func(s *String) *String {
		return s.Append("true")
	}, func(s *String) *String {
		return s.Append("fallbacks to default")
	}).String())
}

func (s *StringTestSuite) TestWhenContains() {
	s.Equal("Tony Stark", Of("stark").WhenContains("tar", func(s *String) *String {
		return s.Prepend("Tony ").Title()
	}, func(s *String) *String {
		return s.Prepend("Arno ").Title()
	}).String())

	s.Equal("stark", Of("stark").WhenContains("xxx", func(s *String) *String {
		return s.Prepend("Tony ").Title()
	}).String())

	s.Equal("Arno Stark", Of("stark").WhenContains("xxx", func(s *String) *String {
		return s.Prepend("Tony ").Title()
	}, func(s *String) *String {
		return s.Prepend("Arno ").Title()
	}).String())
}

func (s *StringTestSuite) TestWhenContainsAll() {
	// Test when all values are present
	s.Equal("Tony Stark", Of("tony stark").WhenContainsAll([]string{"tony", "stark"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())

	// Test when not all values are present
	s.Equal("tony stark", Of("tony stark").WhenContainsAll([]string{"xxx"},
		func(s *String) *String {
			return s.Title()
		},
	).String())

	// Test when some values are present and some are not
	s.Equal("TonyStark", Of("tony stark").WhenContainsAll([]string{"tony", "xxx"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())
}

func (s *StringTestSuite) TestWhenEmpty() {
	// Test when the string is empty
	s.Equal("DEFAULT", Of("").WhenEmpty(
		func(s *String) *String {
			return s.Append("default").Upper()
		}).String())

	// Test when the string is not empty
	s.Equal("non-empty", Of("non-empty").WhenEmpty(
		func(s *String) *String {
			return s.Append("default")
		},
	).String())
}

func (s *StringTestSuite) TestWhenIsAscii() {
	s.Equal("Ascii: A", Of("A").WhenIsAscii(
		func(s *String) *String {
			return s.Prepend("Ascii: ")
		}).String())
	s.Equal("ù", Of("ù").WhenIsAscii(
		func(s *String) *String {
			return s.Prepend("Ascii: ")
		}).String())
	s.Equal("Not Ascii: ù", Of("ù").WhenIsAscii(
		func(s *String) *String {
			return s.Prepend("Ascii: ")
		},
		func(s *String) *String {
			return s.Prepend("Not Ascii: ")
		},
	).String())
}

func (s *StringTestSuite) TestWhenNotEmpty() {
	// Test when the string is not empty
	s.Equal("UPPERCASE", Of("uppercase").WhenNotEmpty(
		func(s *String) *String {
			return s.Upper()
		},
	).String())

	// Test when the string is empty
	s.Equal("", Of("").WhenNotEmpty(
		func(s *String) *String {
			return s.Append("not empty")
		},
		func(s *String) *String {
			return s.Upper()
		},
	).String())
}

func (s *StringTestSuite) TestWhenStartsWith() {
	// Test when the string starts with a specific prefix
	s.Equal("Tony Stark", Of("tony stark").WhenStartsWith([]string{"ton"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())

	// Test when the string starts with any of the specified prefixes
	s.Equal("Tony Stark", Of("tony stark").WhenStartsWith([]string{"ton", "not"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())

	// Test when the string does not start with the specified prefix
	s.Equal("tony stark", Of("tony stark").WhenStartsWith([]string{"xxx"},
		func(s *String) *String {
			return s.Title()
		},
	).String())

	// Test when the string starts with one of the specified prefixes and not the other
	s.Equal("Tony Stark", Of("tony stark").WhenStartsWith([]string{"tony", "xxx"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())
}

func (s *StringTestSuite) TestWhenEndsWith() {
	// Test when the string ends with a specific suffix
	s.Equal("Tony Stark", Of("tony stark").WhenEndsWith([]string{"ark"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())

	// Test when the string ends with any of the specified suffixes
	s.Equal("Tony Stark", Of("tony stark").WhenEndsWith([]string{"kra", "ark"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())

	// Test when the string does not end with the specified suffix
	s.Equal("tony stark", Of("tony stark").WhenEndsWith([]string{"xxx"},
		func(s *String) *String {
			return s.Title()
		},
	).String())

	// Test when the string ends with one of the specified suffixes and not the other
	s.Equal("TonyStark", Of("tony stark").WhenEndsWith([]string{"tony", "xxx"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())
}

func (s *StringTestSuite) TestWhenExactly() {
	// Test when the string exactly matches the expected value
	s.Equal("Nailed it...!", Of("Tony Stark").WhenExactly("Tony Stark",
		func(s *String) *String {
			return Of("Nailed it...!")
		},
		func(s *String) *String {
			return Of("Swing and a miss...!")
		},
	).String())

	// Test when the string does not exactly match the expected value
	s.Equal("Swing and a miss...!", Of("Tony Stark").WhenExactly("Iron Man",
		func(s *String) *String {
			return Of("Nailed it...!")
		},
		func(s *String) *String {
			return Of("Swing and a miss...!")
		},
	).String())

	// Test when the string exactly matches the expected value with no "else" callback
	s.Equal("Tony Stark", Of("Tony Stark").WhenExactly("Iron Man",
		func(s *String) *String {
			return Of("Nailed it...!")
		},
	).String())
}

func (s *StringTestSuite) TestWhenNotExactly() {
	// Test when the string does not exactly match the expected value with an "else" callback
	s.Equal("Iron Man", Of("Tony").WhenNotExactly("Tony Stark",
		func(s *String) *String {
			return Of("Iron Man")
		},
	).String())

	// Test when the string does not exactly match the expected value with both "if" and "else" callbacks
	s.Equal("Swing and a miss...!", Of("Tony Stark").WhenNotExactly("Tony Stark",
		func(s *String) *String {
			return Of("Iron Man")
		},
		func(s *String) *String {
			return Of("Swing and a miss...!")
		},
	).String())
}

func (s *StringTestSuite) TestWhenIs() {
	// Test when the string exactly matches the expected value with an "if" callback
	s.Equal("Winner: /", Of("/").WhenIs("/",
		func(s *String) *String {
			return s.Prepend("Winner: ")
		},
		func(s *String) *String {
			return Of("Try again")
		},
	).String())

	// Test when the string does not exactly match the expected value with an "if" callback
	s.Equal("/", Of("/").WhenIs(" /",
		func(s *String) *String {
			return s.Prepend("Winner: ")
		},
	).String())

	// Test when the string does not exactly match the expected value with both "if" and "else" callbacks
	s.Equal("Try again", Of("/").WhenIs(" /",
		func(s *String) *String {
			return s.Prepend("Winner: ")
		},
		func(s *String) *String {
			return Of("Try again")
		},
	).String())

	// Test when the string matches a pattern using wildcard and "if" callback
	s.Equal("Winner: foo/bar/baz", Of("foo/bar/baz").WhenIs("foo/*",
		func(s *String) *String {
			return s.Prepend("Winner: ")
		},
	).String())
}

func (s *StringTestSuite) TestWhenIsUlid() {
	// Test when the string is a valid ULID with an "if" callback
	s.Equal("Ulid: 01GJSNW9MAF792C0XYY8RX6QFT", Of("01GJSNW9MAF792C0XYY8RX6QFT").WhenIsUlid(
		func(s *String) *String {
			return s.Prepend("Ulid: ")
		},
		func(s *String) *String {
			return s.Prepend("Not Ulid: ")
		},
	).String())

	// Test when the string is not a valid ULID with an "if" callback
	s.Equal("2cdc7039-65a6-4ac7-8e5d-d554a98", Of("2cdc7039-65a6-4ac7-8e5d-d554a98").WhenIsUlid(
		func(s *String) *String {
			return s.Prepend("Ulid: ")
		},
	).String())

	// Test when the string is not a valid ULID with both "if" and "else" callbacks
	s.Equal("Not Ulid: ss-01GJSNW9MAF792C0XYY8RX6QFT", Of("ss-01GJSNW9MAF792C0XYY8RX6QFT").WhenIsUlid(
		func(s *String) *String {
			return s.Prepend("Ulid: ")
		},
		func(s *String) *String {
			return s.Prepend("Not Ulid: ")
		},
	).String())
}

func (s *StringTestSuite) TestWhenIsUuid() {
	// Test when the string is a valid UUID with an "if" callback
	s.Equal("Uuid: 2cdc7039-65a6-4ac7-8e5d-d554a98e7b15", Of("2cdc7039-65a6-4ac7-8e5d-d554a98e7b15").WhenIsUuid(
		func(s *String) *String {
			return s.Prepend("Uuid: ")
		},
		func(s *String) *String {
			return s.Prepend("Not Uuid: ")
		},
	).String())

	s.Equal("2cdc7039-65a6-4ac7-8e5d-d554a98", Of("2cdc7039-65a6-4ac7-8e5d-d554a98").WhenIsUuid(
		func(s *String) *String {
			return s.Prepend("Uuid: ")
		},
	).String())

	s.Equal("Not Uuid: 2cdc7039-65a6-4ac7-8e5d-d554a98", Of("2cdc7039-65a6-4ac7-8e5d-d554a98").WhenIsUuid(
		func(s *String) *String {
			return s.Prepend("Uuid: ")
		},
		func(s *String) *String {
			return s.Prepend("Not Uuid: ")
		},
	).String())
}

func (s *StringTestSuite) TestWhenTest() {
	// Test when the regular expression matches with an "if" callback
	s.Equal("Winner: foo bar", Of("foo bar").WhenTest(`bar*`,
		func(s *String) *String {
			return s.Prepend("Winner: ")
		},
		func(s *String) *String {
			return Of("Try again")
		},
	).String())

	// Test when the regular expression does not match with an "if" callback
	s.Equal("Try again", Of("foo bar").WhenTest(`/link/`,
		func(s *String) *String {
			return s.Prepend("Winner: ")
		},
		func(s *String) *String {
			return Of("Try again")
		},
	).String())

	// Test when the regular expression does not match with both "if" and "else" callbacks
	s.Equal("foo bar", Of("foo bar").WhenTest(`/link/`,
		func(s *String) *String {
			return s.Prepend("Winner: ")
		},
	).String())
}

func (s *StringTestSuite) TestWordCount() {
	s.Equal(2, Of("Hello, world!").WordCount())
	s.Equal(10, Of("Hi, this is my first contribution to the Goravel framework.").WordCount())
}

func (s *StringTestSuite) TestWords() {
	s.Equal("Perfectly balanced, as >>>", Of("Perfectly balanced, as all things should be.").Words(3, " >>>").String())
	s.Equal("Perfectly balanced, as all things should be.", Of("Perfectly balanced, as all things should be.").Words(100).String())
}

func TestFieldsFunc(t *testing.T) {
	tests := []struct {
		input          string
		shouldPreserve []func(rune) bool
		expected       []string
	}{
		// Test case 1: Basic word splitting with space separator.
		{
			input:    "Hello World",
			expected: []string{"Hello", "World"},
		},
		// Test case 2: Splitting with space and preserving hyphen.
		{
			input:          "Hello-World",
			shouldPreserve: []func(rune) bool{func(r rune) bool { return r == '-' }},
			expected:       []string{"Hello", "-World"},
		},
		// Test case 3: Splitting with space and preserving multiple characters.
		{
			input: "Hello-World,This,Is,a,Test",
			shouldPreserve: []func(rune) bool{
				func(r rune) bool { return r == '-' },
				func(r rune) bool { return r == ',' },
			},
			expected: []string{"Hello", "-World", ",This", ",Is", ",a", ",Test"},
		},
		// Test case 4: No splitting when no separator is found.
		{
			input:    "HelloWorld",
			expected: []string{"HelloWorld"},
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := fieldsFunc(test.input, func(r rune) bool { return r == ' ' }, test.shouldPreserve...)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestSubstr(t *testing.T) {
	assert.Equal(t, "world", Substr("Hello, world!", 7, 5))
	assert.Equal(t, "", Substr("Golang", 10))
	assert.Equal(t, "tine", Substr("Goroutines", -5, 4))
	assert.Equal(t, "ic", Substr("Unicode", 2, -3))
	assert.Equal(t, "esting", Substr("Testing", 1, 10))
	assert.Equal(t, "", Substr("", 0, 5))
	assert.Equal(t, "世界！", Substr("你好，世界！", 3, 3))
}

func TestMaximum(t *testing.T) {
	assert.Equal(t, 10, maximum(5, 10))
	assert.Equal(t, 3.14, maximum(3.14, 2.71))
	assert.Equal(t, "banana", maximum("apple", "banana"))
	assert.Equal(t, -5, maximum(-5, -10))
	assert.Equal(t, 42, maximum(42, 42))
}

func TestRandom(t *testing.T) {
	assert.Len(t, Random(10), 10)
	assert.Empty(t, Random(0))
	assert.Panics(t, func() {
		Random(-1)
	})
}

func TestCase2Camel(t *testing.T) {
	assert.Equal(t, "GoravelFramework", Case2Camel("goravel_framework"))
	assert.Equal(t, "GoravelFramework1", Case2Camel("goravel_framework1"))
	assert.Equal(t, "GoravelFramework", Case2Camel("GoravelFramework"))
}

func TestCamel2Case(t *testing.T) {
	assert.Equal(t, "goravel_framework", Camel2Case("GoravelFramework"))
	assert.Equal(t, "goravel_framework1", Camel2Case("GoravelFramework1"))
	assert.Equal(t, "goravel_framework", Camel2Case("goravel_framework"))
}
