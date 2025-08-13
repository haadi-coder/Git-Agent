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
		fmt.Print(err)
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

	agent := agent.NewAgent(openrouter, opts.Instructions)

	if opts.Verbose {
		fmt.Println(color.Cyan("=== Git Agent Session Started ==="))
		fmt.Printf(color.Black("⌛ Start Time: ")+"%s\n", time.Now().Format(time.TimeOnly))
		fmt.Printf(color.Black("🚩 Max Tokens: ")+"%d\n", opts.MaxTokens)
		fmt.Printf(color.Black("🤖 Model: ")+"%s", opts.Model)
		if len(opts.Instructions) > 0 {
			fmt.Println(color.Black("📝 Instructions: "), strings.Join(opts.Instructions, ", "))
		}
	}

	fmt.Println("\n\n🔎 Analizing changes...")

	commitMessage, err := agent.Run(ctx)
	if err != nil {
		return fmt.Errorf(color.Red("Error: %w\n"), err)
	}

	fmt.Println(color.Cyan("\n📜 Generated commit message:"))
	fmt.Println(commitMessage)

	if !opts.NoInteractive {
		fmt.Println("\n❓ Commit with this message? [Y/n]:")
		reader := bufio.NewReader(os.Stdin)
		userInput, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		userInput = strings.ToLower(strings.TrimSpace(userInput))
		if userInput == "n" || userInput == "no" {
			fmt.Println(color.Red("❌ Message not commited"))

			return nil
		}
	}

	if err := perfomCommit(commitMessage); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	fmt.Print(color.Green("✅ Succesfully commited"))

	return nil
}

func perfomCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}
