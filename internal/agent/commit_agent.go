package agent

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/haadi-coder/Git-Agent/internal/llm"
	"github.com/haadi-coder/Git-Agent/internal/tool"
	"github.com/openai/openai-go"
)

type CommitAgent struct {
	*Agent
}

func NewCommitAgent(llmClient *llm.OpenRouter, instructions []string) *CommitAgent {
	tools := []tool.Tool{
		&tool.ReadFile{},
		&tool.ListFilesTool{},
		&tool.Git{},
		&tool.Glob{},
	}

	baseAgent := &Agent{
		client:         *llmClient,
		tools:          tools,
		systemPrompt:   buildSystemPrompt(instructions),
		responseFormat: *responseFormat,
	}

	return &CommitAgent{
		Agent: baseAgent,
	}
}

var responseFormat = &openai.ChatCompletionNewParamsResponseFormatUnion{
	OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
		JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:        "final_output strictly in this format without any spaces and newlines",
			Description: openai.String("final output format"),
			Schema: &openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"commit_message": map[string]any{
						"type":        "string",
						"description": "finaly generated commit message",
					},
				},
			},
		},
	},
}

func buildSystemPrompt(instructions []string) string {
	data := struct {
		Instructions []string
	}{
		Instructions: instructions,
	}

	template, err := template.ParseFiles("system_prompt.md")
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
