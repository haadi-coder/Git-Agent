package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
)

func (a *Agent) handleResponse(ctx context.Context, content string, history []openai.ChatCompletionMessageParamUnion) (string, error) {
	resp, err := parseResponse(content)
	if err != nil {
		return "", err
	}

	if resp.Type == "error" {
		return "", fmt.Errorf("%s", resp.Value)
	}

	if resp.Type == "suggestion" {
		a.Hooks.handleSuggestion(ctx, resp.Value, &history)
	}

	if resp.Type == "result" {
		return resp.Value, nil
	}

	return "", fmt.Errorf("no valid response from LLM")
}

type Response struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func parseResponse(content string) (*Response, error) {
	result := struct {
		Response Response
	}{}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed parse response")
	}

	return &result.Response, nil
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
					"response": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"type": map[string]any{
								"type":        "string",
								"enum":        []string{"error", "suggestion", "result"},
								"description": "The type of response: 'error', 'suggestion' or 'result'.",
							},
							"value": map[string]any{
								"type":        "string",
								"description": "The content of the response (error message, suggestion details, comment from llm describing its descicions, or commit message).",
							},
						},
						"required":             []string{"type", "value"},
						"additionalProperties": false,
					},
				},
				"required":             []string{"response"},
				"additionalProperties": false,
			},
		},
	},
}
