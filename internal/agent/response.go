package agent

import (
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
)

type Response struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func parseResponse(content string) (*Response, error) {
	result := new(Response)

	if content != "" {
		if err := json.Unmarshal([]byte(content), &result); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
	}

	if result.Type == "error" {
		return nil, fmt.Errorf("llm failed with - %s", result.Value)
	}

	return result, nil
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
					"type": map[string]any{
						"type":        "string",
						"enum":        []string{"error", "suggestion", "result"},
						"description": "The type of response: 'error', 'suggestion' or 'result'.",
					},
					"value": map[string]any{
						"type":        "string",
						"description": "The content of the response (error message, suggestion details or commit message).",
					},
				},
				"required":             []string{"type", "value"},
				"additionalProperties": false,
			},
		},
	},
}
