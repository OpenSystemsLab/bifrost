# 🔮 Google GenAI Compatible API

Complete guide to using Bifrost as a drop-in replacement for Google GenAI API with full compatibility and enhanced features.

> **💡 Quick Start:** Change `base_url` from `https://generativelanguage.googleapis.com` to `http://localhost:8080/genai` - that's it!

---

## 📋 Overview

Bifrost provides **100% Google GenAI API compatibility** with enhanced features:

- **Zero code changes** - Works with existing Google GenAI SDK applications
- **Same request/response formats** - Exact Google GenAI API specification
- **Enhanced capabilities** - Multi-provider fallbacks, MCP tools, monitoring
- **Full Gemini model support** - All Gemini models and features
- **Any provider under the hood** - Use any configured provider (Google, OpenAI, Anthropic, etc.)

**Endpoint:** `POST /genai/v1beta/models/{model}:generateContent`

> **🔄 Provider Flexibility:** While using Google GenAI SDK format, you can specify any model like `"gemini-pro"` (uses Google) or `"openai/gpt-4o-mini"` (uses OpenAI) - Bifrost will route to the appropriate provider automatically.

---

## 🔄 Quick Migration

### **Python (Google GenAI SDK)**

```python
import google.generativeai as genai

# Before - Direct Google GenAI
genai.configure(
    api_key="your-google-api-key",
    transport="rest"
)

# After - Via Bifrost
genai.configure(
    api_key="your-google-api-key",
    transport="rest",
    client_options={"api_endpoint": "http://localhost:8080/genai"}  # Only change this
)

# Everything else stays the same
model = genai.GenerativeModel('gemini-pro')
response = model.generate_content("Hello!")
```

### **JavaScript (Google GenAI SDK)**

```javascript
import { GoogleGenerativeAI } from "@google/generative-ai";

// Before - Direct Google GenAI
const genAI = new GoogleGenerativeAI(process.env.GOOGLE_API_KEY);

// After - Via Bifrost
const genAI = new GoogleGenerativeAI(process.env.GOOGLE_API_KEY, {
  baseUrl: "http://localhost:8080/genai", // Only change this
});

// Everything else stays the same
const model = genAI.getGenerativeModel({ model: "gemini-pro" });
const response = await model.generateContent("Hello!");
```

---

## 📊 Supported Features

### **✅ Fully Supported**

| Feature                 | Status     | Notes                           |
| ----------------------- | ---------- | ------------------------------- |
| **GenerateContent**     | ✅ Full    | All parameters supported        |
| **Multi-turn Chat**     | ✅ Full    | Conversation history            |
| **System Instructions** | ✅ Full    | Model behavior control          |
| **Vision/Multimodal**   | ✅ Full    | Images, videos, documents       |
| **Streaming**           | ⚠️ Planned | Currently returns full response |
| **Safety Settings**     | ✅ Full    | Content filtering               |
| **Generation Config**   | ✅ Full    | Temperature, top-k, etc.        |
| **Function Calling**    | ✅ Full    | Google + MCP tools              |

### **🚀 Enhanced Features**

| Feature                      | Enhancement              | Benefit               |
| ---------------------------- | ------------------------ | --------------------- |
| **Multi-provider Fallbacks** | Automatic failover       | Higher reliability    |
| **MCP Tool Integration**     | External tools available | Extended capabilities |
| **Load Balancing**           | Multiple API keys        | Better performance    |
| **Monitoring**               | Prometheus metrics       | Observability         |
| **Cross-provider Tools**     | Use with any provider    | Flexibility           |

---

## 🛠️ Request Examples

### **Basic Content Generation**

```bash
# Use Google provider
curl -X POST http://localhost:8080/genai/v1beta/models/gemini-pro:generateContent \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $GOOGLE_API_KEY" \
  -d '{
    "contents": [{
      "parts": [{"text": "What is the capital of France?"}]
    }]
  }'

# Use OpenAI provider via GenAI SDK format
curl -X POST http://localhost:8080/genai/v1beta/models/openai/gpt-4o-mini:generateContent \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $GOOGLE_API_KEY" \
  -d '{
    "contents": [{
      "parts": [{"text": "What is the capital of France?"}]
    }]
  }'
```

**Response:**

```json
{
  "candidates": [
    {
      "content": {
        "parts": [
          {
            "text": "The capital of France is Paris."
          }
        ],
        "role": "model"
      },
      "finishReason": "STOP",
      "index": 0
    }
  ],
  "usageMetadata": {
    "promptTokenCount": 7,
    "candidatesTokenCount": 7,
    "totalTokenCount": 14
  }
}
```

### **Multi-turn Conversation**

```bash
curl -X POST http://localhost:8080/genai/v1beta/models/gemini-pro:generateContent \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $GOOGLE_API_KEY" \
  -d '{
    "contents": [
      {
        "parts": [{"text": "Hello, who are you?"}],
        "role": "user"
      },
      {
        "parts": [{"text": "I am Gemini, a large language model."}],
        "role": "model"
      },
      {
        "parts": [{"text": "What can you help me with?"}],
        "role": "user"
      }
    ]
  }'
```

