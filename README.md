# RAG Chatbot with Groq & pgvector

A simple terminal-based RAG (Retrieval-Augmented Generation) chatbot built in Go using Groq API for LLM inference and PostgreSQL with pgvector for vector storage.

## Features

- 📚 Load and process text files (no PDF support)
- 🔍 Vector similarity search using pgvector
- 💬 Interactive terminal chat interface
- 🎨 Colorful terminal output
- 🤖 Powered by Groq's LLM API

## Prerequisites

- Go 1.21+
- Neon PostgreSQL account (free tier available)
- Groq API key (get one from [Groq Console](https://console.groq.com))

## Setup Neon PostgreSQL

**Neon is a serverless PostgreSQL with pgvector built-in - no Docker needed!**

1. **Create a Neon account**: Go to [console.neon.tech](https://console.neon.tech) and sign up (free tier available)

2. **Create a new project**: Click "Create a Project"

3. **Enable pgvector**: In your project, go to SQL Editor and run:
   ```sql
   CREATE EXTENSION IF NOT EXISTS vector;
   ```

4. **Initialize the database**: Run the provided schema:
   ```bash
   # Copy the init.sql content and run it in Neon's SQL Editor
   # Or use psql:
   psql "<your-neon-connection-string>" -f init.sql
   ```

5. **Get your connection string**: 
   - In Neon dashboard, click "Connection Details"
   - Copy the connection string (it looks like: `postgres://user:pass@ep-xxx.region.aws.neon.tech/dbname?sslmode=require`)

## Installation

1. Clone this repository:
```bash
cd go-groq
```

2. Install dependencies:
```bash
go mod download
```

3. Create `.env` file:
```bash
cp .env.example .env
```

4. Edit `.env` and add your credentials:
```
GROQ_API_KEY=your_actual_groq_api_key
DATABASE_URL=postgres://username:password@ep-xxx-xxx.region.aws.neon.tech/neondb?sslmode=require
```

   **Get your Neon connection string:**
   - Login to [console.neon.tech](https://console.neon.tech)
   - Select your project
   - Click "Connection Details" → Copy the connection string
   - Paste it as DATABASE_URL in your `.env` file

## Usage

### Prepare Your Data

Place text files in the project directory:
- `data.txt` (will be auto-created with sample data if not found)
- `knowledge.txt`
- `info.txt`

The app will automatically load and process these files on first run.

### Run the Chatbot

```bash
go run .
```

### Chat Commands

- Type your questions naturally
- `stats` - Show database statistics
- `clear` - Clear the screen
- `exit` or `quit` - Exit the chatbot

## Project Structure

```
go-groq/
├── main.go          # Entry point & initialization
├── config.go        # Configuration management
├── embedder.go      # Text chunking & embeddings
├── store.go         # Vector database operations
├── chat.go          # RAG chat logic
├── go.mod           # Go dependencies
├── .env             # Environment variables
└── README.md        # This file
```

## How It Works

1. **Document Loading**: Text files are loaded and split into overlapping chunks
2. **Embedding**: Each chunk is converted to a vector embedding
3. **Storage**: Embeddings are stored in PostgreSQL with pgvector
4. **Query**: User questions are embedded and similar chunks are retrieved
5. **Generation**: Retrieved context + question sent to Groq for answer generation

## Example Interaction

```
╔════════════════════════════════════════╗
║   RAG Chatbot with Groq & pgvector    ║
╚════════════════════════════════════════╝

You: What is Go programming language?

Bot: Go is a statically typed, compiled programming language 
designed at Google. It features fast compilation, built-in 
concurrency with goroutines, and a clean syntax...

You: exit
Goodbye! 👋
```

## Configuration

Edit values in [config.go](config.go):
- `ChunkSize`: Characters per text chunk (default: 500)
- `ChunkOverlap`: Overlap between chunks (default: 50)
- `ChatModel`: Groq model to use (default: llama-3.3-70b-versatile)
- `SystemPrompt`: Bot personality prompt

## Notes

- The current embedding implementation uses a simple hash-based approach for demonstration
- For production, consider using proper embedding models (OpenAI, Cohere, etc.)
- The system automatically creates required database tables and indexes
- pgvector uses cosine similarity for vector search

## Troubleshooting

**Database connection error:**
- Verify your Neon connection string in `.env`
- Ensure `sslmode=require` is in the connection string
- Check that pgvector extension is enabled: `CREATE EXTENSION IF NOT EXISTS vector;`
- Verify your Neon project is active at [console.neon.tech](https://console.neon.tech)

**Groq API error:**
- Verify your GROQ_API_KEY in `.env`
- Check your API quota at [Groq Console](https://console.groq.com)

## License

MIT

## Contributing

Feel free to open issues or submit pull requests!
