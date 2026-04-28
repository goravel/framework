package ai

type Options struct {
	Provider    string
	Model       string
	Middlewares []Middleware
}

type PromptOptions struct {
	Model       string
	Attachments []Attachment
	Middlewares []Middleware
}

// Option applies conversation options for provider selection and model behavior.
type Option func(options *Options)

func (option Option) ApplyPrompt(options *PromptOptions) {
	agentOptions := &Options{}
	option(agentOptions)

	if agentOptions.Model != "" {
		options.Model = agentOptions.Model
	}
	if len(agentOptions.Middlewares) > 0 {
		options.Middlewares = append(options.Middlewares, agentOptions.Middlewares...)
	}
}

type PromptOption interface {
	ApplyPrompt(options *PromptOptions)
}

type PromptOptionFunc func(options *PromptOptions)

func (option PromptOptionFunc) ApplyPrompt(options *PromptOptions) {
	option(options)
}
