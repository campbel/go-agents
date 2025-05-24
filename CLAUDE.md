# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

This project uses the `bolt` CLI for development:
- `bolt up` - Setup development environment with Go 1.24.3
- `go test ./...` - Run all tests
- `go build` - Build the agent package
- `go mod tidy` - Clean up dependencies

## Architecture

This is a Go library that implements an Agent interface around the OpenAI Go SDK. The core architecture consists of:

### Key Components

- **Agent** (`agent.go`): Main implementation that wraps OpenAI client with tool execution capabilities
- **Message** (`message.go`): Message types for chat conversations (User/Assistant roles)
- **Tool** (`tools.go`): Interface for implementing custom tools that agents can execute
- **Response** (`message.go`): Response structure containing content and error information

### Agent Flow

The `ChatCompletionWithTools` method implements a conversation loop:
1. Converts messages to OpenAI format
2. Makes chat completion request with available tools
3. Executes any tool calls and adds results to conversation
4. Continues iterating until no more tools are called (max 100 iterations)
5. Streams responses through a channel

### Tool Integration

Tools must implement the `Tool` interface with:
- `Name()` - tool identifier
- `Description()` - what the tool does
- `Parameters()` - JSON schema for tool inputs
- `Execute()` - tool execution logic

The agent looks up tools by name during execution and marshals/unmarshals tool arguments automatically.

### Testing

Tests use the Anthropic API (claude-sonnet-4-20250514 model) and require `ANTHROPIC_API_KEY` environment variable. The agent supports any OpenAI-compatible API endpoint.