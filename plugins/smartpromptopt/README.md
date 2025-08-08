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

Notes
- Pinecone calls are stubbed in this scaffold. Implement HTTP calls as needed.
- Errors in PreHook should prefer fail-open (no RAG) to avoid blocking the main request path.
