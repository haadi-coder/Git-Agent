package agent

import (
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
)

const (
	ResponseTypeResult     = "result"
	ResponseTypeSuggestion = "suggestion"
	ResponseTypeError      = "error"
)

type Response struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func parseResponse(content string) (*Response, error) {
	resp := new(Response)

	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return resp, nil
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
						"enum":        []string{ResponseTypeError, ResponseTypeSuggestion, ResponseTypeResult},
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
