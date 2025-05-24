package agent

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgent(t *testing.T) {
	testAgent := NewAgent(os.Getenv("ANTHROPIC_API_KEY"), "https://api.anthropic.com/v1/", "claude-sonnet-4-20250514", []Tool{})

	responseChan, err := testAgent.ChatCompletionWithTools(t.Context(), []Message{
		UserMessage("What is the weather in Tokyo?"),
	}, []Tool{})

	assert.Nil(t, err)

	var messages []string
	for response := range responseChan {
		assert.Nil(t, response.Error)
		messages = append(messages, response.Content)
	}

	assert.NotEmpty(t, messages)
}