### **Vision/Multimodal**

```bash
curl -X POST http://localhost:8080/genai/v1beta/models/gemini-pro-vision:generateContent \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $GOOGLE_API_KEY" \
  -d '{
    "contents": [{
      "parts": [
        {"text": "What is in this image?"},
        {
          "inlineData": {
            "mimeType": "image/jpeg",
            "data": "/9j/4AAQSkZJRgABAQEAYABgAAD..."
          }
        }
      ]
    }]
  }'
```

### **Function Calling**

```bash
curl -X POST http://localhost:8080/genai/v1beta/models/gemini-pro:generateContent \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $GOOGLE_API_KEY" \
  -d '{
    "contents": [{
      "parts": [{"text": "What is the weather like in Paris?"}]
    }],
    "tools": [{
      "functionDeclarations": [{
        "name": "get_weather",
        "description": "Get weather information for a location",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {"type": "string", "description": "City name"}
          },
          "required": ["location"]
        }
      }]
    }]
  }'
```

**Response with Function Call:**

```json
{
  "candidates": [
    {
      "content": {
        "parts": [
          {
            "functionCall": {
              "name": "get_weather",
              "args": {
                "location": "Paris"
              }
            }
          }
        ],
        "role": "model"
      },
      "finishReason": "STOP"
    }
  ]
}
```

---

## 🔧 Advanced Usage

### **System Instructions**

```python
import google.generativeai as genai

genai.configure(
    api_key=google_api_key,
    client_options={"api_endpoint": "http://localhost:8080/genai"}
)

model = genai.GenerativeModel(
    'gemini-pro',
    system_instruction="You are a helpful assistant that answers questions about geography."
)

response = model.generate_content("What is the capital of France?")
```

### **Generation Configuration**

```python
import google.generativeai as genai

genai.configure(
    api_key=google_api_key,
    client_options={"api_endpoint": "http://localhost:8080/genai"}
)

generation_config = genai.types.GenerationConfig(
    candidate_count=1,
    max_output_tokens=1000,
    temperature=0.7,
    top_p=0.8,
    top_k=40,
    stop_sequences=["END"]
)

model = genai.GenerativeModel('gemini-pro')
response = model.generate_content(
    "Tell me a story",
    generation_config=generation_config
)
```

### **Safety Settings**

```python
import google.generativeai as genai

genai.configure(
    api_key=google_api_key,
    client_options={"api_endpoint": "http://localhost:8080/genai"}
)

safety_settings = [
    {
        "category": "HARM_CATEGORY_HARASSMENT",
        "threshold": "BLOCK_MEDIUM_AND_ABOVE"
    },
    {
        "category": "HARM_CATEGORY_HATE_SPEECH",
        "threshold": "BLOCK_MEDIUM_AND_ABOVE"
    }
]

model = genai.GenerativeModel('gemini-pro')
response = model.generate_content(
    "Your content here",
    safety_settings=safety_settings
)
```

### **Error Handling**

```python
import google.generativeai as genai
from google.api_core import exceptions

genai.configure(
    api_key=google_api_key,
    client_options={"api_endpoint": "http://localhost:8080/genai"}
)

try:
    model = genai.GenerativeModel('gemini-pro')
    response = model.generate_content("Hello!")
except exceptions.InvalidArgument as e:
    print(f"Invalid argument: {e}")
except exceptions.PermissionDenied as e:
    print(f"Permission denied: {e}")
except Exception as e:
    print(f"Other error: {e}")
```

---

## ⚡ Enhanced Features

### **Automatic MCP Tool Integration**

MCP tools are automatically available in GenAI-compatible requests:

```python
import google.generativeai as genai

genai.configure(
    api_key=google_api_key,
    client_options={"api_endpoint": "http://localhost:8080/genai"}
)

# No tool definitions needed - MCP tools auto-discovered
model = genai.GenerativeModel('gemini-pro')
response = model.generate_content(
    "List the files in the current directory and tell me about the project structure"
)

# Response may include automatic function calls
if response.candidates[0].content.parts[0].function_call:
    function_call = response.candidates[0].content.parts[0].function_call
    print(f"Called MCP tool: {function_call.name}")
```

### **Multi-provider Fallbacks**

Configure fallbacks in Bifrost config.json:

