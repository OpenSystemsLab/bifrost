# 🌐 HTTP Transport

Complete guide to using Bifrost as an HTTP API service for multi-provider AI access, drop-in integrations, and production deployment.

> **💡 Quick Start:** See the [30-second setup](../../quickstart/http-transport.md) to get the HTTP service running quickly.

---

## 📋 HTTP Transport Overview

Bifrost HTTP transport provides a REST API service for:

- **Multi-provider access** through unified endpoints
- **Drop-in replacements** for OpenAI, Anthropic, Google GenAI APIs
- **Language-agnostic integration** with any HTTP client
- **Production-ready deployment** with monitoring and scaling
- **MCP tool execution** via HTTP endpoints

```bash
# Start Bifrost HTTP service
docker run -p 8080:8080 \
  -v $(pwd)/config.json:/app/config/config.json \
  -e OPENAI_API_KEY \
  maximhq/bifrost

# Make requests to any provider
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"provider": "openai", "model": "gpt-4o-mini", "messages": [...]}'
```

---

## 🚀 Core Features

### **Unified API Endpoints**

| Endpoint                    | Purpose            | Documentation                     |
| --------------------------- | ------------------ | --------------------------------- |
| `POST /v1/chat/completions` | Chat conversations | [Endpoints Guide](./endpoints.md) |
| `POST /v1/text/completions` | Text generation    | [Endpoints Guide](./endpoints.md) |
| `POST /v1/mcp/tool/execute` | Tool execution     | [Endpoints Guide](./endpoints.md) |
| `GET /metrics`              | Prometheus metrics | [Endpoints Guide](./endpoints.md) |

### **Drop-in API Compatibility**

| Provider         | Endpoint                            | Compatibility                                                  |
| ---------------- | ----------------------------------- | -------------------------------------------------------------- |
| **OpenAI**       | `POST /openai/v1/chat/completions`  | [OpenAI Compatible](./integrations/openai-compatible.md)       |
| **Anthropic**    | `POST /anthropic/v1/messages`       | [Anthropic Compatible](./integrations/anthropic-compatible.md) |
| **Google GenAI** | `POST /genai/v1beta/models/{model}` | [GenAI Compatible](./integrations/genai-compatible.md)         |

> **📖 Migration:** See [Migration Guide](./integrations/migration-guide.md) for step-by-step migration from existing providers.

---

## ⚙️ Configuration

### **Core Configuration Files**

| Component                                        | Configuration                   | Time to Setup |
| ------------------------------------------------ | ------------------------------- | ------------- |
| **[🔧 Providers](./configuration/providers.md)** | API keys, models, fallbacks     | 5 min         |
| **[🛠️ MCP Integration](./configuration/mcp.md)** | Tool servers and connections    | 10 min        |
| **[🔌 Plugins](./configuration/plugins.md)**     | Custom middleware (coming soon) | 5 min         |

### **Quick Configuration Example**

```json
{
  "providers": {
    "openai": {
      "keys": [
        {
          "value": "env.OPENAI_API_KEY",
          "models": ["gpt-4o-mini"],
          "weight": 1.0
        }
      ]
    },
    "anthropic": {
      "keys": [
        {
          "value": "env.ANTHROPIC_API_KEY",
          "models": ["claude-3-sonnet-20240229"],
          "weight": 1.0
        }
      ]
    }
  },
  "mcp": {
    "client_configs": [
      {
        "name": "filesystem",
        "connection_type": "stdio",
        "stdio_config": {
          "command": "npx",
          "args": ["-y", "@modelcontextprotocol/server-filesystem"]
        }
      }
    ]
  }
}
```

---

## 🔗 Integration Patterns

### **"I want to..."**

| Goal                       | Integration Type       | Guide                                                          |
| -------------------------- | ---------------------- | -------------------------------------------------------------- |
| **Replace OpenAI API**     | Drop-in replacement    | [OpenAI Compatible](./integrations/openai-compatible.md)       |
| **Replace Anthropic API**  | Drop-in replacement    | [Anthropic Compatible](./integrations/anthropic-compatible.md) |
| **Use with existing SDKs** | Change base URL only   | [Migration Guide](./integrations/migration-guide.md)           |
| **Add multiple providers** | Provider configuration | [Providers Config](./configuration/providers.md)               |
| **Add external tools**     | MCP integration        | [MCP Config](./configuration/mcp.md)                           |
| **Custom monitoring**      | Plugin configuration   | [Plugins Config](./configuration/plugins.md)                   |
| **Production deployment**  | Docker + config        | [Deployment Guide](../../quickstart/http-transport.md)         |

### **Language Examples**

