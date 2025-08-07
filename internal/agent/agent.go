package agent

import (
	"context"
	"fmt"

	"github.com/haadi-coder/Git-Agent/internal/llm"
	"github.com/haadi-coder/Git-Agent/internal/tool"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

type Agent struct {
	client         llm.OpenRouter
	tools          []tool.Tool
	systemPrompt   string
	responseFormat openai.ChatCompletionNewParamsResponseFormatUnion
}

func (agent *Agent) Run(ctx context.Context) (string, error) {
	history := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(agent.systemPrompt),
	}

	for {
		response, err := agent.client.CreateChatCompletion(ctx, openai.ChatCompletionNewParams{
			Messages:       history,
			Tools:          agent.provideTools(),
			ResponseFormat: agent.responseFormat,
		})
		if err != nil {
			return "", err
		}

		message := response.Choices[0].Message
		history = append(history, message.ToParam())

		if message.Content != "" {
			fmt.Printf("Agent: %s\n", message.Content)
		}

		if len(message.ToolCalls) == 0 {
			return message.Content, nil
		}

		toolsResult := agent.callTools(ctx, message.ToolCalls)
		history = append(history, toolsResult...)
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

		fmt.Printf("Tool: %s(%s)\n", toolCall.Function.Name, args)
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
