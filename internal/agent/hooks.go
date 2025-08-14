package agent

import (
	"context"

	"github.com/openai/openai-go"
)

type onIntermidiateStep func(ctx context.Context, response *openai.ChatCompletion)
type onCallTool func(ctx context.Context, toolCall *openai.ChatCompletionMessageToolCall)
type onSuggestion func(ctx context.Context, suggestion string, history *[]openai.ChatCompletionMessageParamUnion)

type Hooks struct {
	onIntermidiateStep      []onIntermidiateStep
	onAfterIntermidiateStep []onIntermidiateStep
	onBeforeCallTool        []onCallTool
	onSuggestion            []onSuggestion
}

func (h *Hooks) AddOnIntermidiateStep(hook onIntermidiateStep) {
	h.onIntermidiateStep = append(h.onIntermidiateStep, hook)
}

func (h *Hooks) AddOnAfterIntermidiateStep(hook onIntermidiateStep) {
	h.onAfterIntermidiateStep = append(h.onAfterIntermidiateStep, hook)
}

func (h *Hooks) AddBeforeCallTool(hook onCallTool) {
	h.onBeforeCallTool = append(h.onBeforeCallTool, hook)
}

func (h *Hooks) AddOnSuggestion(hook onSuggestion) {
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

func (h *Hooks) handleAfterIntermidiateStep(ctx context.Context, response *openai.ChatCompletion) {
	if len(h.onAfterIntermidiateStep) == 0 {
		return
	}

	for _, hook := range h.onAfterIntermidiateStep {
		hook(ctx, response)
	}
}

func (h *Hooks) handleBeforeCallTool(ctx context.Context, toolCall *openai.ChatCompletionMessageToolCall) {
	if len(h.onBeforeCallTool) == 0 {
		return
	}

	for _, hook := range h.onBeforeCallTool {
		hook(ctx, toolCall)
	}
}

func (h *Hooks) handleSuggestion(ctx context.Context, suggestion string, history *[]openai.ChatCompletionMessageParamUnion) {
	if len(h.onSuggestion) == 0 {
		return
	}

	for _, hook := range h.onSuggestion {
		hook(ctx, suggestion, history)
	}
}