<details>
<summary><strong>Python (OpenAI SDK)</strong></summary>

```python
from openai import OpenAI

# Change base URL to use Bifrost
client = OpenAI(
    base_url="http://localhost:8080/openai",  # Point to Bifrost
    api_key="your-openai-key"
)

# Use normally - Bifrost handles provider routing
response = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

</details>

<details>
<summary><strong>JavaScript/Node.js</strong></summary>

```javascript
import OpenAI from "openai";

const openai = new OpenAI({
  baseURL: "http://localhost:8080/openai", // Point to Bifrost
  apiKey: process.env.OPENAI_API_KEY,
});

const response = await openai.chat.completions.create({
  model: "gpt-4o-mini",
  messages: [{ role: "user", content: "Hello!" }],
});
```

</details>

<details>
<summary><strong>cURL</strong></summary>

```bash
# Direct Bifrost API
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello!"}],
    "fallbacks": [{"provider": "anthropic", "model": "claude-3-sonnet-20240229"}]
  }'

# OpenAI-compatible endpoint
curl -X POST http://localhost:8080/openai/v1/chat/completions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

</details>

---

## 🚀 Deployment Options

### **Docker (Recommended)**

```bash
# Quick start
docker run -p 8080:8080 \
  -v $(pwd)/config.json:/app/config/config.json \
  -e OPENAI_API_KEY \
  -e ANTHROPIC_API_KEY \
  maximhq/bifrost

# Production with custom settings
docker run -p 8080:8080 \
  -v $(pwd)/config.json:/app/config/config.json \
  -v $(pwd)/logs:/app/logs \
  -e OPENAI_API_KEY \
  -e ANTHROPIC_API_KEY \
  maximhq/bifrost \
  -pool-size 500 \
  -drop-excess-requests
```

### **Binary Deployment**

```bash
# Install
go install github.com/maximhq/bifrost/transports/bifrost-http@latest

# Run
bifrost-http \
  -config config.json \
  -port 8080 \
  -pool-size 300 \
  -plugins maxim
```

### **Kubernetes**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bifrost
spec:
  replicas: 3
  selector:
    matchLabels:
      app: bifrost
  template:
    metadata:
      labels:
        app: bifrost
    spec:
      containers:
        - name: bifrost
          image: maximhq/bifrost:latest
          ports:
            - containerPort: 8080
          env:
            - name: OPENAI_API_KEY
              valueFrom:
                secretKeyRef:
                  name: ai-keys
                  key: openai
          volumeMounts:
            - name: config
              mountPath: /app/config
      volumes:
        - name: config
          configMap:
            name: bifrost-config
```

---

## 📊 Monitoring and Observability

### **Built-in Metrics**

```bash
# Prometheus metrics endpoint
curl http://localhost:8080/metrics

# Key metrics available:
# - bifrost_requests_total{provider, model, status}
# - bifrost_request_duration_seconds{provider, model}
# - bifrost_tokens_total{provider, model, type}
# - bifrost_errors_total{provider, error_type}
```

### **Health Checks**

```bash
# Basic health check
curl http://localhost:8080/v1/chat/completions \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"provider":"openai","model":"gpt-4o-mini","messages":[{"role":"user","content":"test"}]}'
```

---

## 📚 Complete Documentation

### **📖 API Reference**

- **[🌐 Endpoints](./endpoints.md)** - Complete API endpoint documentation
- **[📋 OpenAPI Spec](./openapi.json)** - Machine-readable API specification

### **⚙️ Configuration Guides**

- **[🔧 Provider Setup](./configuration/providers.md)** - Configure AI providers and keys
- **[🛠️ MCP Integration](./configuration/mcp.md)** - Setup external tool integration
- **[🔌 Plugin System](./configuration/plugins.md)** - Configure custom middleware

### **🔗 Integration Guides**

- **[📱 Drop-in Integrations](./integrations/README.md)** - Overview of API compatibility
- **[🔄 Migration Guide](./integrations/migration-guide.md)** - Migrate from existing providers
- **[⚙️ SDK Examples](./integrations/)** - Language-specific integration examples

---

## 🎯 Next Steps

1. **[⚡ Quick Setup](../../quickstart/http-transport.md)** - Get Bifrost HTTP running in 30 seconds
2. **[🔧 Configure Providers](./configuration/providers.md)** - Add your AI provider credentials
3. **[🔗 Choose Integration](./integrations/README.md)** - Pick drop-in replacement or unified API
4. **[🚀 Deploy to Production](../../quickstart/http-transport.md#production-deployment)** - Scale for production workloads

> **🏛️ Architecture:** For HTTP transport design and performance details, see [Architecture Documentation](../../architecture/README.md).
