package agent

import (
	"context"
	"fmt"
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

// TODO: 7,9

// Questions: 8, 10
func (a *Agent) Run(ctx context.Context) (string, error) {
	toolDefs := a.provideTools()
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

		toolsResult := a.callTools(ctx, message.ToolCalls)
		history = append(history, toolsResult...)

		timeSpent := int(time.Now().Unix() - response.Created)
		usedTokens := int(response.Usage.CompletionTokens)
		a.Hooks.Info(usedTokens, timeSpent)
	}
}

func (a *Agent) callTools(ctx context.Context, toolCalls []openai.ChatCompletionMessageToolCall) []openai.ChatCompletionMessageParamUnion {
	toolResults := []openai.ChatCompletionMessageParamUnion{}

	for _, toolCall := range toolCalls {
		args := toolCall.Function.Arguments
		name := toolCall.Function.Name

		var toolResult string
		for _, tool := range a.Tools {
			if tool.Name() == name {
				result, err := tool.Call(ctx, args)
				if err != nil {
					toolResult = err.Error()
					break
				}

				toolResult = result
				break
			}
		}

		if toolResult == "" {
			toolResult = fmt.Sprintf("Unknown tool: %s", name)
		}

		a.Hooks.Tool(toolCall.Function.Name, args)
		toolResults = append(toolResults, openai.ToolMessage(toolResult, toolCall.ID))
	}

	return toolResults
}

func (a *Agent) provideTools() []openai.ChatCompletionToolParam {
	openaiTools := make([]openai.ChatCompletionToolParam, 0, len(a.Tools))

	for _, tool := range a.Tools {
		openaiTools = append(openaiTools, openai.ChatCompletionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        tool.Name(),
				Description: openai.String(tool.Description()),
				Parameters:  tool.Params(),
			},
		})
	}

	return openaiTools
}
