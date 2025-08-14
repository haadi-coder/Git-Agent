package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/haadi-coder/Git-Agent/internal/agent"
	"github.com/haadi-coder/Git-Agent/internal/llm"
	"github.com/haadi-coder/color"
	"github.com/jessevdk/go-flags"
	"github.com/openai/openai-go"
)

const revision = "unknow"

type options struct {
	APIKey        string        `short:"k" long:"api-key" description:"API key for LLM provider" env:"GA_API_KEY" `
	Model         string        `short:"m" long:"model" description:"Model to use" env:"GA_MODEL" default:"anthropic/claude-3.5-haiku"`
	MaxTokens     int64         `short:"t" long:"max-tokens" description:"Maximum tokens per session" env:"GA_MAX_TOKENS" default:"8192"`
	Timeout       time.Duration `long:"timeout" description:"API request timeout" env:"GA_TIMEOUT" default:"30s"`
	Instructions  []string      `short:"i" long:"instruction" description:"Additional instruction for the agent (can be used multiple times)" env:"GA_INSTRUCTIONS" env-delim:"\n"`
	Verbose       bool          `short:"v" long:"verbose" description:"Show detailed agent actions" env:"GA_VERBOSE"`
	NoInteractive bool          `short:"y" long:"non-interactive" description:"Commit without confirmation prompt" env:"GA_NO_INTERACTIVE"`
	Version       bool          `long:"version" description:"Show version information"`
}

func main() {
	opts, args, err := parseOpts()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	if len(args) == 0 || args[0] != "commit" {
		fmt.Println("Usage: ga commit [options]\nUse 'ga commit --help' for more information.")
		os.Exit(1)
	}

	if opts.Version {
		fmt.Printf("Git Agent %s\n", revision)
		os.Exit(0)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, opts); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}

func parseOpts() (*options, []string, error) {
	var opts options

	parser := flags.NewParser(&opts, flags.Default)
	parser.Usage = "AI-powered commit message generator\n\nExample:\n  ga commit [options]"

	args, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			return &opts, nil, fmt.Errorf("failed to parse arguments: %w", err)
		}
	}

	return &opts, args, nil
}

func run(ctx context.Context, opts *options) error {
	openrouter := llm.NewOpenRouter(&llm.OpenRouterConfig{
		APIKey:    opts.APIKey,
		Model:     opts.Model,
		MaxTokens: opts.MaxTokens,
		Timeout:   opts.Timeout,
	})

	hooks := agent.Hooks{}

	hooks.AddOnIntermidiateStep(func(ctx context.Context, response *openai.ChatCompletion) {
		message := response.Choices[0].Message
		fmt.Print("\n")
		fmt.Print(color.Cyan("✦ "))
		fmt.Println(color.Yellow("Agent:"), message.Content)
	})

	hooks.AddOnAfterIntermidiateStep(func(ctx context.Context, response *openai.ChatCompletion) {
		if !opts.Verbose {
			return
		}

		timeSpent := int(time.Now().Unix() - response.Created)
		usedTokens := int(response.Usage.CompletionTokens)

		fmt.Printf(color.Black("  Info: "+"Used Tokens: %d, Time spent: %ds\n"), usedTokens, timeSpent)
	})

	hooks.AddBeforeCallTool(func(ctx context.Context, toolCall *openai.ChatCompletionMessageToolCall) {
		name := toolCall.Function.Name
		args := toolCall.Function.Arguments

		fmt.Printf(color.Blue("  Tool: ")+"%s(%s)\n", name, args)
	})

	hooks.AddOnSuggestion(func(ctx context.Context, suggestion string, history *[]openai.ChatCompletionMessageParamUnion) {
		fmt.Print(color.Cyan("\nSuggestion:\n"))
		fmt.Println(suggestion)
	})

	agent := agent.NewAgent(openrouter, &hooks, opts.Instructions)

	if opts.Verbose {
		fmt.Println(color.Cyan("=== Git Agent Session Started ==="))
		fmt.Printf(color.Black("⌛ Start Time: ")+"%s\n", time.Now().Format(time.TimeOnly))
		fmt.Printf(color.Black("🚩 Max Tokens: ")+"%d\n", opts.MaxTokens)
		fmt.Printf(color.Black("🤖 Model: ")+"%s", opts.Model)
		if len(opts.Instructions) > 0 {
			fmt.Println(color.Black("📝 Instructions: "), strings.Join(opts.Instructions, ", "))
		}
		fmt.Print("\n\n")
	}

	fmt.Println("🔎 Analyzing changes...")

	commitMessage, err := agent.Run(ctx)
	if err != nil {
		return fmt.Errorf(color.Red("Error: %w\n"), err)
	}

	fmt.Println(color.Cyan("\n📜 Generated commit message:"))
	fmt.Println(commitMessage)

	if !opts.NoInteractive {
		fmt.Print("\n❓ Commit with this message? [Y/n]:")

		userInput, err := getUserMessage(ctx)
		if err != nil {
			return fmt.Errorf("\nfailed to get user message: %w", err)
		}

		prepared := strings.ToLower(strings.TrimSpace(userInput))
		if prepared == "n" || prepared == "no" {
			fmt.Println(color.Red("❌ Message not commited"))
			return nil
		}
	}

	if err := perfomCommit(ctx, commitMessage); err != nil {
		return fmt.Errorf("\nfailed to commit: %w", err)
	}

	fmt.Print(color.Green("✅ Succesfully commited"))

	return nil
}

func perfomCommit(ctx context.Context, message string) error {
	cmd := exec.CommandContext(ctx, "git", "commit", "-m", message)
	return cmd.Run()
}

func getUserMessage(ctx context.Context) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	resultChan := make(chan string)

	go func() {
		if scanner.Scan() {
			resultChan <- scanner.Text()
		}
	}()

	select {
	case text := <-resultChan:
		return text, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
