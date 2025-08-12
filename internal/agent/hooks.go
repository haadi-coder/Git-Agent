package agent

import (
	"context"

	"github.com/openai/openai-go"
)

type onIntermidiateStep func(ctx context.Context, response *openai.ChatCompletion)
type OnBeforeCallTool func(ctx context.Context, toolCall *openai.ChatCompletionMessageToolCall)
type OnAfterCallTool func(ctx context.Context, response *openai.ChatCompletion)
type onSuggestion func(ctx context.Context, suggestion string)

type Hooks struct {
	onIntermidiateStep []onIntermidiateStep
	onBeforeCallTool   []OnBeforeCallTool
	onAfterCallTool    []OnAfterCallTool
	onSuggestion       []onSuggestion
}

func (h *Hooks) AddOnIntermidiateStep(hook func(ctx context.Context, response *openai.ChatCompletion)) {
	h.onIntermidiateStep = append(h.onIntermidiateStep, hook)
}

func (h *Hooks) AddBeforeToolCall(hook func(ctx context.Context, toolCall *openai.ChatCompletionMessageToolCall)) {
	h.onBeforeCallTool = append(h.onBeforeCallTool, hook)
}

func (h *Hooks) AddAfterToolCall(hook func(ctx context.Context, response *openai.ChatCompletion)) {
	h.onAfterCallTool = append(h.onAfterCallTool, hook)
}

func (h *Hooks) AddOnSuggestion(hook func(ctx context.Context, suggestion string)) {
	h.onSuggestion = append(h.onSuggestion, hook)
}

func (h *Hooks) handleIntermidiateStep(ctx context.Context, response *openai.ChatCompletion) {
	if len(h.onIntermidiateStep) == 0 {
		return
	}

	for _, hook := range h.onIntermidiateStep {
		hook(ctx, response)
	}
}

func (h *Hooks) handleBeforeToolCall(ctx context.Context, toolCall *openai.ChatCompletionMessageToolCall) {
	if len(h.onBeforeCallTool) == 0 {
		return
	}

	for _, hook := range h.onBeforeCallTool {
		hook(ctx, toolCall)
	}
}

func (h *Hooks) handleAfterToolCall(ctx context.Context, response *openai.ChatCompletion) {
	if len(h.onAfterCallTool) == 0 {
		return
	}

	for _, hook := range h.onAfterCallTool {
		hook(ctx, response)
	}
}

func (h *Hooks) handleSuggestion(ctx context.Context, suggestion string) {
	if len(h.onSuggestion) == 0 {
		return
	}

	for _, hook := range h.onSuggestion {
		hook(ctx, suggestion)
	}
}