```json
{
  "providers": {
    "vertex": {
      "meta_config": {
        "project_id": "env.VERTEX_PROJECT_ID",
        "region": "us-central1"
      }
    },
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

Requests automatically fallback to OpenAI if Google fails:

```python
# This request tries Google first, falls back to OpenAI if needed
model = genai.GenerativeModel('gemini-pro')  # Will fallback to gpt-4o-mini
response = model.generate_content("Hello!")
```

### **Load Balancing**

Multiple Google Cloud projects automatically load balanced:

```json
{
  "providers": {
    "vertex": {
      "keys": [
        { "value": "env.GOOGLE_API_KEY_1", "weight": 0.7 },
        { "value": "env.GOOGLE_API_KEY_2", "weight": 0.3 }
      ]
    }
  }
}
```

---

## 🧪 Testing & Validation

### **Compatibility Testing**

Test your existing Google GenAI code with Bifrost:

```python
import google.generativeai as genai

def test_bifrost_compatibility():
    # Test with Bifrost
    genai.configure(
        api_key=google_api_key,
        client_options={"api_endpoint": "http://localhost:8080/genai"}
    )
    bifrost_model = genai.GenerativeModel('gemini-pro')

    # Test with direct Google GenAI (for comparison)
    genai.configure(
        api_key=google_api_key,
        client_options={}  # Reset to default
    )
    google_model = genai.GenerativeModel('gemini-pro')

    test_prompt = "Hello, test!"

    # Both should work identically
    bifrost_response = bifrost_model.generate_content(test_prompt)
    google_response = google_model.generate_content(test_prompt)

    # Compare response structure
    assert bifrost_response.candidates[0].content.parts[0].text is not None
    assert google_response.candidates[0].content.parts[0].text is not None

    print("✅ Bifrost Google GenAI compatibility verified")

test_bifrost_compatibility()
```

### **Function Calling Testing**

```python
import google.generativeai as genai

def test_function_calling():
    genai.configure(
        api_key=google_api_key,
        client_options={"api_endpoint": "http://localhost:8080/genai"}
    )

    # Define a test function
    def get_time():
        """Get current time"""
        return "2024-01-01 12:00:00"

    model = genai.GenerativeModel('gemini-pro')
    response = model.generate_content(
        "What time is it?",
        tools=[get_time]
    )

    # Should include function call
    if response.candidates[0].content.parts[0].function_call:
        print("✅ Function calling compatibility verified")
    else:
        print("⚠️ Function calling not triggered")

test_function_calling()
```

---

## 🔧 Configuration

### **Bifrost Config for Google GenAI**

```json
{
  "providers": {
    "vertex": {
      "keys": [],
      "meta_config": {
        "project_id": "env.VERTEX_PROJECT_ID",
        "region": "us-central1",
        "auth_credentials": "env.VERTEX_CREDENTIALS"
      },
      "network_config": {
        "default_request_timeout_in_seconds": 30,
        "max_retries": 2,
        "retry_backoff_initial_ms": 100,
        "retry_backoff_max_ms": 2000
      },
      "concurrency_and_buffer_size": {
        "concurrency": 3,
        "buffer_size": 10
      }
    }
  }
}
```

### **Environment Variables**

```bash
# Required for Google GenAI
export GOOGLE_API_KEY="your-api-key"

# OR for Vertex AI
export VERTEX_PROJECT_ID="your-project-id"
export VERTEX_CREDENTIALS="path/to/service-account.json"

# Optional - for enhanced features
export OPENAI_API_KEY="sk-..."  # For fallbacks
export BIFROST_LOG_LEVEL="info"
```

---

## 🚨 Common Issues & Solutions

### **Issue: "API Key not valid"**

**Problem:** Google API key not being passed correctly

**Solution:**

```python
# Ensure API key is properly set
import os
genai.configure(
    api_key=os.getenv("GOOGLE_API_KEY"),  # Explicit env var
    client_options={"api_endpoint": "http://localhost:8080/genai"}
)
```

### **Issue: "Model not found"**

**Problem:** Gemini model not available in your region/project

**Solution:** Configure fallback in config.json:

```json
{
  "providers": {
    "vertex": {
      "meta_config": {
        "project_id": "env.VERTEX_PROJECT_ID",
        "region": "us-central1"
      }
    },
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

### **Issue: "Authentication failed"**

**Problem:** Service account credentials not configured

**Solution:**

```bash
# Set up service account for Vertex AI
export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account.json"
export VERTEX_PROJECT_ID="your-project-id"
```

### **Issue: "Generation failed"**

**Problem:** Content blocked by safety filters

**Solution:**

```python
# Adjust safety settings
safety_settings = [
    {
        "category": "HARM_CATEGORY_HARASSMENT",
        "threshold": "BLOCK_ONLY_HIGH"  # Less restrictive
    }
]

response = model.generate_content(
    "Your content",
    safety_settings=safety_settings
)
```

---

## 📚 Related Documentation

- **[🔗 Drop-in Overview](./README.md)** - All provider integrations
- **[🌐 Endpoints](../endpoints.md)** - Complete API reference
- **[🔧 Configuration](../configuration/providers.md)** - Provider setup
- **[🔄 Migration Guide](./migration-guide.md)** - Step-by-step migration

> **🏛️ Architecture:** For Google GenAI integration implementation details, see [Architecture Documentation](../../../architecture/README.md).
