package main

import (
	"context"
	"log"
	"os"

	"github.com/fatih/color"
)

func main() {
	ctx := context.Background()

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize vector store
	vectorStore, err := NewVectorStore(config)
	if err != nil {
		log.Fatalf("Failed to initialize vector store: %v", err)
	}
	defer vectorStore.Close()

	// Initialize embedder
	embedder := NewEmbedder(config)

	// Check if we need to load documents
	count, err := vectorStore.Count(ctx)
	if err != nil {
		log.Fatalf("Failed to check document count: %v", err)
	}

	yellow := color.New(color.FgYellow)
	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)

	if count == 0 {
		yellow.Println("\n📚 No documents in database. Let's load some!")

		// Check for data files
		dataFiles := []string{"data.txt", "knowledge.txt", "info.txt"}
		existingFiles := []string{}

		for _, file := range dataFiles {
			if _, err := os.Stat(file); err == nil {
				existingFiles = append(existingFiles, file)
			}
		}

		if len(existingFiles) == 0 {
			yellow.Println("\n⚠️  No data files found. Creating sample data.txt...")

			// Create sample data file
			sampleData := `Go is a statically typed, compiled programming language designed at Google.
It is syntactically similar to C, but with memory safety, garbage collection, and structural typing.
Go was designed to address criticisms of other languages while maintaining their positive characteristics.

Key features of Go include:
- Fast compilation times
- Built-in concurrency support with goroutines and channels
- Simple and clean syntax
- Strong standard library
- Cross-platform support

Go is widely used for:
- Web servers and APIs
- Cloud services and infrastructure
- Command-line tools
- Distributed systems
- Microservices

The Go mascot is the Gopher, designed by Renée French.
Go was announced in November 2009 and became open source in 2012.
Major companies using Go include Google, Uber, Docker, and Kubernetes.`

			if err := os.WriteFile("data.txt", []byte(sampleData), 0644); err != nil {
				log.Fatalf("Failed to create sample data: %v", err)
			}
			existingFiles = []string{"data.txt"}
			green.Println("✓ Created data.txt with sample content")
		}

		// Load and process files
		cyan.Printf("\n📥 Loading files: %v\n", existingFiles)
		chunks, err := embedder.ProcessFiles(existingFiles)
		if err != nil {
			log.Fatalf("Failed to process files: %v", err)
		}

		green.Printf("✓ Created %d text chunks\n", len(chunks))

		// Store chunks with embeddings
		cyan.Println("\n🔢 Generating embeddings and storing in database...")
		for i, chunk := range chunks {
			embedding, err := embedder.GetEmbedding(ctx, chunk.Content)
			if err != nil {
				log.Printf("Failed to get embedding for chunk %d: %v", i, err)
				continue
			}

			if err := vectorStore.StoreChunk(ctx, chunk, embedding); err != nil {
				log.Printf("Failed to store chunk %d: %v", i, err)
				continue
			}
		}

		green.Printf("✓ Stored %d chunks in vector database\n", len(chunks))
	} else {
		green.Printf("\n✓ Database contains %d document chunks\n", count)
	}

	// Initialize chatbot
	chatBot := NewChatBot(vectorStore, embedder, config)

	// Run interactive chat
	if err := chatBot.RunInteractive(ctx); err != nil {
		log.Fatalf("Chat error: %v", err)
	}
}
