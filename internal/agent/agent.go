package agent

import (
	"context"
	"fmt"

	"github.com/haadi-coder/Git-Agent/internal/llm"
	"github.com/haadi-coder/Git-Agent/internal/tool"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

var tools = [5]tool.Tool{
	&tool.Read{},
	&tool.LS{},
	&tool.Git{},
	&tool.Glob{},
	&tool.Grep{},
}
var (
	toolLookup  = make(map[string]tool.Tool, len(tools))
	openaiTools = make([]openai.ChatCompletionToolParam, len(tools))
)

func init() {
	for i, t := range tools {
		toolLookup[t.Name()] = t

		openaiTools[i] = openai.ChatCompletionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        t.Name(),
				Description: openai.String(t.Description()),
				Parameters:  t.Params(),
			},
		}
	}
}

type Agent struct {
	llm          *llm.OpenRouter
	systemPrompt string
	hooks        *Hooks
}

func NewAgent(llm *llm.OpenRouter, instructions []string, hooks *Hooks) (*Agent, error) {
	systemPrompt, err := buildSystemPrompt(instructions)
	if err != nil {
		return nil, fmt.Errorf("failed to build system prompt: %w", err)
	}

	return &Agent{
		llm:          llm,
		systemPrompt: systemPrompt,
		hooks:        hooks,
	}, nil
}

func (a *Agent) Run(ctx context.Context) (*Response, error) {
	history := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(a.systemPrompt),
	}

	for {
		resp, err := a.llm.GenerateContent(ctx, openai.ChatCompletionNewParams{
			Messages:       history,
			Tools:          openaiTools,
			ResponseFormat: *responseFormat,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to generate content: %w", err)
		}

		message := resp.Choices[0].Message

		isFinalStep := len(message.ToolCalls) == 0
		if isFinalStep {
			parsed, err := parseResponse(message.Content)
			if err != nil {
				return nil, fmt.Errorf("failed to parse response: %w", err)
			}

			return parsed, nil
		}

		a.hooks.handleIntermidiateStep(ctx, resp)

		history = append(history, message.ToParam())

		toolResults := a.callTools(ctx, message.ToolCalls)
		history = append(history, toolResults...)

		a.hooks.handleAfterIntermidiateStep(ctx, resp)
	}
}

func (a *Agent) callTools(ctx context.Context, toolCalls []openai.ChatCompletionMessageToolCall) []openai.ChatCompletionMessageParamUnion {
	toolResults := make([]openai.ChatCompletionMessageParamUnion, len(toolCalls))

	for i, toolCall := range toolCalls {
		a.hooks.handleBeforeCallTool(ctx, &toolCall)

		var toolResult string
		name := toolCall.Function.Name
		args := toolCall.Function.Arguments

		tool, ok := toolLookup[name]

		if !ok {
			toolResult = fmt.Sprintf("Unknown tool: %s", name)
		} else {
			result, err := tool.Call(ctx, args)
			if err != nil {
				toolResult = fmt.Sprintf("Error: %s", err.Error())
			} else {
				toolResult = result
			}
		}

		toolResults[i] = openai.ToolMessage(toolResult, toolCall.ID)
	}

	return toolResults
}
