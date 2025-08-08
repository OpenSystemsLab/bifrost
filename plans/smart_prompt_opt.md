# Smart Prompt Optimization Plugin Plan

Goal
- Create a PreHook-first plugin that optimizes requests to improve quality and reduce token/cost usage before sending to providers.
- Optionally leverage Pinecone for retrieval-augmented compression/grounding and prompt memory.
- Keep it configurable and safe: no change to provider behavior; only request shaping and optional short-circuiting for cache-like behavior.

References
- Plugin interface from docs/usage/go-package/plugins.md
  - GetName() string
  - PreHook(ctx *context.Context, req *BifrostRequest) (*BifrostRequest, *PluginShortCircuit, error)
  - PostHook(ctx *context.Context, result *BifrostResponse, err *BifrostError) (*BifrostResponse, *BifrostError, error)

Plugin Name
- smart_prompt_opt

Primary Strategies (PreHook)
1) Input deduplication and trimming
   - Collapse repeated whitespace, remove duplicate paragraphs, strip boilerplate signatures.
   - Trim oversized history keeping most-recent conversational turns within a token budget target.
2) Instruction strengthening
   - Prepend/merge a concise system-style prefix for: role, constraints, output format, and reasoning depth.
   - Add explicit cost-aware guidance (be concise, avoid redundancy, cite sources if asked, etc.).
3) Provider/model-aware shaping
   - If model context window is known in config, enforce soft token budget for request content.
   - Autoset temperature/topP defaults via request overrides when not provided (configurable).
4) Semantic compression (optional)
   - Use an on-device embedder or provider-specific embedding to cluster and keep only representative chunks from long inputs.
   - If Pinecone is enabled, upsert embeddings for recurring user docs and perform top-k retrieval to replace raw long context with compressed summaries.
5) Prompt template library
   - Support named templates from config; choose template based on task hints (e.g., summarize, classify, generate code).
6) Short-circuit for empty/invalid input
   - If after normalization the prompt is empty or invalid by policy, return PluginShortCircuit with a friendly error.

Optional Pinecone Integration
- Use Pinecone for vector storage (namespaces per tenant/project).
- Server-side embeddings (no local embedding required)
  - Upsert by raw text: send text and metadata; Pinecone handles embedding creation.
  - Search by raw text: send the query text; Pinecone embeds on the fly and returns matches.
- Operations
  - Upsert: store user-supplied long context chunks (with metadata hash+timestamp, redacted if configured).
  - Query: when input exceeds budget, retrieve top-k similar chunks and replace with a compact summary of retrieved content.
- Failure Mode
  - If Pinecone errors, fail open: continue without RAG.

Config Structure (examples)
- SmartPromptOptConfig
  - Enabled: bool
  - TokenBudget: int (soft cap for combined instructions+messages)
  - MaxHistoryTurns: int
  - StrengthenInstructions: bool
  - InstructionPrefix: string (template with {task}, {format}, {style})
  - ProviderDefaults: map[string]ModelDefaults {Temperature, TopP}
  - SemanticCompression:
    - Enabled: bool
    - ChunkSize: int
    - CompressThresholdTokens: int
  - Pinecone:
    - Enabled: bool
    - ApiKeyEnv: string (e.g., PINECONE_API_KEY)
    - Environment: string
    - IndexName: string
    - Namespace: string
    - TopK: int
    - UpsertOnLargeInputs: bool
    - MetadataRedaction: { Emails: bool, PhoneNumbers: bool, SecretsPatterns: []string }
    - TimeoutMs: int

Data Flow (PreHook)
1) Normalize request (trim, dedupe, whitespace, collapse history)
2) Validate prompt; short-circuit if empty after cleaning
3) Apply instruction strengthening prefix (idempotent merge)
4) Apply provider defaults if missing (temperature/topP)
5) If token estimate > TokenBudget
   - If SemanticCompression.Enabled
     - Chunk input and optionally upsert raw text to Pinecone (server-side embedding).
     - Query Pinecone using the user query text; get topK relevant chunks.
     - Summarize retrieved chunks into compact bullets with citations/ids.
     - Replace long raw content with summary context block.
   - Else trim history/content by heuristic
6) Return modified request

PostHook (minimal)
- Optionally capture completion usage metrics from provider and log for future budget tuning.
- No response mutation unless configured (e.g., append citations from RAG metadata).

Edge Cases
- Streaming requests: PreHook safe (no stream consumption). PostHook should pass through untouched.
- Tool-calling requests: do not alter tool schemas; only adjust instructions.
- Multi-modal inputs: only operate on text fields; skip binary.

Telemetry
- Lightweight counters (optimizations applied, tokens estimated, compression applied, RAG hits).
- Respect a SamplingRate to avoid overhead.

Advanced Logging & Debugging
- Structured logs with correlation/request IDs carried via context.
- Log stages and decisions: normalization actions, tokens estimated, budget triggers, trimming/compression applied, Pinecone timings and status codes.
- Log levels: Debug (detailed), Info (high-level), Warn/Error (failures, fallbacks). Configurable via plugin config.
- Include feature flags snapshot and effective config in first log line of each request (redact secrets).
- Optional in-memory ring buffer for last N decisions to aid reproduction in tests.

Interfaces and Placement
- Package path: plugins/smartpromptopt
- Public constructor: NewSmartPromptOpt(cfg SmartPromptOptConfig) schemas.Plugin
- Internal helpers: tokenizer estimator interface, pinecone client (minimal), dedupe utilities.

Token Estimation
- Provide simple estimator (approx chars/4) and allow pluggable real tokenizer if available.

Security/Privacy
- Never send secrets to Pinecone; store only content hashes and text chunks intended for embedding.
- Namespacing by tenant/project to avoid cross-leakage.
- Configurable redaction of emails/IDs before upsert.

Deliverables
- Code for plugin with unit tests covering:
  - Deduplication, trimming, instruction strengthening idempotence
  - Budget enforcement with and without compression
  - Pinecone client fallback logic (upsert/search by text, fail-open paths)
- README: usage, config, examples
- Example wiring in docs

Additional Improvement Suggestions
- Intent detection from prompt to choose templates and adjust budgets dynamically.
- Automatic few-shot selection from Pinecone based on task similarity.
- Compact JSON or YAML output enforcement with schema-aware validators to reduce retries.
- Context window forecasting using provider metadata; preemptively refuse oversized attachments with guidance.
- User preference memory to persist concise style/format instructions per user.
- On-device redaction of PII before any external calls.
- Adaptive summarization aggressiveness based on past provider usage metrics.
- Cost guardrails: set a per-request max estimated cost; short-circuit with advisory if exceeded.
- Cache normalized prompts to avoid repeating normalization work across retries.

Usage Example
- client, _ := bifrost.Init(schemas.BifrostConfig{
    Account: &MyAccount{},
    Plugins: []schemas.Plugin{
      smartpromptopt.NewSmartPromptOpt(smartpromptopt.SmartPromptOptConfig{ ... }),
    },
  })

Next Steps
1) Scaffold plugin package and config structs
2) Implement normalization + strengthening + defaults
3) Add token estimator and budget enforcement
4) Add optional semantic compression + Pinecone client
5) Add tests and examples
