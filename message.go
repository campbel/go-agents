package agent

type MessageKind string

const (
	MessageKindText  MessageKind = "text"
	MessageKindFile  MessageKind = "file"
	MessageKindImage MessageKind = "image"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

type Message struct {
	role Role
	kind MessageKind

	text  string
	file  File
	image Image
}

type File struct {
	Data []byte
	Name string
}

type Image struct {
	Data []byte
	Name string
}

func (m Message) Role() Role {
	return m.role
}

func (m Message) Kind() MessageKind {
	return m.kind
}

func (m Message) IsText() bool {
	return m.kind == MessageKindText
}

func (m Message) IsFile() bool {
	return m.kind == MessageKindFile
}

func (m Message) IsImage() bool {
	return m.kind == MessageKindImage
}

func (m Message) Text() string {
	if m.kind != MessageKindText {
		return ""
	}
	return m.text
}

func (m Message) File() File {
	if m.kind != MessageKindFile {
		return File{}
	}
	return m.file
}

func (m Message) Image() Image {
	if m.kind != MessageKindImage {
		return Image{}
	}
	return m.image
}

func UserTextMessage(text string) Message {
	return Message{
		role: RoleUser,
		kind: MessageKindText,
		text: text,
	}
}

func UserFileMessage(file File) Message {
	return Message{
		role: RoleUser,
		kind: MessageKindFile,
		file: file,
	}
}

func UserImageMessage(image Image) Message {
	return Message{
		role:  RoleUser,
		kind:  MessageKindImage,
		image: image,
	}
}

func AssistantTextMessage(content string) Message {
	return Message{
		role: RoleAssistant,
		kind: MessageKindText,
		text: content,
	}
}

func SystemMessage(text string) Message {
	return Message{
		role: RoleSystem,
		kind: MessageKindText,
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
