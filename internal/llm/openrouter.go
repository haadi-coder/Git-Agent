package llm

import (
	"context"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	DefaultMaxTokens = 8192
	DefaultModel     = "anthropic/claude-3.5-haiku"
)

type OpenRouter struct {
	config *OpenRouterConfig
	client *openai.Client
}

// TODO: Потом убрать отсюда APIURL
type OpenRouterConfig struct {
	APIKey    string
	APIURL    string
	Model     string
	MaxTokens int64
	Timeout   time.Duration
}

func NewOpenRouter(config *OpenRouterConfig) *OpenRouter {
	if config.MaxTokens == 0 {
		config.MaxTokens = DefaultMaxTokens
	}

	if config.Model == "" {
		config.Model = DefaultModel
	}

	client := openai.NewClient(
		option.WithAPIKey(config.APIKey),
		option.WithBaseURL(config.APIURL),
		option.WithRequestTimeout(config.Timeout),
	)

	return &OpenRouter{
		client: &client,
		config: config,
	}
}

func (c *OpenRouter) CreateChatCompletion(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	params.Model = c.config.Model
	params.MaxTokens.Value = c.config.MaxTokens

	return c.client.Chat.Completions.New(ctx, params)
}
