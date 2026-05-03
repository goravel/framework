package ai

import "context"

// ToolCall represents a single tool invocation requested by the model.
type ToolCall struct {
	// ID is the model-generated unique identifier for this call.
	ID string
	// Name is the tool function name the model wants to invoke.
	Name string
	// Args contains the decoded JSON arguments for the call.
	Args map[string]any
	// RawArgs is the raw JSON string of the arguments as returned by the model.
	// It is used when replaying tool calls back to the API to preserve exact formatting.
	RawArgs string
}

// Tool is a callable capability the model can invoke during a conversation.
type Tool interface {
	// Name returns the function name sent to the model. Must be unique per agent.
	Name() string
	// Description tells the model what the tool does and when to use it.
	Description() string
	// Parameters returns the JSON Schema object describing the tool's input.
	// Return nil if the tool takes no parameters.
	Parameters() map[string]any
	// Execute is called by the framework when the model invokes the tool.
	Execute(ctx context.Context, args map[string]any) (string, error)
}
