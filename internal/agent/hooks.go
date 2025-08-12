package agent

import (
	"context"

	"github.com/openai/openai-go"
)

type Hooks struct {
	onAgentContent   []func(ctx context.Context, response *openai.ChatCompletion)
	onBeforeToolCall []func(ctx context.Context, toolCall *openai.ChatCompletionMessageToolCall)
	onAfterToolCall  []func(ctx context.Context, response *openai.ChatCompletion)
	onSuggestion     []func(ctx context.Context, suggestion string)
}

func (h *Hooks) AddOnAgentContent(hook func(ctx context.Context, response *openai.ChatCompletion)) {
	h.onAgentContent = append(h.onAgentContent, hook)
}

func (h *Hooks) AddBeforeToolCall(hook func(ctx context.Context, toolCall *openai.ChatCompletionMessageToolCall)) {
	h.onBeforeToolCall = append(h.onBeforeToolCall, hook)
}

func (h *Hooks) AddAfterToolCall(hook func(ctx context.Context, response *openai.ChatCompletion)) {
	h.onAfterToolCall = append(h.onAfterToolCall, hook)
}

func (h *Hooks) AddOnSuggestion(hook func(ctx context.Context, suggestion string)) {
	h.onSuggestion = append(h.onSuggestion, hook)
}
