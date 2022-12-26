package validation

type Option func(map[string]any)

//go:generate mockery --name=Validation
type Validation interface {
	Make(data any, rules map[string]string, options ...Option) (Validator, error)
	AddRules([]Rule) error
	Rules() []Rule
}

//go:generate mockery --name=Validator
type Validator interface {
	Bind(ptr any) error
	Errors() Errors
	Fails() bool
}

//go:generate mockery --name=Errors
type Errors interface {
	One(key ...string) string
	Get(key string) map[string]string
	All() map[string]map[string]string
	Has(key string) bool
}

type Data interface {
	Get(key string) (val any, exist bool)
	Set(key string, val any) error
}

type Rule interface {
	Signature() string
	Passes(data Data, val any, options ...any) bool
	Message() string
}
