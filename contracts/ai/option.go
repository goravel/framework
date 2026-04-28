package ai

type Options struct {
	Provider    string
	Model       string
	Middlewares []Middleware
}

type ConversationOptions struct {
	Model       string
	Attachments []Attachment
	Middlewares []Middleware
}

// Option applies conversation options for provider selection and model behavior.
type Option func(options *Options)

type ConversationOption interface {
	ApplyConversation(options *ConversationOptions)
}
