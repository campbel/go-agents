# go-agents

A Go library that implements an Agent interface around the OpenAI Go SDK, enabling chat completions with tool execution capabilities.

## Features

- Chat completions with streaming responses
- Tool execution with automatic conversation management
- Support for any OpenAI-compatible API endpoint
- Type-safe message and tool interfaces
- Configurable iteration limits for tool conversations

## Installation

```bash
go get github.com/campbel/go-agents
```

## Quick Start

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
        []agent.Tool{}, // Add your tools here
    )
    
    // Start a conversation
    messages := []agent.Message{
        agent.UserMessage("Hello, how are you?"),
    }
    
    responseChan, err := agent.ChatCompletionWithTools(
        context.Background(),
        messages,
        []agent.Tool{}, // Tools available for this conversation
    )
    if err != nil {
        panic(err)
    }
    
    // Read responses
    for response := range responseChan {
        if response.Error != nil {
            fmt.Printf("Error: %v\n", response.Error)
            continue
        }
        fmt.Print(response.Content)
    }
}
```

## Creating Tools

Implement the `Tool` interface to create custom tools:

```go
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
    location := input["location"].(string)
    // Implement weather lookup logic
    return fmt.Sprintf("Weather in %s: Sunny, 72°F", location), nil
}
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
