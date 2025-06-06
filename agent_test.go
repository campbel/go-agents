package agent

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTool implements the Tool interface for testing
type MockTool struct {
	name        string
	description string
	parameters  Parameters
	executeFunc func(ctx context.Context, input map[string]any) (any, error)
}

func (m MockTool) Name() string {
	return m.name
}

func (m MockTool) Description() string {
	return m.description
}

func (m MockTool) Parameters() Parameters {
	return m.parameters
}

func (m MockTool) Execute(ctx context.Context, input map[string]any) (any, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, input)
	}
	return "mock result", nil
}

func TestAgent(t *testing.T) {
	// Create a weather tool
	weatherTool := MockTool{
		name:        "get_weather",
		description: "Get current weather for a location",
		parameters: Parameters{
			Properties: map[string]any{
				"location": map[string]any{
					"type":        "string",
					"description": "The city name",
				},
			},
			Required: []string{"location"},
		},
		executeFunc: func(ctx context.Context, input map[string]any) (any, error) {
			location, ok := input["location"].(string)
			if !ok {
				return nil, assert.AnError
			}
			return map[string]any{
				"location":    location,
				"temperature": "22°C",
				"condition":   "Sunny",
				"humidity":    "65%",
			}, nil
		},
	}

	testAgent := NewAgent(os.Getenv("ANTHROPIC_API_KEY"), "https://api.anthropic.com/v1/", "claude-sonnet-4-20250514", WithTools([]Tool{weatherTool}))

	responseChan, err := testAgent.StreamChatCompletion(context.Background(), []Message{
		UserTextMessage("What is the weather in Tokyo? Use the get_weather tool."),
	})

	assert.Nil(t, err)

	var messages []string
	for response := range responseChan {
		if response.IsErrorResponse() {
			t.Fatalf("Unexpected error: %v", response.Error())
		}
		if response.IsContentResponse() {
			messages = append(messages, response.Content())
		}
	}

	assert.NotEmpty(t, messages, "Expected at least one non-empty response")

	// Check that the weather tool was used and we got some weather information
	allContent := ""
	for _, msg := range messages {
		allContent += msg + " "
	}
	assert.Contains(t, allContent, "Tokyo", "Response should mention Tokyo")
	assert.Contains(t, allContent, "22°C", "Response should contain the temperature from the tool")
}

func TestNewAgent(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		baseURL string
		model   string
		tools   []Tool
	}{
		{
			name:    "basic agent creation",
			apiKey:  "test-key",
			baseURL: "https://api.openai.com/v1",
			model:   "gpt-4",
			tools:   []Tool{},
		},
		{
			name:    "agent with tools",
			apiKey:  "test-key",
			baseURL: "https://api.openai.com/v1",
			model:   "gpt-3.5-turbo",
			tools: []Tool{
				MockTool{name: "test_tool", description: "Test tool"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewAgent(tt.apiKey, tt.baseURL, tt.model, WithTools(tt.tools))
			require.NotNil(t, agent)
			assert.Equal(t, tt.model, agent.model)
			assert.Equal(t, len(tt.tools), len(agent.tools))
			assert.Equal(t, 100, agent.maxIterations)
		})
	}
}

func TestUserMessage(t *testing.T) {
	content := "Hello, world!"
	msg := UserTextMessage(content)

	assert.Equal(t, RoleUser, msg.Role())
	assert.Equal(t, content, msg.Text())
}

func TestAssistantMessage(t *testing.T) {
	content := "Hello back!"
	msg := AssistantTextMessage(content)

	assert.Equal(t, RoleAssistant, msg.Role())
	assert.Equal(t, content, msg.Text())
}

func TestConvertMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
		expected int
	}{
		{
			name:     "empty messages",
			messages: []Message{},
			expected: 0,
		},
		{
			name: "single user message",
			messages: []Message{
				UserTextMessage("Hello"),
			},
			expected: 1,
		},
		{
			name: "mixed message types",
			messages: []Message{
				SystemMessage("You are a helpful assistant"),
				UserTextMessage("Hello"),
				AssistantTextMessage("Hi there!"),
			},
			expected: 3,
		},
		{
			name: "unknown role defaults to user",
			messages: []Message{
				UserTextMessage("Test"),
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertMessages(tt.messages)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestConvertParameters(t *testing.T) {
	params := Parameters{
		Properties: map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "The name parameter",
			},
			"age": map[string]any{
				"type":        "integer",
				"description": "The age parameter",
			},
		},
		Required: []string{"name"},
	}

	result := convertParameters(params)

	assert.NotNil(t, result)
	assert.Contains(t, result, "properties")
	assert.Contains(t, result, "required")
	assert.Equal(t, params.Properties, result["properties"])
	assert.Equal(t, params.Required, result["required"])
}

func TestMockTool(t *testing.T) {
	tool := MockTool{
		name:        "test_tool",
		description: "A test tool",
		parameters: Parameters{
			Properties: map[string]any{
				"input": map[string]any{
					"type": "string",
				},
			},
			Required: []string{"input"},
		},
		executeFunc: func(ctx context.Context, input map[string]any) (any, error) {
			return "custom result", nil
		},
	}

	assert.Equal(t, "test_tool", tool.Name())
	assert.Equal(t, "A test tool", tool.Description())
	assert.Equal(t, []string{"input"}, tool.Parameters().Required)

	result, err := tool.Execute(context.Background(), map[string]any{"input": "test"})
	assert.NoError(t, err)
	assert.Equal(t, "custom result", result)
}

func TestMockToolDefaultExecution(t *testing.T) {
	tool := MockTool{
		name:        "simple_tool",
		description: "A simple tool",
	}

	result, err := tool.Execute(context.Background(), map[string]any{})
	assert.NoError(t, err)
	assert.Equal(t, "mock result", result)
}

func TestAgentWithImage(t *testing.T) {
	// Skip if no API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	// Read the test image
	imageData, err := os.ReadFile("testdata/claude.png")
	require.NoError(t, err, "Failed to read test image")

	testAgent := NewAgent(os.Getenv("ANTHROPIC_API_KEY"), "https://api.anthropic.com/v1/", "claude-sonnet-4-20250514")

	responseChan, err := testAgent.StreamChatCompletion(context.Background(), []Message{
		UserImageMessage(Image{
			Data: imageData,
			Name: "claude.png",
		}),
		UserTextMessage("What do you see in this image? Please mention if you see any logos or text."),
	})

	assert.Nil(t, err)

	var messages []string
	for response := range responseChan {
		if response.IsErrorResponse() {
			t.Fatalf("Unexpected error: %v", response.Error())
		}
		if response.IsContentResponse() {
			messages = append(messages, response.Content())
		}
	}

	assert.NotEmpty(t, messages, "Expected at least one non-empty response")

	// Check that the model recognized something in the image
	allContent := strings.ToLower(strings.Join(messages, " "))

	// Look for common keywords that might appear when describing the Claude logo/image
	hasImageDescription := strings.Contains(allContent, "logo") ||
		strings.Contains(allContent, "text") ||
		strings.Contains(allContent, "image") ||
		strings.Contains(allContent, "claude") ||
		strings.Contains(allContent, "anthropic") ||
		strings.Contains(allContent, "see")

	assert.True(t, hasImageDescription, "Response should contain image description keywords. Got: %s", allContent)
}

func TestAgentWithSystemPromptAndInstructions(t *testing.T) {
	// Skip if no API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	systemPrompt := "You are a helpful AI assistant that always ends responses with 'SYSTEM_TEST_MARKER'."
	instructions := "Please be very brief in your responses and include the word 'INSTRUCTION_TEST_MARKER' somewhere."

	testAgent := NewAgent(
		os.Getenv("ANTHROPIC_API_KEY"),
		"https://api.anthropic.com/v1/",
		"claude-sonnet-4-20250514",
		WithSystemPrompt(systemPrompt),
		WithInstructions(instructions),
	)

	responseChan, err := testAgent.StreamChatCompletion(context.Background(), []Message{
		UserTextMessage("Hello, can you tell me what 2+2 equals?"),
	})

	assert.Nil(t, err)

	var messages []string
	for response := range responseChan {
		if response.IsErrorResponse() {
			t.Fatalf("Unexpected error: %v", response.Error())
		}
		if response.IsContentResponse() {
			messages = append(messages, response.Content())
		}
	}

	assert.NotEmpty(t, messages, "Expected at least one non-empty response")

	allContent := strings.Join(messages, " ")

	// Check that both the system prompt and instructions influenced the response
	assert.Contains(t, allContent, "SYSTEM_TEST_MARKER", "Response should contain system prompt marker")
	assert.Contains(t, allContent, "INSTRUCTION_TEST_MARKER", "Response should contain instruction marker")
	assert.Contains(t, allContent, "4", "Response should contain the answer to 2+2")
}

func TestChatCompletion(t *testing.T) {
	// Skip if no API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	testAgent := NewAgent(os.Getenv("ANTHROPIC_API_KEY"), "https://api.anthropic.com/v1/", "claude-sonnet-4-20250514")

	completion, err := testAgent.ChatCompletion(context.Background(), []Message{
		UserTextMessage("What is 2+2? Please respond with just the number."),
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, completion.Messages, "Expected at least one message in completion")
	assert.NotEmpty(t, completion.Responses, "Expected at least one response in completion")

	// Check that we have usage information
	assert.Greater(t, completion.Usage.TotalTokens, int64(0), "Expected non-zero total tokens")
	assert.Greater(t, completion.Usage.PromptTokens, int64(0), "Expected non-zero prompt tokens")
	assert.Greater(t, completion.Usage.CompletionTokens, int64(0), "Expected non-zero completion tokens")

	// Check that the completion contains the expected answer
	allContent := strings.Join(completion.Messages, " ")
	assert.Contains(t, allContent, "4", "Response should contain the answer to 2+2")
}

func TestNewAgentOptionsPattern(t *testing.T) {
	// Test simple agent with no options
	agent1 := NewAgent("test-key", "https://api.example.com", "test-model")
	assert.Equal(t, "test-model", agent1.model)
	assert.Equal(t, []Tool{}, agent1.tools)
	assert.Equal(t, 100, agent1.maxIterations)
	assert.Equal(t, "", agent1.systemPrompt)
	assert.Equal(t, "", agent1.instructions)

	// Test agent with single option
	agent2 := NewAgent("test-key", "https://api.example.com", "test-model",
		WithSystemPrompt("You are helpful"))
	assert.Equal(t, "You are helpful", agent2.systemPrompt)

	// Test agent with multiple options
	tools := []Tool{MockTool{name: "test_tool"}}
	agent3 := NewAgent("test-key", "https://api.example.com", "test-model",
		WithSystemPrompt("You are helpful"),
		WithInstructions("Be concise"),
		WithTools(tools),
		WithMaxIterations(50),
	)
	assert.Equal(t, "You are helpful", agent3.systemPrompt)
	assert.Equal(t, "Be concise", agent3.instructions)
	assert.Equal(t, tools, agent3.tools)
	assert.Equal(t, 50, agent3.maxIterations)
}
