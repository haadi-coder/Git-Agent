package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/haadi-coder/Git-Agent/internal/llm"
	"github.com/haadi-coder/Git-Agent/internal/tool"
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
			Name:        "commit_message_output",
			Description: openai.String("A compact JSON object containing a single commit_message string, formatted with no spaces or newlines (e.g., {\"commit_message\":\"example\"})"),
			Strict:      openai.Bool(true),
			Schema: &openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"commit_message": map[string]any{
						"type":        "string",
						"description": "The commit message as a string. Must be non-empty and contain no newlines or leading/trailing spaces.",
					},
				},
				"required":             []string{"commit_message"},
				"additionalProperties": false,
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

	template, err := template.New("system_prompt").Parse(systemPrompt)
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

func (ca *CommitAgent) RunCommit(ctx context.Context) string {
	response, err := ca.Run(ctx)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	return extractCommitMessage(response)
}

func extractCommitMessage(content string) string {
	var result struct {
		CommitMessage string `json:"commit_message"`
	}

	lines := strings.SplitSeq(content, "\n")
	for line := range lines {
		if err := json.Unmarshal([]byte(line), &result); err == nil && result.CommitMessage != "" {
			return result.CommitMessage
		}
	}

	return ""
}
