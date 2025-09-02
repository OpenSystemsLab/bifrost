package tests

import (
	"os"
	"testing"

	"github.com/maximhq/bifrost/tests/core-providers/config"

	"github.com/maximhq/bifrost/core/schemas"
)

func TestBedrock(t *testing.T) {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		t.Skip("Skipping Bedrock embedding: AWS credentials not set")
	}

	client, ctx, cancel, err := config.SetupTest()
	if err != nil {
		t.Fatalf("Error initializing test setup: %v", err)
	}
	defer cancel()
	defer client.Cleanup()

	testConfig := config.ComprehensiveTestConfig{
		Provider:  schemas.Bedrock,
		ChatModel: "anthropic.claude-3-sonnet-20240229-v1:0",
		TextModel: "", // Bedrock Claude doesn't support text completion
		EmbeddingModel: "amazon.titan-embed-text-v2:0",
		Scenarios: config.TestScenarios{
			TextCompletion:        false, // Not supported for Claude
			SimpleChat:            true,
			ChatCompletionStream:  true,
			MultiTurnConversation: true,
			ToolCalls:             true,
			MultipleToolCalls:     true,
			End2EndToolCalling:    true,
			AutomaticFunctionCall: true,
			ImageURL:              false,
			ImageBase64:           true,
			MultipleImages:        false,
			CompleteEnd2End:       true,
			ProviderSpecific:      true,
			Embedding:             true,
		},
	}

	runAllComprehensiveTests(t, client, ctx, testConfig)
}
