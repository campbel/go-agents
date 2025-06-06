package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

// AgentOption is a functional option for configuring an Agent
type AgentOption func(*Agent)

// WithSystemPrompt sets the system prompt for the agent
func WithSystemPrompt(prompt string) AgentOption {
	return func(a *Agent) {
		a.systemPrompt = prompt
	}
}

// WithInstructions sets the instructions for the agent
func WithInstructions(instructions string) AgentOption {
	return func(a *Agent) {
		a.instructions = instructions
	}
}

// WithTools sets the tools for the agent
func WithTools(tools []Tool) AgentOption {
	return func(a *Agent) {
		a.tools = tools
	}
}

// WithMaxIterations sets the maximum number of iterations for the agent
func WithMaxIterations(max int) AgentOption {
	return func(a *Agent) {
		a.maxIterations = max
	}
}

// Agent implements the Agent interface using the OpenAI-compatible API
type Agent struct {
	client        openai.Client
	model         string
	tools         []Tool
	maxIterations int
	systemPrompt  string
	instructions  string
}

// NewAgent creates a new Agent with the given API key, base URL, and model
func NewAgent(apiKey string, baseURL string, model string, opts ...AgentOption) *Agent {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	// Create agent with defaults
	agent := &Agent{
		client:        client,
		model:         model,
		tools:         []Tool{},
		maxIterations: 100,
		systemPrompt:  "",
		instructions:  "",
	}

	// Apply options
	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

// NewAgentWithClient creates a new Agent with an existing OpenAI client
func NewAgentWithClient(client openai.Client, model string, opts ...AgentOption) *Agent {
	// Create agent with defaults
	agent := &Agent{
		client:        client,
		model:         model,
		tools:         []Tool{},
		maxIterations: 100,
		systemPrompt:  "",
		instructions:  "",
	}

	// Apply options
	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

func (agent *Agent) ChatCompletion(
	ctx context.Context,
	messages []Message,
) (Completion, error) {
	responseChan, err := agent.StreamChatCompletion(ctx, messages)
	if err != nil {
		return Completion{}, err
	}

	var completion Completion

	for response := range responseChan {
		completion.Responses = append(completion.Responses, response)
		if response.IsUsageResponse() {
			usage := response.Usage()
			completion.Usage.PromptTokens += usage.PromptTokens
			completion.Usage.CompletionTokens += usage.CompletionTokens
			completion.Usage.TotalTokens += usage.TotalTokens
		}
		if response.IsContentResponse() {
			completion.Messages = append(completion.Messages, response.Content())
		}
		if response.IsErrorResponse() {
			return Completion{}, response.Error()
		}
	}

	return completion, nil
}

// StreamChatCompletion implements the Agent interface
func (agent *Agent) StreamChatCompletion(
	ctx context.Context,
	messages []Message,
) (<-chan Response, error) {
	responseChan := make(chan Response)

	// Convert the messages to OpenAI format and inject system prompt and instructions
	chatMessages := agent.buildMessages(messages)

	// Initialize tools params
	var openAITools []openai.ChatCompletionToolParam
	for _, tool := range agent.tools {
		openAITools = append(openAITools, openai.ChatCompletionToolParam{
			Type: "function",
			Function: openai.FunctionDefinitionParam{
				Name:        tool.Name(),
				Description: openai.String(tool.Description()),
				Parameters:  convertParameters(tool.Parameters()),
			},
		})
	}

	// Create params for the completion
	params := openai.ChatCompletionNewParams{
		Messages: chatMessages,
		Model:    openai.ChatModel(agent.model),
		Tools:    openAITools,
	}

	go func() {
		defer close(responseChan)
		err := func() error {
			for range agent.maxIterations {
				// Start streaming completion
				response, err := agent.client.Chat.Completions.New(ctx, params)
				if err != nil {
					return err
				}

				responseChan <- NewUsageResponse(Usage{
					PromptTokens:     response.Usage.PromptTokens,
					CompletionTokens: response.Usage.CompletionTokens,
					TotalTokens:      response.Usage.TotalTokens,
				})

				// Check if there are tool calls
				hasToolCalls := len(response.Choices[0].Message.ToolCalls) > 0

				// Add the AI message to our conversation only if it has content or tool calls
				if response.Choices[0].Message.Content != "" || hasToolCalls {
					params.Messages = append(params.Messages, response.Choices[0].Message.ToParam())
				}

				// Send content to response channel if present
				if response.Choices[0].Message.Content != "" {
					responseChan <- NewContentResponse(response.Choices[0].Message.Content)
				}

				// Handle any tool calls
				if hasToolCalls {
					for _, toolCall := range response.Choices[0].Message.ToolCalls {
						// TODO: add a lookup map
						var tool Tool
						for _, t := range agent.tools {
							if t.Name() == toolCall.Function.Name {
								tool = t
								break
							}
						}

						// Execute the tool using the tool executor
						var args map[string]any
						if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
							return err
						}
						toolResult, err := tool.Execute(ctx, args)
						if err != nil {
							return err
						}

						switch v := toolResult.(type) {
						case string:
							params.Messages = append(params.Messages, openai.ToolMessage(v, toolCall.ID))
						case map[string]any, []any:
							data, err := json.Marshal(v)
							if err != nil {
								return err
							}
							params.Messages = append(params.Messages, openai.ToolMessage(string(data), toolCall.ID))
						default:
							data, err := json.Marshal(v)
							if err != nil {
								return err
							}
							params.Messages = append(params.Messages, openai.ToolMessage(string(data), toolCall.ID))
						}
					}
				} else {
					// No tool calls, exit the loop
					break
				}
			}
			return nil
		}()
		if err != nil {
			responseChan <- NewErrorResponse(err)
		}
	}()

	return responseChan, nil
}

// convertMessages converts models.Message to OpenAI format
func convertMessages(messages []Message) []openai.ChatCompletionMessageParamUnion {
	var chatMessages []openai.ChatCompletionMessageParamUnion
	for _, msg := range messages {
		switch msg.Role() {
		case RoleSystem:
			chatMessages = append(chatMessages, openai.SystemMessage(msg.Text()))
		case RoleAssistant:
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Text()))
		case RoleUser:
			switch msg.Kind() {
			case MessageKindText:
				chatMessages = append(chatMessages, openai.UserMessage(msg.Text()))
			case MessageKindFile:
				base64Data := base64.StdEncoding.EncodeToString(msg.File().Data)
				chatMessages = append(chatMessages, openai.ChatCompletionMessageParamUnion{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							OfArrayOfContentParts: []openai.ChatCompletionContentPartUnionParam{
								{
									OfFile: &openai.ChatCompletionContentPartFileParam{
										File: openai.ChatCompletionContentPartFileFileParam{
											FileData: openai.String(base64Data),
											Filename: openai.String(msg.File().Name),
										},
									},
								},
							},
						},
					},
				})
			case MessageKindImage:
				base64Data := base64.StdEncoding.EncodeToString(msg.Image().Data)
				chatMessages = append(chatMessages, openai.ChatCompletionMessageParamUnion{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							OfArrayOfContentParts: []openai.ChatCompletionContentPartUnionParam{
								{
									OfImageURL: &openai.ChatCompletionContentPartImageParam{
										ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
											URL: "data:image/png;base64," + base64Data,
										},
									},
								},
							},
						},
					},
				})
			}
		default:
			chatMessages = append(chatMessages, openai.UserMessage(msg.Text()))
		}
	}
	return chatMessages
}

// buildMessages converts messages and injects system prompt and instructions
func (agent *Agent) buildMessages(messages []Message) []openai.ChatCompletionMessageParamUnion {
	var chatMessages []openai.ChatCompletionMessageParamUnion

	// Add system prompt if provided
	if agent.systemPrompt != "" {
		chatMessages = append(chatMessages, openai.SystemMessage(agent.systemPrompt))
	}

	// Add instructions as first user message if provided
	if agent.instructions != "" {
		chatMessages = append(chatMessages, openai.UserMessage(agent.instructions))
	}

	// Convert and append the provided messages
	userMessages := convertMessages(messages)
	chatMessages = append(chatMessages, userMessages...)

	return chatMessages
}

func convertParameters(parameters Parameters) shared.FunctionParameters {
	return shared.FunctionParameters{
		"type":       "object",
		"properties": parameters.Properties,
		"required":   parameters.Required,
	}
}
