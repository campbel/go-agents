package agent

type MessageKind string

const (
	MessageKindText MessageKind = "text"
	MessageKindFile MessageKind = "file"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

type Message struct {
	Role Role        `json:"role"`
	Kind MessageKind `json:"kind"`

	text string
	file File
}

type File struct {
	Data []byte
	Name string
}

func (m Message) IsText() bool {
	return m.Kind == MessageKindText
}

func (m Message) IsFile() bool {
	return m.Kind == MessageKindFile
}

func (m Message) Text() string {
	if m.Kind != MessageKindText {
		return ""
	}
	return m.text
}

func (m Message) File() File {
	if m.Kind != MessageKindFile {
		return File{}
	}
	return m.file
}

func UserTextMessage(text string) Message {
	return Message{
		Role: RoleUser,
		Kind: MessageKindText,
		text: text,
	}
}

func UserFileMessage(file File) Message {
	return Message{
		Role: RoleUser,
		Kind: MessageKindFile,
		file: file,
	}
}

func AssistantTextMessage(content string) Message {
	return Message{
		Role: RoleAssistant,
		Kind: MessageKindText,
		text: content,
	}
}

func SystemMessage(text string) Message {
	return Message{
		Role: RoleSystem,
		Kind: MessageKindText,
		text: text,
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
