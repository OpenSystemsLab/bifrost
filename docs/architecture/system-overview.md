# 🌐 System Overview

Bifrost's high-level architecture designed for **enterprise-grade performance** with **10,000+ RPS throughput**, advanced concurrency management, and extensible plugin system.

---

## 🎯 Architecture Principles

| Principle                      | Implementation                                   | Benefit                                       |
| ------------------------------ | ------------------------------------------------ | --------------------------------------------- |
| **🔄 Asynchronous Processing** | Channel-based worker pools per provider          | High concurrency, no blocking operations      |
| **💾 Memory Pool Management**  | Object pooling for channels, messages, responses | Minimal GC pressure, sustained throughput     |
| **🏗️ Provider Isolation**      | Independent resources and workers per provider   | Fault tolerance, no cascade failures          |
| **🔌 Plugin-First Design**     | Middleware pipeline without core modifications   | Extensible business logic injection           |
| **⚡ Connection Optimization** | HTTP/2, keep-alive, intelligent pooling          | Reduced latency, optimal resource utilization |
| **📊 Built-in Observability**  | Native Prometheus metrics                        | Zero-dependency monitoring                    |

---

## 🏗️ High-Level Architecture

```mermaid
graph TB
    subgraph "Client Applications"
        WebApp[Web Applications]
        Mobile[Mobile Apps]
        Services[Microservices]
        CLI[CLI Tools]
    end

    subgraph "Transport Layer"
        HTTP[HTTP Transport<br/>:8080]
        SDK[Go SDK<br/>Direct Integration]
        Future[gRPC Transport<br/>Planned]
    end

    subgraph "Bifrost Core Engine"
        subgraph "Request Processing"
            Router[Request Router<br/>& Load Balancer]
            PluginPipeline[Plugin Pipeline<br/>Pre/Post Hooks]
            MCPManager[MCP Manager<br/>Tool Discovery]
        end

        subgraph "Memory Management"
            ChannelPool[Channel Pool<br/>Reusable Objects]
            MessagePool[Message Pool<br/>Request/Response]
            ResponsePool[Response Pool<br/>Result Objects]
        end

        subgraph "Worker Management"
            QueueManager[Queue Manager<br/>Request Distribution]
            WorkerPoolMgr[Worker Pool Manager<br/>Concurrency Control]
        end
    end

    subgraph "Provider Layer"
        subgraph "OpenAI Workers"
            OAI1[Worker 1]
            OAI2[Worker 2]
            OAIN[Worker N]
        end
        subgraph "Anthropic Workers"
            ANT1[Worker 1]
            ANT2[Worker 2]
            ANTN[Worker N]
        end
        subgraph "Other Providers"
            BED[Bedrock Workers]
            VER[Vertex Workers]
            MIS[Mistral Workers]
            AZU[Azure Workers]
        end
    end

    subgraph "External Systems"
        OPENAI_API[OpenAI API]
        ANTHROPIC_API[Anthropic API]
        BEDROCK_API[Amazon Bedrock]
        VERTEX_API[Google Vertex]
        MCP_SERVERS[MCP Servers<br/>Tools & Functions]
    end

    WebApp --> HTTP
    Mobile --> HTTP
    Services --> SDK
    CLI --> HTTP

    HTTP --> Router
    SDK --> Router
    Future --> Router

    Router --> PluginPipeline
    PluginPipeline --> MCPManager
    MCPManager --> QueueManager
    QueueManager --> WorkerPoolMgr

    WorkerPoolMgr --> ChannelPool
    WorkerPoolMgr --> MessagePool
    WorkerPoolMgr --> ResponsePool

    WorkerPoolMgr --> OAI1
    WorkerPoolMgr --> ANT1
    WorkerPoolMgr --> BED

    OAI1 --> OPENAI_API
    OAI2 --> OPENAI_API
    OAIN --> OPENAI_API

    ANT1 --> ANTHROPIC_API
    ANT2 --> ANTHROPIC_API
    ANTN --> ANTHROPIC_API

    BED --> BEDROCK_API
    VER --> VERTEX_API
    MIS --> ANTHROPIC_API

    MCPManager --> MCP_SERVERS
```

---

## ⚙️ Core Components

### **1. Transport Layer**

**Purpose:** Multiple interface options for different integration patterns

| Transport          | Use Case                                   | Performance | Integration Effort |
| ------------------ | ------------------------------------------ | ----------- | ------------------ |
| **HTTP Transport** | Microservices, web apps, language-agnostic | High        | Minimal (REST API) |
| **Go SDK**         | Go applications, maximum performance       | Maximum     | Low (Go package)   |
| **gRPC Transport** | Service mesh, type-safe APIs               | High        | Medium (protobuf)  |

