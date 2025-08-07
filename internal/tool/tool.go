package tool

import "context"

type Tool interface {
	Name() string
	Description() string
	Params() map[string]any
	Call(ctx context.Context, input string) (string, error)
}
