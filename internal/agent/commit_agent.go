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

	baseAgent := &Agent{
		client:         *llmClient,
		tools:          tools,
		systemPrompt:   buildSystemPrompt(instructions),
		responseFormat: *responseFormat,
		logHook:        logHook,
	}

	return &CommitAgent{
		Agent: baseAgent,
	}
}

var responseFormat = &openai.ChatCompletionNewParamsResponseFormatUnion{
	OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
		JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:        "commit_message_output",
			Description: openai.String("A strict and compact JSON object containing a single commit_message string, formatted with no spaces or newlines (e.g., {\"commit_message\":\"example\"})"),
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

func logHook(messageType string, args ...string) {
	switch messageType {
	case "agent":
		fmt.Println(color.Yellow("Agent:"), args[0])
	case "tool":
		fmt.Printf(color.Blue("Tool: ")+"%s(%s)\n", args[0], args[1])
	case "info":
		fmt.Printf(color.Black("Info: "+"Used Tokens: %s, Time spent: %ss\n\n"), args[0], args[1])
	}
}
