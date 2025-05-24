# go-agents

A Go library that implements an Agent interface around the OpenAI Go SDK, enabling chat completions with tool execution capabilities.

## Features

- Chat completions with both streaming and non-streaming responses
- Tool execution with automatic conversation management
- Support for any OpenAI-compatible API endpoint
- Type-safe message and tool interfaces
- Configurable iteration limits for tool conversations
- Functional options pattern for agent configuration
- System prompts and instructions support
- Image and file message support

## Installation

```bash
go get github.com/campbel/go-agents
```

## Quick Start

### Basic Usage with Streaming

```go
package main

import (
    "context"
    "fmt"
    "os"
    
    "github.com/campbel/go-agents"
)

func main() {
    // Create a new agent
    agent := agent.NewAgent(
        os.Getenv("OPENAI_API_KEY"),
        "https://api.openai.com/v1",
        "gpt-4",
    )
    
    // Start a conversation with streaming
    messages := []agent.Message{
        agent.UserTextMessage("Hello, how are you?"),
    }
    
    responseChan, err := agent.StreamChatCompletion(
        context.Background(),
        messages,
    )
    if err != nil {
        panic(err)
    }
    
    // Read streaming responses
    for response := range responseChan {
        if response.IsErrorResponse() {
            fmt.Printf("Error: %v\n", response.Error())
            continue
        }
        if response.IsContentResponse() {
            fmt.Print(response.Content())
        }
    }
}
```

### Non-Streaming Usage

```go
// Get complete response at once
completion, err := agent.ChatCompletion(
    context.Background(),
    []agent.Message{
        agent.UserTextMessage("What is 2+2?"),
    },
)
if err != nil {
    panic(err)
}

// Access all messages and usage information
for _, message := range completion.Messages {
    fmt.Println(message)
}
fmt.Printf("Total tokens used: %d\n", completion.Usage.TotalTokens)
```

## Configuration Options

The agent supports functional options for flexible configuration:

```go
// Create agent with system prompt and tools
agent := agent.NewAgent(
    os.Getenv("OPENAI_API_KEY"),
    "https://api.openai.com/v1",
    "gpt-4",
    agent.WithSystemPrompt("You are a helpful assistant that always provides detailed explanations."),
    agent.WithInstructions("Please be concise in your responses."),
    agent.WithTools([]agent.Tool{weatherTool, calculatorTool}),
    agent.WithMaxIterations(50),
)
```

Available options:
- `WithSystemPrompt(string)` - Set a system prompt for the agent
- `WithInstructions(string)` - Add instructions as the first user message
- `WithTools([]Tool)` - Configure tools available to the agent
- `WithMaxIterations(int)` - Set maximum tool execution iterations (default: 100)

## Creating Tools

Implement the `Tool` interface to create custom tools:

```go
import (
    "context"
    "fmt"
    
    "github.com/campbel/go-agents"
)

type WeatherTool struct{}

func (w WeatherTool) Name() string {
    return "get_weather"
}

func (w WeatherTool) Description() string {
    return "Get current weather for a location"
}

func (w WeatherTool) Parameters() agent.Parameters {
    return agent.Parameters{
        Properties: map[string]any{
            "location": map[string]any{
                "type":        "string",
                "description": "The city name",
            },
        },
        Required: []string{"location"},
    }
}

func (w WeatherTool) Execute(ctx context.Context, input map[string]any) (any, error) {
    location, ok := input["location"].(string)
    if !ok {
        return nil, fmt.Errorf("location must be a string")
    }
    
    // Implement weather lookup logic
    return map[string]any{
        "location":    location,
        "temperature": "72°F",
        "condition":   "Sunny",
        "humidity":    "65%",
    }, nil
}

// Use the tool with an agent
func main() {
    weatherTool := WeatherTool{}
    
    agent := agent.NewAgent(
        os.Getenv("OPENAI_API_KEY"),
        "https://api.openai.com/v1",
        "gpt-4",
        agent.WithTools([]agent.Tool{weatherTool}),
    )
    
    // The agent can now use the weather tool automatically
    completion, err := agent.ChatCompletion(context.Background(), []agent.Message{
        agent.UserTextMessage("What's the weather like in New York?"),
    })
    // ... handle response
}
```

## Advanced Features

### Image and File Support

The agent supports sending images and files as part of the conversation:

```go
import (
    "os"
)

// Send an image
imageData, err := os.ReadFile("image.png")
if err != nil {
    panic(err)
}

messages := []agent.Message{
    agent.UserImageMessage(agent.Image{
        Data: imageData,
        Name: "image.png",
    }),
    agent.UserTextMessage("What do you see in this image?"),
}

// Send a file
fileData, err := os.ReadFile("document.pdf")
if err != nil {
    panic(err)
}

messages = append(messages, agent.UserFileMessage(agent.File{
    Data: fileData,
    Name: "document.pdf",
}))
```

### Message Types

Different message types are available:

```go
// User messages
userText := agent.UserTextMessage("Hello")
userImage := agent.UserImageMessage(agent.Image{Data: imageData, Name: "image.png"})
userFile := agent.UserFileMessage(agent.File{Data: fileData, Name: "file.txt"})

// Assistant and system messages
assistant := agent.AssistantTextMessage("Hello back!")
system := agent.SystemMessage("You are a helpful assistant")
```

### Error Handling and Usage Tracking

```go
completion, err := agent.ChatCompletion(ctx, messages)
if err != nil {
    log.Fatal(err)
}

// Check individual responses for errors
for _, response := range completion.Responses {
    if response.IsErrorResponse() {
        fmt.Printf("Error in response: %v\n", response.Error())
    }
}

// Track token usage
fmt.Printf("Prompt tokens: %d\n", completion.Usage.PromptTokens)
fmt.Printf("Completion tokens: %d\n", completion.Usage.CompletionTokens)
fmt.Printf("Total tokens: %d\n", completion.Usage.TotalTokens)
```

## API Compatibility

This library works with any OpenAI-compatible API including:
- OpenAI GPT models
- Anthropic Claude (via compatibility layer)
- Local models via ollama, vllm, etc.
- Azure OpenAI Service

## Development

This project uses the `bolt` CLI for development:

```bash
# Setup environment
bolt up

# Run tests
go test ./...

# Build
go build
```

## License

MIT