**Key Features:**

- **OpenAPI Compatible** - Drop-in replacement for OpenAI/Anthropic APIs
- **Unified Interface** - Consistent API across all providers
- **Content Negotiation** - JSON, protobuf (planned)

### **2. Request Router & Load Balancer**

**Purpose:** Intelligent request distribution and provider selection

```mermaid
graph LR
    Request[Incoming Request] --> Router{Request Router}
    Router --> Provider[Provider Selection]
    Provider --> Key[API Key Selection<br/>Weighted Random]
    Key --> Worker[Worker Assignment]

    Router --> Fallback{Fallback Logic}
    Fallback --> Retry[Retry with<br/>Alternative Provider]
```

**Capabilities:**

- **Provider Selection** - Based on model availability and configuration
- **Load Balancing** - Weighted API key distribution
- **Fallback Chains** - Automatic provider switching on failures
- **Circuit Breaker** - Provider health monitoring and isolation

### **3. Plugin Pipeline**

**Purpose:** Extensible middleware for custom business logic

```mermaid
sequenceDiagram
    participant Request
    participant PreHooks
    participant Core
    participant PostHooks
    participant Response

    Request->>PreHooks: Raw Request
    PreHooks->>PreHooks: Auth, Rate Limiting, Transformation
    PreHooks->>Core: Modified Request
    Core->>Core: Provider Processing
    Core->>PostHooks: Raw Response
    PostHooks->>PostHooks: Logging, Caching, Analytics
    PostHooks->>Response: Final Response
```

**Plugin Types:**

- **Authentication** - API key validation, JWT verification
- **Rate Limiting** - Per-user, per-provider limits
- **Monitoring** - Request/response logging, metrics collection
- **Transformation** - Request/response modification
- **Caching** - Response caching strategies

### **4. MCP Manager**

**Purpose:** Model Context Protocol integration for external tools

**Architecture:**

```mermaid
graph TB
    MCPManager[MCP Manager] --> Discovery[Tool Discovery]
    MCPManager --> Registry[Tool Registry]
    MCPManager --> Execution[Tool Execution]

    Discovery --> STDIO[STDIO Servers]
    Discovery --> HTTP[HTTP Servers]
    Discovery --> SSE[SSE Servers]

    Registry --> Tools[Available Tools]
    Registry --> Filtering[Tool Filtering]

    Execution --> Invoke[Tool Invocation]
    Execution --> Results[Result Processing]
```

**Key Features:**

- **Dynamic Discovery** - Runtime tool discovery and registration
- **Multiple Protocols** - STDIO, HTTP, SSE support
- **Tool Filtering** - Request-level tool inclusion/exclusion
- **Async Execution** - Non-blocking tool invocation

### **5. Memory Management System**

**Purpose:** High-performance object pooling to minimize garbage collection

```go
// Simplified memory pool architecture
type MemoryManager struct {
    channelPool  sync.Pool  // Reusable communication channels
    messagePool  sync.Pool  // Request/response message objects
    responsePool sync.Pool  // Final response objects
    bufferPool   sync.Pool  // Byte buffers for network I/O
}
```

**Performance Impact:**

- **81% reduction** in processing overhead (11μs vs 59μs)
- **96% faster** queue wait times
- **Predictable latency** through object reuse

### **6. Worker Pool Manager**

**Purpose:** Provider-isolated concurrency with configurable resource limits

```mermaid
graph TB
    WorkerPoolMgr[Worker Pool Manager] --> Config[Configuration]
    WorkerPoolMgr --> Scheduling[Work Scheduling]
    WorkerPoolMgr --> Monitoring[Resource Monitoring]

    Config --> Concurrency[Concurrency Limits]
    Config --> BufferSize[Buffer Sizes]
    Config --> Timeouts[Timeout Settings]

    Scheduling --> Distribution[Work Distribution]
    Scheduling --> Queuing[Request Queuing]

    Monitoring --> Health[Worker Health]
    Monitoring --> Metrics[Performance Metrics]
```

**Isolation Benefits:**

- **Fault Tolerance** - Provider failures don't affect others
- **Resource Control** - Independent rate limiting per provider
- **Performance Tuning** - Provider-specific optimization
- **Scaling** - Independent scaling per provider load

---

## 🔄 Data Flow Architecture

### **Request Processing Pipeline**

