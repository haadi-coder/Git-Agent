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

const Version = "1.0.0"

type Options struct {
	APIKey        string        `short:"k" long:"api-key" description:"API key for LLM provider" env:"GA_API_KEY" required:"true"`
	Model         string        `short:"m" long:"model" description:"Model to use" env:"GA_MODEL" default:"anthropic/claude-3.5-haiku"`
	MaxTokens     int64         `short:"t" long:"max-tokens" description:"Maximum tokens per session" env:"GA_MAX_TOKENS" default:"8192"`
	Timeout       time.Duration `long:"timeout" description:"API request timeout" env:"GA_TIMEOUT" default:"30s"`
	Instructions  []string      `short:"i" long:"instruction" description:"Additional instruction for the agent (can be used multiple times)" env:"GA_INSTRUCTIONS"`
	Verbose       bool          `short:"v" long:"verbose" description:"Show detailed agent actions" env:"GA_VERBOSE"`
	NoInteractive bool          `short:"y" long:"non-interactive" description:"Commit without confirmation prompt" env:"GA_NO_INTERACTIVE"`
	Version       bool          `long:"version" description:"Show version information"`
	Help          bool          `short:"h" long:"help" description:"Show this help message"`
}

func main() {
	var opts Options

	parser := flags.NewParser(&opts, flags.IniDefault)
	parser.Usage = "AI-powered commit message generator\n\nExample:\n  ga commit [options]"

	args, err := parser.Parse()
	if err != nil {
		fmt.Printf("Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	if len(args) == 0 || args[0] != "commit" {
		fmt.Println("Usage: ga commit [options]\nUse 'ga commit --help' for more information.")
		os.Exit(1)
	}

	if opts.Version {
		fmt.Printf("Git Agent v%s\n", Version)
		os.Exit(0)
	}

	if opts.Help {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	envInstructions := os.Getenv("GA_INSTRUCTIONS")
	if envInstructions != "" {
		for instruction := range strings.FieldsSeq(envInstructions) {
			opts.Instructions = append(opts.Instructions, instruction)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Print("\nExited\n")
		cancel()
		os.Exit(1)
	}()

	openrouter := llm.NewOpenRouter(&llm.OpenRouterConfig{
		APIKey:    opts.APIKey,
		Model:     opts.Model,
		MaxTokens: opts.MaxTokens,
		Timeout:   opts.Timeout,
	})

	agent := agent.NewCommitAgent(openrouter, opts.Instructions)

	if err = Run(ctx, agent, &opts); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func Run(ctx context.Context, agent *agent.Agent, opts *Options) error {
	if opts.Verbose {
		fmt.Println(color.Cyan("=== Git Agent Session Started ==="))
		fmt.Printf(color.Black("Start Time: ")+"%s\n", time.Now().Format(time.TimeOnly))
		fmt.Printf(color.Black("Max Tokens: ")+"%d\n", opts.MaxTokens)
		fmt.Printf(color.Black("Model: ")+"%s", opts.Model)
		if len(opts.Instructions) > 0 {
			fmt.Println(color.Black("Instructions: "), strings.Join(opts.Instructions, ", "))
		}
	}

	fmt.Println("\n\nAnalizing changes...")

	commitMessage, err := agent.Run(ctx)
	if err != nil {
		fmt.Print(color.Redf("Error: %v\n", err))
		os.Exit(1)
	}

	fmt.Println(color.Cyan("\nGenerated commit message:"))
	fmt.Println(commitMessage)

	if !opts.NoInteractive {
		fmt.Println("\nCommit with this message? [Y/n]:")
		reader := bufio.NewReader(os.Stdin)
		userInput, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		userInput = strings.ToLower(strings.TrimSpace(userInput))
		if userInput == "n" || userInput == "no" {
			fmt.Println(color.Red("Message not commited"))
			os.Exit(0)
		}
	}

	if err := perfomCommit(commitMessage); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	fmt.Print(color.Green("Succesfully commited"))
	return nil
}

func perfomCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}
