package llm

import (
	"context"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const BaseURL = "https://openrouter.ai/api/v1"

type OpenRouter struct {
	cfg    *OpenRouterConfig
	client *openai.Client
}

type OpenRouterConfig struct {
	APIKey    string
	Model     string
	MaxTokens int64
	Timeout   time.Duration
}

func NewOpenRouter(cfg *OpenRouterConfig) *OpenRouter {
	client := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(BaseURL),
		option.WithRequestTimeout(cfg.Timeout),
	)

	return &OpenRouter{
		client: &client,
		cfg:    cfg,
	}
}

func (c *OpenRouter) GenerateContent(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	params.Model = c.cfg.Model
	params.MaxTokens.Value = c.cfg.MaxTokens

	return c.client.Chat.Completions.New(ctx, params)
}
