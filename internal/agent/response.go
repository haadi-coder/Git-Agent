package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
)

func (a *Agent) handleResponse(ctx context.Context, content string) (string, error) {
	resp := parseResponse(content)

	if resp.Error != "" {
		return "", fmt.Errorf("%s", resp.Error)
	}

	if resp.Suggestion != "" {
		a.Hooks.handleSuggestion(ctx, resp.Suggestion)
	}

	if resp.Result != "" {
		return resp.Result, nil
	}

	return "", fmt.Errorf("no valid response from LLM")
}

type Response struct {
	Error      string `json:"error,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	Result     string `json:"result,omitempty"`
}

func parseResponse(content string) *Response {
	result := new(Response)

	lines := strings.SplitSeq(content, "\n")

	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if err := json.Unmarshal([]byte(line), &result); err != nil {
			continue
		}
	}

	return result
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
