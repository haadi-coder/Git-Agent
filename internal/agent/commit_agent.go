package agent

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/haadi-coder/Git-Agent/internal/llm"
	"github.com/haadi-coder/Git-Agent/internal/tool"
	"github.com/haadi-coder/color"
	"github.com/openai/openai-go"

	_ "embed"
)

type CommitAgent struct {
	*Agent
}

func NewCommitAgent(llmClient *llm.OpenRouter, instructions []string) *CommitAgent {
	tools := []tool.Tool{
		&tool.Read{},
		&tool.LS{},
		&tool.Git{},
		&tool.Glob{},
		&tool.Grep{},
	}

	hooks := Hooks{}

	hooks.AddOnAgentContent(func(ctx context.Context, response *openai.ChatCompletion) {
		message := response.Choices[0].Message
		fmt.Println(color.Yellow("Agent:"), message.Content)
	})

	hooks.AddBeforeToolCall(func(ctx context.Context, toolCall *openai.ChatCompletionMessageToolCall) {
		name := toolCall.Function.Name
		args := toolCall.Function.Arguments

		fmt.Printf(color.Blue("Tool: ")+"%s(%s)\n", name, args)
	})

	hooks.AddAfterToolCall(func(ctx context.Context, response *openai.ChatCompletion) {
		timeSpent := int(time.Now().Unix() - response.Created)
		usedTokens := int(response.Usage.CompletionTokens)

		fmt.Printf(color.Black("Info: "+"Used Tokens: %d, Time spent: %ds\n\n"), usedTokens, timeSpent)
	})

	hooks.AddOnSuggestion(func(ctx context.Context, suggestion string) {
		fmt.Print(color.Cyan("\nSuggestion:\n"))
		fmt.Println(suggestion)
	})

	baseAgent := &Agent{
		LLM:            llmClient,
		Tools:          tools,
		SystemPrompt:   buildSystemPrompt(instructions),
		ResponseFormat: *responseFormat,
		Hooks:          &hooks,
	}

	return &CommitAgent{
		Agent: baseAgent,
	}
}

var responseFormat = &openai.ChatCompletionNewParamsResponseFormatUnion{
	OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
		JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:        "commit_response",
			Description: openai.String("Response format for commit generation with error handling and suggestions"),
			Strict:      openai.Bool(true),
			Schema: &openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"error": map[string]any{
						"type":        "string",
						"description": "Error message if something went wrong (e.g., no git repo, no staged changes).",
					},
					"suggestion": map[string]any{
						"type":        "object",
						"description": "Optional suggestion from the LLM (e.g., to split large commits).",
					},
					"result": map[string]any{
						"type":        "string",
						"description": "finaly result output. It should result message, that is ready for commiting",
					},
				},
				"additionalProperties": false,
				"anyOf": []any{
					map[string]any{"required": []string{"error"}},
					map[string]any{"required": []string{"suggestion"}},
					map[string]any{"required": []string{"result"}},
				},
			},
		},
	},
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

// ?: Возможно уже не нужно выносить в отдельную функцию
func (ca *CommitAgent) RunCommit(ctx context.Context) string {
	response, err := ca.Run(ctx)
	if err != nil {
		fmt.Print(color.Redf("Error: %v\n", err))
		os.Exit(1)
	}

	return response
}
