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
}
