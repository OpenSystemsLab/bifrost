package smartpromptopt

import (
	"fmt"

	"github.com/maximhq/bifrost/core/schemas"
)

// Default configuration values
const (
	DefaultChunkSize     = 2000
	DefaultTopK          = 5
	DefaultTimeoutMs     = 2000
	DefaultTokensPerChar = 4 // Crude approximation: 1 token ≈ 4 characters
)

type ModelDefaults struct {
	Temperature *float64
	TopP        *float64
}

type SemanticCompressionConfig struct {
	Enabled                 bool
	ChunkSize               int
	CompressThresholdTokens int
}

type PineconeConfig struct {
	Enabled             bool
	ApiKeyEnv           string
	Environment         string
	IndexName           string
	Namespace           string
	TopK                int
	UpsertOnLargeInputs bool
	MetadataRedaction   struct {
		Emails          bool
		PhoneNumbers    bool
		SecretsPatterns []string
	}
	TimeoutMs int
}

type AdvancedLoggingConfig struct {
	RingSize int
}

type Config struct {
	Enabled                bool
	TokenBudget            int
	MaxHistoryTurns        int
	StrengthenInstructions bool
	InstructionPrefix      string
	ProviderDefaults       map[string]ModelDefaults // key: provider+":"+model
	SemanticCompression    SemanticCompressionConfig
	Pinecone               PineconeConfig
	AdvancedLogging        AdvancedLoggingConfig
	Logger                 schemas.Logger
}

// Validate validates the configuration.
// It returns an error if the configuration is invalid.
func (c *Config) Validate() error {
	if c.SemanticCompression.Enabled {
		if c.SemanticCompression.ChunkSize <= 0 {
			c.SemanticCompression.ChunkSize = DefaultChunkSize
		}
		if c.Pinecone.Enabled {
			if c.Pinecone.TopK <= 0 {
				c.Pinecone.TopK = DefaultTopK
			}
			if c.Pinecone.TimeoutMs <= 0 {
				c.Pinecone.TimeoutMs = DefaultTimeoutMs
			}
			if c.Pinecone.IndexName == "" {
				return fmt.Errorf("pinecone.index_name is required when Pinecone is enabled")
			}
		}
	}
	return nil
}
