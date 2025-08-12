package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/haadi-coder/Git-Agent/internal/llm"
	"github.com/haadi-coder/Git-Agent/internal/tool"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

type Agent struct {
	LLM            *llm.OpenRouter
	Tools          []tool.Tool
	SystemPrompt   string
	ResponseFormat openai.ChatCompletionNewParamsResponseFormatUnion
	Hooks          *Hooks
}

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

func (a *Agent) Run(ctx context.Context) (string, error) {
	toolDefs := a.toolDefs()
	history := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(a.SystemPrompt),
	}

	for {
		response, err := a.LLM.CreateChatCompletion(ctx, openai.ChatCompletionNewParams{
			Messages:       history,
			Tools:          toolDefs,
			ResponseFormat: a.ResponseFormat,
		})
		if err != nil {
			return "", err
		}

		message := response.Choices[0].Message

		if message.Content != "" {
			for _, hook := range a.Hooks.onAgentContent {
				hook(ctx, response)
			}
		}

		if len(message.ToolCalls) == 0 {
			return a.handleResponse(ctx, message.Content)
		}

		history = append(history, message.ToParam())

		toolResults := a.callTools(ctx, message.ToolCalls)
		history = append(history, toolResults...)

		for _, hook := range a.Hooks.onAfterToolCall {
			hook(ctx, response)
		}
	}
}

func (a *Agent) callTools(ctx context.Context, toolCalls []openai.ChatCompletionMessageToolCall) []openai.ChatCompletionMessageParamUnion {
	definedTools := a.Tools
	toolResults := make([]openai.ChatCompletionMessageParamUnion, len(toolCalls))

	for i, toolCall := range toolCalls {
		args := toolCall.Function.Arguments
		name := toolCall.Function.Name

		toolIdx := slices.IndexFunc(definedTools, func(t tool.Tool) bool {
			return t.Name() == name
		})

		var toolResult string

		if toolIdx != -1 {
			result, err := definedTools[toolIdx].Call(ctx, args)
			if err != nil {
				toolResult = err.Error()
			} else {
				toolResult = result
			}
		}

		if toolResult == "" {
			toolResult = fmt.Sprintf("Unknown tool: %s", name)
		}

		for _, hook := range a.Hooks.onBeforeToolCall {
			hook(ctx, &toolCall)
		}

		toolResults[i] = openai.ToolMessage(toolResult, toolCall.ID)
	}

	return toolResults
}

func (a *Agent) toolDefs() []openai.ChatCompletionToolParam {
	openaiTools := make([]openai.ChatCompletionToolParam, len(a.Tools))

	for i, tool := range a.Tools {
		openaiTools[i] = openai.ChatCompletionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        tool.Name(),
				Description: openai.String(tool.Description()),
				Parameters:  tool.Params(),
			},
		}
	}

	return openaiTools
}

func (a *Agent) handleResponse(ctx context.Context, content string) (string, error) {
	resp := parseResponse(content)

	if resp.Error != "" {
		return "", fmt.Errorf("%s", resp.Error)
	}

	if resp.Suggestion != "" {
		for _, hook := range a.Hooks.onSuggestion {
			hook(ctx, resp.Suggestion)
		}
	}

	if resp.Result != "" {
		return resp.Result, nil
	}

	return "", fmt.Errorf("no valid response from LLM")
}

type AgentResponse struct {
	Error      string `json:"error,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	Result     string `json:"result,omitempty"`
}

func parseResponse(content string) *AgentResponse {
	result := new(AgentResponse)

	lines := strings.SplitSeq(content, "\n")

	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if err := json.Unmarshal([]byte(line), &result); err != nil {
			continue
		}
	}

	return result
}
