package testing

type AssertableJSON interface {
	Count(key string, value int) AssertableJSON
	First(key string, callback func(AssertableJSON)) AssertableJSON
	Each(key string, callback func(AssertableJSON)) AssertableJSON
	Has(key string) AssertableJSON
	HasAll(keys []string) AssertableJSON
	HasAny(keys []string) AssertableJSON
	HasWithScope(key string, length int, callback func(AssertableJSON)) AssertableJSON
	Json() map[string]any
	Missing(key string) AssertableJSON
	MissingAll(keys []string) AssertableJSON
	Where(key string, value any) AssertableJSON
	WhereNot(key string, value any) AssertableJSON
}