```mermaid
sequenceDiagram
    participant Client
    participant Transport
    participant Router
    participant Plugin
    participant MCP
    participant Worker
    participant Provider

    Client->>Transport: HTTP/SDK Request
    Transport->>Router: Parse & Route
    Router->>Plugin: Pre-processing
    Plugin->>MCP: Tool Discovery
    MCP->>Worker: Queue Request
    Worker->>Provider: AI API Call
    Provider-->>Worker: AI Response
    Worker-->>MCP: Process Tools
    MCP-->>Plugin: Post-processing
    Plugin-->>Router: Final Response
    Router-->>Transport: Format Response
    Transport-->>Client: HTTP/SDK Response
```

### **Memory Object Lifecycle**

```mermaid
stateDiagram-v2
    [*] --> Pool: Object Creation
    Pool --> Acquired: Get from Pool
    Acquired --> Processing: Request Processing
    Processing --> Modified: Data Population
    Modified --> Cleanup: Reset State
    Cleanup --> Pool: Return to Pool
    Pool --> Garbage: Pool Full
    Garbage --> [*]: GC Collection
```

### **Concurrency Model**

```mermaid
graph TB
    subgraph "Request Concurrency"
        HTTP1[HTTP Request 1] --> Queue1[Provider Queue 1]
        HTTP2[HTTP Request 2] --> Queue1
        HTTP3[HTTP Request 3] --> Queue2[Provider Queue 2]

        Queue1 --> Worker1[Worker Pool 1<br/>OpenAI]
        Queue2 --> Worker2[Worker Pool 2<br/>Anthropic]

        Worker1 --> API1[OpenAI API]
        Worker2 --> API2[Anthropic API]
    end

    subgraph "Memory Concurrency"
        Pool[Object Pool] --> W1[Worker 1]
        Pool --> W2[Worker 2]
        Pool --> W3[Worker N]

        W1 --> Return1[Return Objects]
        W2 --> Return2[Return Objects]
        W3 --> Return3[Return Objects]

        Return1 --> Pool
        Return2 --> Pool
        Return3 --> Pool
    end
```

---

## 📊 Component Interactions

### **Configuration Hierarchy**

```mermaid
graph TB
    Global[Global Config] --> Provider[Provider Config]
    Provider --> Worker[Worker Config]
    Worker --> Request[Request Config]

    Global --> Pool[Pool Sizes]
    Global --> Plugins[Plugin Config]
    Global --> MCP[MCP Config]

    Provider --> Keys[API Keys]
    Provider --> Network[Network Config]
    Provider --> Fallbacks[Fallback Config]

    Worker --> Concurrency[Concurrency Limits]
    Worker --> Buffer[Buffer Sizes]
    Worker --> Timeout[Timeout Settings]
```

### **Error Propagation**

```mermaid
flowchart TD
    Error[Provider Error] --> Fallback{Fallback Available?}
    Fallback -->|Yes| NextProvider[Try Next Provider]
    Fallback -->|No| Plugin[Plugin Error Handler]

    NextProvider --> Success{Success?}
    Success -->|Yes| Response[Return Response]
    Success -->|No| Fallback

    Plugin --> Transform[Transform Error]
    Transform --> Client[Return to Client]
```

---

## 🚀 Scalability Architecture

### **Horizontal Scaling**

```mermaid
graph TB
    LoadBalancer[Load Balancer] --> B1[Bifrost Instance 1]
    LoadBalancer --> B2[Bifrost Instance 2]
    LoadBalancer --> BN[Bifrost Instance N]

    B1 --> Providers1[Provider APIs]
    B2 --> Providers2[Provider APIs]
    BN --> ProvidersN[Provider APIs]

    B1 --> SharedMCP[Shared MCP Servers]
    B2 --> SharedMCP
    BN --> SharedMCP
```

### **Vertical Scaling**

| Component            | Scaling Strategy        | Configuration              |
| -------------------- | ----------------------- | -------------------------- |
| **Memory Pools**     | Increase pool sizes     | `initial_pool_size: 25000` |
| **Worker Pools**     | More concurrent workers | `concurrency: 50`          |
| **Buffer Sizes**     | Larger request queues   | `buffer_size: 500`         |
| **Connection Pools** | More HTTP connections   | Provider-specific settings |

---

## 🔗 Related Architecture Documentation

- **[🔄 Request Flow](./request-flow.md)** - Detailed request processing pipeline
- **[⚙️ Concurrency Model](./concurrency.md)** - Worker pools and threading details
- **[🔌 Plugin System](./plugins.md)** - Plugin architecture and execution
- **[🛠️ MCP System](./mcp.md)** - Model Context Protocol implementation
- **[📊 Benchmarks](../benchmarks.md)** - Performance benchmarks and optimization strategies
- **[💡 Design Decisions](./design-decisions.md)** - Architecture rationale and trade-offs

---

**🎯 Next Step:** Understand how requests flow through the system in **[Request Flow](./request-flow.md)**.
