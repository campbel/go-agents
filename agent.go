package agent

import (
	"context"
	"encoding/json"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

// Agent implements the Agent interface using the OpenAI-compatible API
type Agent struct {
	client        openai.Client
	model         string
	tools         []Tool
	maxIterations int
}

// NewAgent creates a new Agent
func NewAgent(apiKey string, baseURL string, model string, tools []Tool) *Agent {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)
	return &Agent{
		client:        client,
		model:         model,
		tools:         tools,
		maxIterations: 100,
	}
}

func NewAgentWithClient(client openai.Client, model string, tools []Tool) *Agent {
	return &Agent{
		client:        client,
		model:         model,
		tools:         tools,
		maxIterations: 100,
	}
}

// ChatCompletionWithTools implements the Agent interface with tools support
func (agent *Agent) ChatCompletionWithTools(
	ctx context.Context,
	messages []Message,
) (<-chan Response, error) {
	responseChan := make(chan Response)

	// Convert the messages to OpenAI format
	chatMessages := convertMessages(messages)

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
						case any, map[string]any, []any:
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
		switch msg.Role {
		case "system":
			chatMessages = append(chatMessages, openai.SystemMessage(msg.Content))
		case "assistant":
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
		case "user":
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		default:
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		}
	}
	return chatMessages
}

func convertParameters(parameters Parameters) shared.FunctionParameters {
	return shared.FunctionParameters{
		"type":       "object",
		"properties": parameters.Properties,
		"required":   parameters.Required,
	}
}
