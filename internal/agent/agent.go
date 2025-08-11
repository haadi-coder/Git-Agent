package agent

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/haadi-coder/Git-Agent/internal/llm"
	"github.com/haadi-coder/Git-Agent/internal/tool"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

type LogHook func(name string, args ...string)

type Agent struct {
	LLM            *llm.OpenRouter
	Tools          []tool.Tool
	SystemPrompt   string
	ResponseFormat openai.ChatCompletionNewParamsResponseFormatUnion
	Hooks          *Hooks
}

type Hooks struct {
	Info  func(usedTokens int, timeSpent int)
	Agent func(content string)
	Tool  func(name string, args string)
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
			a.Hooks.Agent(message.Content)
		}

		if len(message.ToolCalls) == 0 {
			return message.Content, nil
		}

		history = append(history, message.ToParam())

		toolResults := a.callTools(ctx, message.ToolCalls)
		history = append(history, toolResults...)

		timeSpent := int(time.Now().Unix() - response.Created)
		usedTokens := int(response.Usage.CompletionTokens)
		a.Hooks.Info(usedTokens, timeSpent)
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

		a.Hooks.Tool(toolCall.Function.Name, args)
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
