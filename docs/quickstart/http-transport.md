# 🌐 HTTP Transport Quick Start

Get Bifrost running as an HTTP API in 30 seconds using Docker. Perfect for any programming language.

## ⚡ 30-Second Setup

### 1. Create `config.json`

This file should contain your provider settings and API keys.

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
    }
  }
}
```

### 2. Set Up Your Environment

Add your environment variable to the session.

```bash
export OPENAI_API_KEY="your-openai-api-key"
```

### 3. Start the Bifrost HTTP Server

You can run using Docker or Go binary.

```bash
# Docker
docker pull maximhq/bifrost
docker run -p 8080:8080 \
  -v $(pwd)/config.json:/app/config/config.json \
  -e OPENAI_API_KEY \
  maximhq/bifrost

# OR Go Binary (Make sure Go in your PATH)
go install github.com/maximhq/bifrost/transports/bifrost-http@latest
bifrost-http -config config.json -port 8080
```

### 4. Test the API

```bash
# Make your first request
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello, Bifrost!"}]
  }'
```

**🎉 Success!** You should see an AI response in JSON format.

> **📋 Note**: All Bifrost responses follow OpenAI's response structure, regardless of the underlying provider. This ensures consistent integration across different AI providers.

---

## 🔄 Drop-in Integrations (Zero Code Changes!)

**Already using OpenAI, Anthropic, or Google GenAI?** Get instant benefits with **zero code changes**:

### 🤖 **OpenAI SDK Replacement**

```python
# Before
from openai import OpenAI
client = OpenAI(api_key="your-key")

# After - Just change base_url!
from openai import OpenAI
client = OpenAI(
    api_key="dummy",  # Not used
    base_url="http://localhost:8080/openai"
)

# All your existing code works unchanged! ✨
response = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

### 🧠 **Anthropic SDK Replacement**

```python
# Before
from anthropic import Anthropic
client = Anthropic(api_key="your-key")

# After - Just change base_url!
from anthropic import Anthropic
client = Anthropic(
    api_key="dummy",  # Not used
    base_url="http://localhost:8080/anthropic"
)

# All your existing code works unchanged! ✨
```

### 🔍 **Google GenAI Replacement**

```python
# Before
from google import genai
client = genai.Client(api_key="your-key")

# After - Just change base_url!
from google import genai
client = genai.Client(
    api_key="dummy",  # Not used
    http_options=genai.types.HttpOptions(
        base_url="http://localhost:8080/genai"
    )
)

# All your existing code works unchanged! ✨
```

---

## 🚀 Next Steps (2 minutes each)

### **🔗 Add Multiple Providers**

```bash
# Create config.json
echo '{
  "providers": {
    "openai": {
      "keys": [{"value": "env.OPENAI_API_KEY", "models": ["gpt-4o-mini"], "weight": 1.0}]
    },
    "anthropic": {
      "keys": [{"value": "env.ANTHROPIC_API_KEY", "models": ["claude-3-sonnet-20240229"], "weight": 1.0}]
    }
  }
}' > config.json

# Set environment variables
export ANTHROPIC_API_KEY="your-anthropic-key"

# Start with config
docker run -p 8080:8080 \
  -v $(pwd)/config.json:/app/config/config.json \
  -e OPENAI_API_KEY -e ANTHROPIC_API_KEY \
  maximhq/bifrost
```

### **⚡ Test Different Providers**

```bash
# Use OpenAI
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"provider": "openai", "model": "gpt-4o-mini", "messages": [{"role": "user", "content": "Hello from OpenAI!"}]}'

# Use Anthropic
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"provider": "anthropic", "model": "claude-3-sonnet-20240229", "messages": [{"role": "user", "content": "Hello from Anthropic!"}]}'
```

### **🔄 Add Automatic Fallbacks**

```bash
# Request with fallback
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "model": "gpt-4o-mini",
    "messages": [{"role": "user", "content": "Hello!"}],
    "fallbacks": [{"provider": "anthropic", "model": "claude-3-sonnet-20240229"}]
  }'
```

---

## 🔗 Language Examples

### Python

```python
import requests

response = requests.post(
    "http://localhost:8080/v1/chat/completions",
    json={
        "provider": "openai",
        "model": "gpt-4o-mini",
        "messages": [{"role": "user", "content": "Hello from Python!"}]
    }
)
print(response.json())
```

### JavaScript/Node.js

```javascript
const response = await fetch("http://localhost:8080/v1/chat/completions", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    provider: "openai",
    model: "gpt-4o-mini",
    messages: [{ role: "user", content: "Hello from Node.js!" }],
  }),
});
console.log(await response.json());
```

### Go

```go
response, err := http.Post(
    "http://localhost:8080/v1/chat/completions",
    "application/json",
    strings.NewReader(`{
        "provider": "openai",
        "model": "gpt-4o-mini",
        "messages": [{"role": "user", "content": "Hello from Go!"}]
    }`)
)
```

---

## 🔧 Setup Methods Comparison

| Method        | Pros                                            | Use When                         |
| ------------- | ----------------------------------------------- | -------------------------------- |
| **Docker**    | No Go installation needed, isolated environment | Production, CI/CD, quick testing |
| **Go Binary** | Direct execution, easier debugging              | Development, custom builds       |

Both methods require the same `config.json` file and environment variables.

---

## 💬 Need Help?

**🔗 [Join our Discord](https://discord.gg/qPaAuTCv)** for real-time setup assistance and HTTP integration support!

---

## 📚 Learn More

| What You Want                  | Where to Go                                                | Time      |
| ------------------------------ | ---------------------------------------------------------- | --------- |
| **Drop-in integrations guide** | [🔄 Integrations](../usage/http-transport/integrations/)   | 5 min     |
| **Complete HTTP setup**        | [📖 HTTP Transport Usage](../usage/http-transport/)        | 10 min    |
| **Production configuration**   | [🔧 Configuration](../usage/http-transport/configuration/) | 15 min    |
| **All endpoints**              | [🎯 API Endpoints](../usage/http-transport/endpoints.md)   | Reference |
| **OpenAPI specification**      | [📊 OpenAPI Spec](../usage/http-transport/openapi.json)    | Reference |

---

## 🔄 Prefer Go Package?

If you're building a Go application and want direct integration, try the **[Go Package Quick Start](go-package.md)** instead.

---

## 💡 Why HTTP Transport?

- ✅ **Language agnostic** - Use from Python, Node.js, PHP, etc.
- ✅ **Drop-in replacement** - Zero code changes for existing apps
- ✅ **OpenAI compatible** - All responses follow OpenAI structure
- ✅ **Microservices ready** - Centralized AI gateway
- ✅ **Production features** - Health checks, metrics, monitoring

**🎯 Ready for production? Check out [Complete HTTP Usage Guide](../usage/http-transport/) →**
