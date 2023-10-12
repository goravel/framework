package validation

type Option func(map[string]any)

//go:generate mockery --name=Validation
type Validation interface {
	// Make create a new validator instance.
	Make(data any, rules map[string]string, options ...Option) (Validator, error)
	// AddRules add the custom rules.
	AddRules([]Rule) error
	// Rules get the custom rules.
	Rules() []Rule
}

//go:generate mockery --name=Validator
type Validator interface {
	// Bind the data to the validation.
	Bind(ptr any) error
	// Errors get the validation errors.
	Errors() Errors
	// Fails determine if the validation fails.
	Fails() bool
}

//go:generate mockery --name=Errors
type Errors interface {
	// One gets the first error message for a given field.
	One(key ...string) string
	// Get gets all the error messages for a given field.
	Get(key string) map[string]string
	// All gets all the error messages.
	All() map[string]map[string]string
	// Has checks if there are any error messages for a given field.
	Has(key string) bool
}

type Data interface {
	// Get the value from the given key.
	Get(key string) (val any, exist bool)
	// Set the value for a given key.
	Set(key string, val any) error
}

type Rule interface {
	// Signature set the unique signature of the rule.
	Signature() string
	// Passes determine if the validation rule passes.
	Passes(data Data, val any, options ...any) bool
	// Message gets the validation error message.
	Message() string
}
