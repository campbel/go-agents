package agent

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

func UserMessage(content string) Message {
	return Message{
		Role:    string(RoleUser),
		Content: content,
	}
}

func AssistantMessage(content string) Message {
	return Message{
		Role:    string(RoleAssistant),
		Content: content,
	}
}

type Response struct {
	Content string `json:"content"`
	Error   error  `json:"error"`
}
