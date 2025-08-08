# smart_prompt_opt plugin

A PreHook-first plugin that optimizes inputs for quality and cost. Optionally uses Pinecone with server-side embeddings (text upsert/search) for semantic compression when inputs exceed a token budget.

Key features
- Normalization: whitespace collapse, paragraph deduplication, history trimming
- Instruction strengthening: prepend concise system-style prefix (configurable)
- Provider defaults: temperature/top_p overrides per provider+model
- Budget enforcement: estimate tokens and trim or compress
- Pinecone integration: upsert/search by raw text (no local embeddings)
- Advanced logging: structured logs with correlation IDs and optional in-memory ring buffer

Config (example)
- Enabled: true
- TokenBudget: 2000
- MaxHistoryTurns: 12
- StrengthenInstructions: true
- InstructionPrefix: "You are a concise assistant. Answer tersely, avoid redundancy."
- ProviderDefaults: {"openai:gpt-4o": {Temperature: 0.2}}
- SemanticCompression: { Enabled: true, ChunkSize: 2000, CompressThresholdTokens: 1800 }
- Pinecone: { Enabled: true, ApiKeyEnv: "PINECONE_API_KEY", Environment: "us-east-1", IndexName: "bifrost-smart", Namespace: "default", TopK: 5, UpsertOnLargeInputs: true, TimeoutMs: 2000 }

Usage
- p, _ := smartpromptopt.NewSmartPromptOpt(smartpromptopt.Config{ /* ... */ })
- client, _ := bifrost.Init(schemas.BifrostConfig{ Account: &MyAccount{}, Plugins: []schemas.Plugin{ p } })

Installation
- Add dependency: go get github.com/pinecone-io/go-pinecone/v4/pinecone
- Import the plugin: import smartpromptopt "github.com/maximhq/bifrost/plugins/smartpromptopt"

Notes
- The plugin now includes a Pinecone SDK client (pinecone_client.go) that uses github.com/pinecone-io/go-pinecone/v4/pinecone
- Set your Pinecone API key in the environment variable specified by ApiKeyEnv (e.g., PINECONE_API_KEY)
- The SDK client assumes your Pinecone index supports text-based operations. Adjust the implementation based on your index configuration:
  - For server-side embeddings: Use the inference API or a text field that triggers embedding
  - For client-side embeddings: Add an embedding service before upsert/query operations
- To use the real Pinecone client, enable the SDK-backed implementation:
- Add the dependency: go get github.com/pinecone-io/go-pinecone/v4/pinecone
  - Build with tag: -tags pinecone
  - Ensure PINECONE_API_KEY (or your configured ApiKeyEnv) is set in the environment
- Pinecone integration here assumes server-side embeddings available by text upsert/search. Replace the stub client with real HTTP calls if you do not want to use the SDK.
- If Pinecone is disabled or fails, the plugin falls back to heuristic trimming to stay within budget.
- Use ProviderDefaults to set per-provider/model defaults for temperature/top_p.
- Pinecone calls are stubbed in this scaffold. Implement HTTP calls as needed.
- Errors in PreHook should prefer fail-open (no RAG) to avoid blocking the main request path.
