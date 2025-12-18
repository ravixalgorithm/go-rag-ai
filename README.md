# Go RAG AI – Multi-LLM Terminal Chatbot

A terminal-based RAG (Retrieval-Augmented Generation) chatbot built in Go with **multi-provider LLM support**. Switch between providers on the fly without restarting!

## ✨ Features

- 🔄 **Multi-LLM Support** – Groq, OpenAI, Anthropic, Gemini, OpenRouter
- 🔀 **Runtime Model Switching** – Use `/model <provider>` to switch mid-chat
- 💬 **Conversation Memory** – Maintains context across messages
- 🎨 **Colorful Terminal UI** – Syntax highlighting for code blocks
- ⌨️ **Slash Commands** – `/clear`, `/history`, `/exit`, `/model`
- 🛡️ **Graceful Exit** – Clean shutdown with Ctrl+C

## 🚀 Supported Providers

| Provider | Default Model | API Key Env Var |
|----------|---------------|-----------------|
| Groq | llama-3.3-70b-versatile | `GROQ_API_KEY` |
| OpenAI | gpt-4o-mini | `OPENAI_API_KEY` |
| Anthropic | claude-3-5-sonnet-20241022 | `ANTHROPIC_API_KEY` |
| Gemini | gemini-1.5-flash | `GEMINI_API_KEY` |
| OpenRouter | meta-llama/llama-3.1-8b-instruct:free | `OPENROUTER_API_KEY` |

## 📋 Prerequisites

- Go 1.21+
- API key from at least one provider

## 🛠️ Installation

```bash
# Clone the repository
git clone https://github.com/ravixalgorithm/go-rag-ai.git
cd go-rag-ai

# Install dependencies
go mod download

# Create .env file
cp .env.example .env

# Edit .env and add your API keys
```

## ▶️ Usage

```bash
go run .
```

Or build and run the executable:

```bash
go build -o go-rag-ai.exe .
./go-rag-ai.exe
```

### Commands

| Command | Description |
|---------|-------------|
| `/model <provider> [model]` | Switch LLM provider (e.g., `/model openai gpt-4o`) |
| `/history` | View conversation history |
| `/clear` | Clear the screen |
| `/exit` | Exit the chatbot |
| `Ctrl+C` | Graceful exit |

### Example

```
You (12:00:00): /model anthropic
✅ Switched to anthropic / claude-3-5-sonnet-20241022

You (12:00:05): Hello!
anthropic (12:00:07): Hello! How can I assist you today?
```

## 📁 Project Structure

```
go-rag-ai/
├── main.go              # Entry point
├── config.go            # Configuration & env loading
├── chat.go              # Chat loop & commands
├── internal/llm/        # LLM provider clients
│   ├── client.go        # LLMClient interface
│   ├── factory.go       # Provider factory
│   ├── groq_client.go
│   ├── openai_client.go
│   ├── anthropic_client.go
│   ├── gemini_client.go
│   └── openrouter_client.go
├── .env.example         # Environment template
└── README.md
```

## 🔧 Configuration

Set environment variables in `.env`:

```bash
# Choose provider: groq, openai, anthropic, gemini, openrouter
LLM_PROVIDER=groq

# Add API keys for providers you want to use
GROQ_API_KEY=your_key
OPENAI_API_KEY=your_key
ANTHROPIC_API_KEY=your_key
GEMINI_API_KEY=your_key
OPENROUTER_API_KEY=your_key

# Optional: override default model
LLM_MODEL=llama-3.3-70b-versatile
```

## 📄 License

[MIT](./LICENSE)

## 🤝 Contributing

Contributions welcome! Feel free to open issues or submit pull requests.
