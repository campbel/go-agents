package agent

import "context"

type Tool interface {
	Name() string
	Description() string
	Parameters() Parameters
	Execute(ctx context.Context, input map[string]any) (any, error)
}

type Parameters struct {
	Properties map[string]any `json:"properties"`
	Required   []string       `json:"required"`
}
