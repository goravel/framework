package testing

type AssertableJSON interface {
	Json() map[string]any
	Count(key string, value int) AssertableJSON
	Has(key string) AssertableJSON
	HasAll(keys []string) AssertableJSON
	HasAny(keys []string) AssertableJSON
	Missing(key string) AssertableJSON
	MissingAll(keys []string) AssertableJSON
	Where(key string, value any) AssertableJSON
	WhereNot(key string, value any) AssertableJSON
}
