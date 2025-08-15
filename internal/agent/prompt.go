package agent

import (
	"bytes"
	"fmt"
	"text/template"

	_ "embed"
)

//go:embed system_prompt.md
var systemPrompt string

func buildSystemPrompt(instructions []string) (string, error) {
	data := struct {
		Instructions []string
	}{
		Instructions: instructions,
	}

	tmpl, err := template.New("improved_system_prompt").Parse(systemPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute execute: %w", err)
	}

	return buf.String(), nil
}
