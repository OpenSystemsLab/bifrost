package smartpromptopt

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	pinecone "github.com/pinecone-io/go-pinecone/v4/pinecone"
)

type PineconeSDKClient struct {
	cfg    PineconeConfig
	client *pinecone.Client
	index  *pinecone.IndexConnection
}

type SearchResult struct {
	ID       string
	Score    float64
	Text     string
	Metadata map[string]string
}

func NewPineconeSDKClient(cfg PineconeConfig) (*PineconeSDKClient, error) {
	apiKey := os.Getenv(cfg.ApiKeyEnv)
	if apiKey == "" {
		return nil, fmt.Errorf("%s is not set", cfg.ApiKeyEnv)
	}

	params := pinecone.NewClientParams{
		ApiKey: apiKey,
	}

	client, err := pinecone.NewClient(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create pinecone client: %w", err)
	}

	idx, err := client.DescribeIndex(context.Background(), cfg.IndexName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe index: %w", err)
	}

	connParams := pinecone.NewIndexConnParams{Host: idx.Host}
	index, err := client.Index(connParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create index connection: %w", err)
	}

	return &PineconeSDKClient{cfg: cfg, client: client, index: index}, nil
}

func (c *PineconeSDKClient) UpsertTexts(ctx context.Context, texts []string, metadata map[string]string) error {
	if len(texts) == 0 {
		return nil
	}

	var records []*pinecone.IntegratedRecord
	for i, t := range texts {
		rec := pinecone.IntegratedRecord{
			"_id":  fmt.Sprintf("doc-%d-%d", time.Now().UnixNano(), i),
			"text": t,
		}
		for k, v := range metadata {
			if k != "text" && k != "_id" && k != "id" {
				rec[k] = v
			}
		}
		records = append(records, &rec)
	}

	batchSize := 25
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		err := c.index.UpsertRecords(ctx, batch)
		if err != nil {
			return fmt.Errorf("failed to upsert records batch %d-%d: %w", i, end, err)
		}
		fmt.Printf("Upserted records batch %d-%d\n", i, end)
	}
	return nil
}

func (c *PineconeSDKClient) QueryByText(ctx context.Context, text string, topK int) ([]*SearchResult, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text is empty")
	}
	// Use SearchRecords for text search with integrated embedding
	req := &pinecone.SearchRecordsRequest{
		Query: pinecone.SearchRecordsQuery{
			TopK: int32(topK),
			Inputs: &map[string]interface{}{
				"text": text,
			},
		},
		Fields: &[]string{"content", "title", "file_path", "section"},
	}

	resp, err := c.index.SearchRecords(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search records: %w", err)
	}

	var results []*SearchResult
	for _, hit := range resp.Result.Hits {
		result := &SearchResult{
			ID:    hit.Id,
			Score: float64(hit.Score),
		}
		// Extract metadata if available
		if hit.Fields != nil {
			metadata := hit.Fields
			if text, ok := metadata["text"].(string); ok {
				result.Text = text
			} else {
				result.Text = "Text not available"
			}

			// Convert metadata to string map
			result.Metadata = make(map[string]string)
			for key, value := range metadata {
				if strValue, ok := value.(string); ok {
					result.Metadata[key] = strValue
				}
			}
		} else {
			// Fallback values if no metadata
			result.Text = "Text not available"
		}
		results = append(results, result)
	}
	return results, nil
}
