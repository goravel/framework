package ai

type Options struct {
	// ProviderChain is the ordered primary and failover provider list.
	ProviderChain []string
	// Model overrides the selected provider's default model.
	Model string
	// Middlewares appends middleware for the current agent request.
	Middlewares []Middleware
}

type ConversationOptions struct {
	Attachments []Attachment
}

// Option applies AI options for provider selection and model behavior.
type Option func(options *Options)

type ConversationOption func(options *ConversationOptions)
