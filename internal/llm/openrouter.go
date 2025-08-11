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

func NewOpenRouter(config *OpenRouterConfig) *OpenRouter {
	client := openai.NewClient(
		option.WithAPIKey(config.APIKey),
		option.WithBaseURL(BaseURL),
		option.WithRequestTimeout(config.Timeout),
	)

	return &OpenRouter{
		client: &client,
		cfg:    config,
	}
}

func (c *OpenRouter) CreateChatCompletion(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	params.Model = c.cfg.Model
	params.MaxTokens.Value = c.cfg.MaxTokens

	return c.client.Chat.Completions.New(ctx, params)
}
