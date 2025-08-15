# Git Agent

> AI-powered commit message generator that analyzes your changes and creates meaningful commits

## ✨ Features

- **Smart Analysis**: Examines staged changes and repository context
- **Conventional Commits**: Follows your project's commit conventions automatically  
- **Interactive Mode**: Review and approve messages before committing
- **Multiple Providers**: Supports OpenRouter with various AI models

## 🚀 Installation

```bash
go install github.com/haadi-coder/Git-Agent/cmd/ga@latest
```

## 🎯 Quick Start

1. Stage your changes:
   ```bash
   git add .
   ```

2. Generate and commit:
   ```bash
   export GA_API_KEY="your-openrouter-api-key"
   ga commit
   ```

## ⚙️ Configuration

| Flag | Environment | Default | Description |
|------|-------------|---------|-------------|
| `-k, --api-key` | `GA_API_KEY` | - | OpenRouter API key |
| `-m, --model` | `GA_MODEL` | `openai/gpt-4o` | AI model to use |
| `-t, --max-tokens` | `GA_MAX_TOKENS` | `8192` | Maximum tokens per session |
| `-i, --instruction` | `GA_INSTRUCTIONS` | - | Custom instructions (repeatable) |
| `-v, --verbose` | `GA_VERBOSE` | `false` | Show detailed output |
| `-y, --non-interactive` | `GA_NO_INTERACTIVE` | `false` | Skip confirmation |

## 💡 Examples

```bash
# Basic usage
ga commit

# With custom instructions  
ga commit -i "Use imperative mood" -i "Keep under 50 characters"

# Non-interactive mode
ga commit -y

# Different model
ga commit -m "openai/gpt-4"
```

## 🤖 How It Works

1. **Analyzes** your staged changes using `git status` and `git diff --staged`
2. **Understands** your project structure and commit history
3. **Generates** a commit message following your project's conventions
4. **Confirms** with you before committing (unless `-y` is used)

## 🔧 Requirements

- Go 1.24.5+
- Git repository
- OpenRouter API key

<div align="center">
  <sub>Built with ❤️ for developers who care about clean Git history</sub>
</div>
