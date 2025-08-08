# Smart Prompt Optimization Plugin (smart_prompt_opt)

This plugin optimizes prompts to improve quality and reduce cost before requests reach providers. It uses PreHook to normalize inputs, strengthen instructions, enforce a token budget, and optionally apply Pinecone-backed semantic compression via text upsert/search (no local embeddings needed).

Install
- The plugin is part of the repo under plugins/smartpromptopt.

Quick usage

```go
package main

import (
    bifrost "github.com/maximhq/bifrost/core"
    "github.com/maximhq/bifrost/core/schemas"
    smartpromptopt "github.com/maximhq/bifrost/plugins/smartpromptopt"
)

func main() {
    p, _ := smartpromptopt.NewSmartPromptOpt(smartpromptopt.Config{
        Enabled: true,
        TokenBudget: 2000,
        MaxHistoryTurns: 12,
        StrengthenInstructions: true,
        InstructionPrefix: "You are a concise assistant. Answer tersely and avoid redundancy.",
        SemanticCompression: smartpromptopt.SemanticCompressionConfig{
            Enabled: true,
            ChunkSize: 2000,
            CompressThresholdTokens: 1800,
        },
        Pinecone: smartpromptopt.PineconeConfig{
            Enabled: true,
            ApiKeyEnv: "PINECONE_API_KEY",
            Environment: "us-east-1",
            IndexName: "bifrost-smart",
            Namespace: "default",
            TopK: 5,
            UpsertOnLargeInputs: true,
            TimeoutMs: 2000,
        },
    })

    client, _ := bifrost.Init(schemas.BifrostConfig{
        Account: &MyAccount{},
        Plugins: []schemas.Plugin{ p },
    })
    defer client.Cleanup()
}
```

Notes
- Pinecone integration here assumes server-side embeddings available by text upsert/search. Replace the stub client with real HTTP calls to Pinecone's API when deploying.
- If Pinecone is disabled or fails, the plugin falls back to heuristic trimming to stay within budget.
- Use ProviderDefaults to set per-provider/model defaults for temperature/top_p.
