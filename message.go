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

type ResponseKind string

const (
	ResponseKindContent ResponseKind = "content"
	ResponseKindUsage   ResponseKind = "usage"
	ResponseKindError   ResponseKind = "error"
)

type Response struct {
	Kind ResponseKind

	content string
	err     error
	usage   Usage
}

func (r Response) IsContentResponse() bool {
	return r.Kind == ResponseKindContent
}

func (r Response) IsUsageResponse() bool {
	return r.Kind == ResponseKindUsage
}

func (r Response) IsErrorResponse() bool {
	return r.Kind == ResponseKindError
}

func (r Response) Usage() Usage {
	if r.Kind != ResponseKindUsage {
		return Usage{}
	}
	return r.usage
}

func (r Response) Content() string {
	if r.Kind != ResponseKindContent {
		return ""
	}
	return r.content
}

func (r Response) Error() error {
	if r.Kind != ResponseKindError {
		return nil
	}
	return r.err
}

func NewContentResponse(content string) Response {
	return Response{
		Kind:    ResponseKindContent,
		content: content,
	}
}

func NewUsageResponse(usage Usage) Response {
	return Response{
		Kind:  ResponseKindUsage,
		usage: usage,
	}
}

func NewErrorResponse(err error) Response {
	return Response{
		Kind: ResponseKindError,
		err:  err,
	}
}

type Usage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}
