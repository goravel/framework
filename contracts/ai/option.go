package ai

type Options struct {
	Provider    string
	Model       string
	Middlewares []Middleware
}

type ConversationOptions struct {
	Attachments []Attachment
}

// Option applies AI options for provider selection and model behavior.
type Option func(options *Options)

type ConversationOption func(options *ConversationOptions)
