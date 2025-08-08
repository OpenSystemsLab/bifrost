//go:build pinecone

package smartpromptopt

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	pinecone "github.com/pinecone-io/go-pinecone/v4/pinecone"
)

// pineconeSDKClient is an implementation of PineconeClient using go-pinecone SDK v4.
// This file is only built when the 'pinecone' build tag is provided.
// Enable with: go build -tags pinecone
// Ensure your module depends on the SDK: go get github.com/pinecone-io/go-pinecone/v4/pinecone

type pineconeSDKClient struct {
	cfg   PineconeConfig
	cli   *pinecone.Client
	index pinecone.IndexClient
}

func NewPineconeHTTPClient(cfg PineconeConfig) PineconeClient {
	// When the pinecone build tag is enabled, use the real SDK client; otherwise, the stub in pinecone.go will be used.
	apiKey := os.Getenv(cfg.ApiKeyEnv)
	if apiKey == "" {
		// Fallback to stub-like behavior if no key is present to avoid runtime panics
		return &pineconeHTTPClient{cfg: cfg}
	}

	client, err := pinecone.NewClient(pinecone.Config{APIKey: apiKey, Environment: cfg.Environment})
	if err != nil {
		return &pineconeHTTPClient{cfg: cfg}
	}

	idx := client.Index(cfg.IndexName)
	return &pineconeSDKClient{cfg: cfg, cli: client, index: idx}
}

func (c *pineconeSDKClient) UpsertTexts(ctx context.Context, texts []string, metadata map[string]string) error {
	if len(texts) == 0 { return nil }
	// Prepare records; rely on server-side embedding by setting the Text field.
	records := make([]pinecone.Record, 0, len(texts))
	for i, t := range texts {
		rec := pinecone.Record{
			ID:        fmt.Sprintf("doc-%d-%d", time.Now().UnixNano(), i),
			Text:      t,
			Metadata:  map[string]any{},
		}
		for k, v := range metadata { rec.Metadata[k] = v }
		records = append(records, rec)
	}
	_, err := c.index.Upsert(ctx, pinecone.UpsertRequest{Records: records, Namespace: &c.cfg.Namespace})
	return err
}

func (c *pineconeSDKClient) QueryByText(ctx context.Context, text string, topK int) []PineconeMatch {
	if strings.TrimSpace(text) == "" { return nil }
	if topK <= 0 { topK = c.cfg.TopK }
	req := pinecone.QueryRequest{TopK: topK, Namespace: &c.cfg.Namespace, IncludeValues: false, IncludeMetadata: true, Text: &text}
	resp, err := c.index.Query(ctx, req)
	if err != nil || resp == nil || len(resp.Matches) == 0 { return nil }
	out := make([]PineconeMatch, 0, len(resp.Matches))
	for _, m := range resp.Matches {
		pm := PineconeMatch{ID: m.ID, Score: m.Score}
		if m.Metadata != nil {
			if txt, ok := m.Metadata["text"].(string); ok { pm.Text = txt }
		}
		out = append(out, pm)
	}
	return out
}
