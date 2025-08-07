package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/haadi-coder/Git-Agent/internal/agent"
	"github.com/haadi-coder/Git-Agent/internal/llm"
	"github.com/haadi-coder/color"
)

const (
	APIUrl        = "https://openrouter.ai/api/v1"
	NoInteractive = false
)

func main() {
	openrouter := llm.NewOpenRouter(&llm.OpenRouterConfig{
		APIKey: "ha-ha",
		APIURL: APIUrl,
	})

	agent := agent.NewCommitAgent(openrouter, []string{})

	Run(context.Background(), agent)
}

func Run(ctx context.Context, agent *agent.CommitAgent) {
	fmt.Println("Analizing changes...")

	commitMessage := agent.RunCommit(ctx)

	fmt.Println(color.Blue("Generated commit message:"))
	fmt.Println(commitMessage)

	if !NoInteractive {
		fmt.Println("Commit with this message? [Y/n]:")
		reader := bufio.NewReader(os.Stdin)
		userInput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("failed to read input: %v", err)
			os.Exit(1)
		}

		userInput = strings.ToLower(strings.TrimSpace(userInput))
		if userInput == "n" || userInput == "no" {
			fmt.Println(color.Red("Message not commited"))
			return
		}
	}

	if err := perfomCommit(commitMessage); err != nil {
		fmt.Printf("Error committing: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(color.Green("Succesfully commited"))
}

func perfomCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}
