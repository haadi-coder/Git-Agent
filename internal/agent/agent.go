package agent

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/haadi-coder/Git-Agent/internal/llm"
	"github.com/haadi-coder/Git-Agent/internal/tool"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"

	_ "embed"
)

type Agent struct {
	LLM          *llm.OpenRouter
	SystemPrompt string
	Hooks        *Hooks
}

var tools [5]tool.Tool
var toolLookup = make(map[string]tool.Tool)
var toolsDefinition = make([]openai.ChatCompletionToolParam, len(tools))

func init() {
	tools = [5]tool.Tool{
		&tool.Read{},
		&tool.LS{},
		&tool.Git{},
		&tool.Glob{},
		&tool.Grep{},
	}

	for i, t := range tools {
		toolLookup[t.Name()] = t

		toolsDefinition[i] = openai.ChatCompletionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        t.Name(),
				Description: openai.String(t.Description()),
				Parameters:  t.Params(),
			},
		}
	}
}

func NewAgent(llmClient *llm.OpenRouter, hooks *Hooks, instructions []string) *Agent {
	return &Agent{
		LLM:          llmClient,
		SystemPrompt: buildSystemPrompt(instructions),
		Hooks:        hooks,
	}
}

//go:embed system_prompt.md
var systemPrompt string

func buildSystemPrompt(instructions []string) string {
	data := struct {
		Instructions []string
	}{
		Instructions: instructions,
	}

	template, err := template.New("improved_system_prompt").Parse(systemPrompt)
	if err != nil {
		fmt.Printf("Template reading error: %v\n", err)
	}

	var buf bytes.Buffer
	err = template.Execute(&buf, data)
	if err != nil {
		fmt.Println("Executing template error:", err)
	}

	return buf.String()
}

func (a *Agent) Run(ctx context.Context) (string, error) {
	history := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(a.SystemPrompt),
	}

	for {
		response, err := a.LLM.CreateChatCompletion(ctx, openai.ChatCompletionNewParams{
			Messages:       history,
			Tools:          toolsDefinition,
			ResponseFormat: *responseFormat,
		})
		if err != nil {
			return "", err
		}

		message := response.Choices[0].Message

		isIntermidiateStep := message.Content != "" && len(message.ToolCalls) != 0
		if isIntermidiateStep {
			a.Hooks.handleIntermidiateStep(ctx, response)
		}

		if len(message.ToolCalls) == 0 {
			return a.handleResponse(ctx, message.Content)
		}

		history = append(history, message.ToParam())

		toolResults := a.callTools(ctx, message.ToolCalls)
		history = append(history, toolResults...)

		a.Hooks.handleAfterToolCall(ctx, response)
	}
}

func (a *Agent) callTools(ctx context.Context, toolCalls []openai.ChatCompletionMessageToolCall) []openai.ChatCompletionMessageParamUnion {
	toolResults := make([]openai.ChatCompletionMessageParamUnion, len(toolCalls))

	for i, toolCall := range toolCalls {
		args := toolCall.Function.Arguments
		name := toolCall.Function.Name

		var toolResult string

		if tool, ok := toolLookup[name]; ok {
			result, err := tool.Call(ctx, args)
			if err != nil {
				toolResult = err.Error()
			} else {
				toolResult = result
			}
		}

		if toolResult == "" {
			toolResult = fmt.Sprintf("Unknown tool: %s", name)
		}

		a.Hooks.handleBeforeToolCall(ctx, &toolCall)

		toolResults[i] = openai.ToolMessage(toolResult, toolCall.ID)
	}

	return toolResults
}
