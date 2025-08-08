package agent

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/haadi-coder/Git-Agent/internal/llm"
	"github.com/haadi-coder/Git-Agent/internal/tool"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

type LogHook func(name string, args ...string)

type Agent struct {
	client         llm.OpenRouter
	tools          []tool.Tool
	systemPrompt   string
	responseFormat openai.ChatCompletionNewParamsResponseFormatUnion
	logHook        LogHook
}

func (a *Agent) Run(ctx context.Context) (string, error) {
	history := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(a.systemPrompt),
	}

	for {
		response, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionNewParams{
			Messages:       history,
			Tools:          a.provideTools(),
			ResponseFormat: a.responseFormat,
		})
		if err != nil {
			return "", err
		}

		message := response.Choices[0].Message
		history = append(history, message.ToParam())

		if message.Content != "" {
			a.logHook("agent", message.Content)
		}

		if len(message.ToolCalls) == 0 {
			return message.Content, nil
		}

		toolsResult := a.callTools(ctx, message.ToolCalls)
		history = append(history, toolsResult...)

		timeSpent := int(time.Now().Unix() - response.Created)
		usedTokens := int(response.Usage.CompletionTokens)
		a.logHook("info", strconv.Itoa(usedTokens), strconv.Itoa(timeSpent))
	}
}

func (a *Agent) callTools(ctx context.Context, toolCalls []openai.ChatCompletionMessageToolCall) []openai.ChatCompletionMessageParamUnion {
	toolsResult := []openai.ChatCompletionMessageParamUnion{}

	for _, toolCall := range toolCalls {
		args := toolCall.Function.Arguments
		name := toolCall.Function.Name

		var toolResult string
		for _, tool := range a.tools {
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

		a.logHook("tool", toolCall.Function.Name, args)
		toolsResult = append(toolsResult, openai.ToolMessage(toolResult, toolCall.ID))
	}

	return toolsResult
}

func (a *Agent) provideTools() []openai.ChatCompletionToolParam {
	openaiTools := make([]openai.ChatCompletionToolParam, 0, len(a.tools))

	for _, tool := range a.tools {
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
