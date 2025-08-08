package smartpromptopt

import (
	"fmt"
	"time"

	"github.com/maximhq/bifrost/core/schemas"
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
	Enabled           bool
	ApiKeyEnv         string
	Environment       string
	IndexName         string
	Namespace         string
	TopK              int
	UpsertOnLargeInputs bool
	MetadataRedaction struct {
		Emails        bool
		PhoneNumbers  bool
		SecretsPatterns []string
	}
	TimeoutMs int
}

type AdvancedLoggingConfig struct {
	RingSize int
}

type Config struct {
	Enabled               bool
	TokenBudget           int
	MaxHistoryTurns       int
	StrengthenInstructions bool
	InstructionPrefix     string
	ProviderDefaults      map[string]ModelDefaults // key: provider+":"+model
	SemanticCompression   SemanticCompressionConfig
	Pinecone              PineconeConfig
	AdvancedLogging       AdvancedLoggingConfig
	Logger                schemas.Logger
}

func (c *Config) Validate() error {
	if c.SemanticCompression.Enabled {
		if c.SemanticCompression.ChunkSize <= 0 {
			c.SemanticCompression.ChunkSize = 2000
		}
		if c.Pinecone.Enabled {
			if c.Pinecone.TopK <= 0 { c.Pinecone.TopK = 5 }
			if c.Pinecone.TimeoutMs <= 0 { c.Pinecone.TimeoutMs = int((2 * time.Second).Milliseconds()) }
			if c.Pinecone.IndexName == "" {
				return fmt.Errorf("pinecone.index_name is required when Pinecone is enabled")
			}
		}
	}
	return nil
}
