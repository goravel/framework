package validation

import (
	"strings"

	httpvalidate "github.com/goravel/framework/contracts/validation"

	"github.com/gookit/validate"
)

func Rules(rules map[string]string) httpvalidate.Option {
	return func(options map[string]any) {
		if rules != nil && len(rules) > 0 {
			options["rules"] = rules
		}
	}
}

func CustomRules(rules []httpvalidate.Rule) httpvalidate.Option {
	return func(options map[string]any) {
		if rules != nil && len(rules) > 0 {
			options["customRules"] = rules
		}
	}
}

func Messages(messages map[string]string) httpvalidate.Option {
	return func(options map[string]any) {
		if messages != nil && len(messages) > 0 {
			options["messages"] = messages
		}
	}
}

func Attributes(attributes map[string]string) httpvalidate.Option {
	return func(options map[string]any) {
		if attributes != nil && len(attributes) > 0 {
			options["attributes"] = attributes
		}
	}
}

func PrepareForValidation(prepare func(data httpvalidate.Data)) httpvalidate.Option {
	return func(options map[string]any) {
		options["prepareForValidation"] = prepare
	}
}

func GenerateOptions(options []httpvalidate.Option) map[string]any {
	realOptions := make(map[string]any)
	for _, option := range options {
		option(realOptions)
	}

	return realOptions
}

func AppendOptions(validator *validate.Validation, options map[string]any) {
	if options["rules"] != nil {
		rules := options["rules"].(map[string]string)
		for key, value := range rules {
			validator.StringRule(key, value)
		}
	}

	if options["messages"] != nil {
		messages := options["messages"].(map[string]string)
		for key, value := range messages {
			messages[key] = strings.ReplaceAll(value, ":attribute", "{field}")
		}
		validator.AddMessages(messages)
	}

	if options["attributes"] != nil && len(options["attributes"].(map[string]string)) > 0 {
		validator.AddTranslates(options["attributes"].(map[string]string))
	}

	if options["customRules"] != nil {
		customRules := options["customRules"].([]httpvalidate.Rule)
		for _, customRule := range customRules {
			customRule := customRule
			validator.AddMessages(map[string]string{
				customRule.Signature(): strings.ReplaceAll(customRule.Message(), ":attribute", "{field}"),
			})
			validator.AddValidator(customRule.Signature(), func(val any, options ...any) bool {
				return customRule.Passes(validator, val, options...)
			})
		}
	}

	validator.Trans().FieldMap()
}
