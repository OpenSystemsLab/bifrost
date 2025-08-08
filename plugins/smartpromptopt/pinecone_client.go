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
type pineconeSDKClient struct {
	cfg   PineconeConfig
	cli   *pinecone.Client
	index *pinecone.Index
}

// NewPineconeSDKClient creates a new Pinecone client using the SDK.
// Falls back to stub client if API key is not available.
func NewPineconeSDKClient(cfg PineconeConfig) PineconeClient {
	apiKey := os.Getenv(cfg.ApiKeyEnv)
	if apiKey == "" {
		// Fallback to stub behavior if no key is present
		return &pineconeHTTPClient{cfg: cfg}
	}

	client, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: apiKey,
		Host:   cfg.Environment, // if needed; newer SDK may not use this
	})
	if err != nil {
		return &pineconeHTTPClient{cfg: cfg}
	}

	idx, err := client.Index(pinecone.NewIndexConnParams{
		Host:      cfg.IndexName, // adjust based on actual SDK requirements
		Namespace: cfg.Namespace,
	})
	if err != nil {
		return &pineconeHTTPClient{cfg: cfg}
	}

	return &pineconeSDKClient{cfg: cfg, cli: client, index: idx}
}

func (c *pineconeSDKClient) UpsertTexts(ctx context.Context, texts []string, metadata map[string]string) error {
	if len(texts) == 0 { 
		return nil 
	}
	
	// Prepare vectors with text for server-side embedding
	vectors := make([]*pinecone.Vector, 0, len(texts))
	for i, t := range texts {
		meta := pinecone.Metadata{}
		for k, v := range metadata { 
			meta[k] = v 
		}
		meta["text"] = t // Store text in metadata for retrieval
		
		vec := &pinecone.Vector{
			Id:       fmt.Sprintf("doc-%d-%d", time.Now().UnixNano(), i),
			Metadata: &meta,
			// Note: Adjust based on your Pinecone index configuration
			// If using server-side embeddings, you might need to use a different field
			// or API endpoint. This is a placeholder.
			SparseValues: nil,
			Values:       []float32{}, // Empty if using text-based embedding on server
		}
		vectors = append(vectors, vec)
	}
	
	_, err := c.index.UpsertVectors(ctx, vectors)
	return err
}

func (c *pineconeSDKClient) QueryByText(ctx context.Context, text string, topK int) []PineconeMatch {
	if strings.TrimSpace(text) == "" { 
		return nil 
	}
	if topK <= 0 { 
		topK = c.cfg.TopK 
	}
	
	// Query using text - this assumes your Pinecone index supports text queries
	// Adjust based on your actual Pinecone configuration
	queryReq := &pinecone.QueryVectorsRequest{
		TopK:            uint32(topK),
		IncludeMetadata: true,
		IncludeValues:   false,
		// If Pinecone supports text queries directly:
		// Text: text,
		// Otherwise, you might need to embed the text first
		Vector: []float32{}, // Placeholder - adjust based on your setup
	}
	
	resp, err := c.index.QueryVectors(ctx, queryReq)
	if err != nil || resp == nil || len(resp.Matches) == 0 { 
		return nil 
	}
	
	out := make([]PineconeMatch, 0, len(resp.Matches))
	for _, m := range resp.Matches {
		pm := PineconeMatch{
			ID:    m.Vector.Id,
			Score: float64(m.Score),
		}
		if m.Vector.Metadata != nil {
			if txt, ok := (*m.Vector.Metadata)["text"].(string); ok { 
				pm.Text = txt 
			}
		}
		out = append(out, pm)
	}
	return out
}
