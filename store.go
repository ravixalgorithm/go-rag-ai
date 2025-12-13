package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

// VectorStore handles vector database operations
type VectorStore struct {
	db     *sql.DB
	config *Config
}

// NewVectorStore creates a new vector store instance
func NewVectorStore(config *Config) (*VectorStore, error) {
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &VectorStore{
		db:     db,
		config: config,
	}

	// Initialize database schema
	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// initSchema creates the necessary tables and extensions
func (vs *VectorStore) initSchema() error {
	queries := []string{
		// Enable pgvector extension
		`CREATE EXTENSION IF NOT EXISTS vector;`,

		// Create documents table
		`CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			source TEXT NOT NULL,
			metadata JSONB,
			embedding vector(384),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		// Create index for vector similarity search
		`CREATE INDEX IF NOT EXISTS documents_embedding_idx ON documents 
		 USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);`,
	}

	for _, query := range queries {
		if _, err := vs.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// StoreChunk stores a text chunk with its embedding
func (vs *VectorStore) StoreChunk(ctx context.Context, chunk TextChunk, embedding []float32) error {
	id := uuid.New().String()

	// Convert float32 slice to pgvector.Vector
	vec := pgvector.NewVector(embedding)

	query := `
		INSERT INTO documents (id, content, source, metadata, embedding)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			embedding = EXCLUDED.embedding
	`

	_, err := vs.db.ExecContext(ctx, query, id, chunk.Content, chunk.Source, nil, vec)
	if err != nil {
		return fmt.Errorf("failed to store chunk: %w", err)
	}

	return nil
}

// SearchResult represents a search result with similarity score
type SearchResult struct {
	ID         string
	Content    string
	Source     string
	Similarity float64
}

// Search performs vector similarity search
func (vs *VectorStore) Search(ctx context.Context, embedding []float32, topK int) ([]SearchResult, error) {
	vec := pgvector.NewVector(embedding)

	query := `
		SELECT id, content, source, 1 - (embedding <=> $1) as similarity
		FROM documents
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $2
	`

	rows, err := vs.db.QueryContext(ctx, query, vec, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer rows.Close()

	results := []SearchResult{}
	for rows.Next() {
		var result SearchResult
		if err := rows.Scan(&result.ID, &result.Content, &result.Source, &result.Similarity); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// Clear removes all documents from the store
func (vs *VectorStore) Clear(ctx context.Context) error {
	_, err := vs.db.ExecContext(ctx, "DELETE FROM documents")
	return err
}

// Count returns the number of documents in the store
func (vs *VectorStore) Count(ctx context.Context) (int, error) {
	var count int
	err := vs.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM documents").Scan(&count)
	return count, err
}

// Close closes the database connection
func (vs *VectorStore) Close() error {
	return vs.db.Close()
}
