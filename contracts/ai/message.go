package ai

// MessageRole is the speaker role for a conversation message.
type MessageRole string

const (
	RoleAssistant  MessageRole = "assistant"
	RoleToolResult MessageRole = "tool_result"
	RoleUser       MessageRole = "user"
)

// Message is a unit of conversation history.
type Message struct {
	Role    MessageRole
	Content string
	// ToolCallID is set on RoleToolResult messages to correlate a tool
	// result with the assistant ToolCall that requested it.
	ToolCallID string
	// ToolCalls is set on RoleAssistant messages when the model requests
	// one or more tool invocations instead of (or before) returning text.
	ToolCalls []ToolCall
}
