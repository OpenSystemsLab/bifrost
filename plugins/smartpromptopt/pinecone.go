package smartpromptopt

import (
	"context"
)

type PineconeMatch struct {
	ID    string
	Text  string
	Score float64
}

type PineconeClient interface {
	UpsertTexts(ctx context.Context, texts []string, metadata map[string]string) error
	QueryByText(ctx context.Context, text string, topK int) []PineconeMatch
}

type pineconeHTTPClient struct {
	cfg PineconeConfig
}

func NewPineconeHTTPClient(cfg PineconeConfig) PineconeClient {
	return &pineconeHTTPClient{cfg: cfg}
}

func (c *pineconeHTTPClient) UpsertTexts(ctx context.Context, texts []string, metadata map[string]string) error {
	// Stub: Intentionally no-op for initial scaffold. Replace with real HTTP calls.
	return nil
}

func (c *pineconeHTTPClient) QueryByText(ctx context.Context, text string, topK int) []PineconeMatch {
	// Stub: return empty results for initial scaffold.
	return nil
}
