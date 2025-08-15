# 🛠️ Model Context Protocol (MCP)

Bifrost's Model Context Protocol integration enables AI models to seamlessly discover and execute external tools, transforming static chat models into dynamic, action-capable agents.

---

## 🎯 What is MCP?

**Model Context Protocol (MCP)** is a system that allows AI models to interact with external tools and services in real-time. Instead of being limited to text generation, AI models can:

- **Execute file operations** (read, write, list directories)
- **Search the web** and retrieve current information
- **Query databases** and APIs
- **Run custom business logic** through external tools
- **Interact with cloud services** and third-party systems

### Key Benefits

| Feature                    | Description                                     |
| -------------------------- | ----------------------------------------------- |
| **🔍 Dynamic Discovery**   | Tools are discovered at runtime, not hardcoded  |
| **🛡️ Client-Side Control** | Bifrost manages all tool execution for security |
| **🌐 Multiple Protocols**  | InProcess, STDIO, HTTP, and SSE connections     |
| **🎯 Granular Filtering**  | Control tool availability per request           |
| **⚡ High Performance**    | Async execution with minimal overhead           |

---

## 🚀 Quick Example

```bash
# Configure MCP in your Bifrost setup
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai/gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "List the files in the /tmp directory"}
    ]
  }'

# Response includes automatic tool usage
{
  "choices": [{
    "message": {
      "tool_calls": [{
        "id": "call_123",
        "function": {
          "name": "list_files",
          "arguments": "{\"path\": \"/tmp\"}"
        }
      }]
    }
  }]
}
```

---

## 📚 Documentation

### 🏗️ Architecture & Design

**[Architecture Documentation](architecture/mcp.md)**

- Deep dive into MCP system design
- Connection types and protocols
- Tool discovery and registration
- Security and filtering mechanisms
- Performance considerations

### 🔧 Go Package Integration

**[Go Package MCP Guide](usage/go-package/mcp.md)**

- MCP configuration in Go applications
- Automatic and manual tool execution
- Custom tool registration
- Error handling and best practices
- Complete code examples

### 🌐 HTTP Transport Configuration

**[HTTP Transport MCP Configuration](usage/http-transport/configuration/mcp.md)**

- JSON configuration for HTTP service
- Multi-turn tool calling examples
- Docker and environment setup
- Popular MCP servers integration
- Testing and troubleshooting

---

## 🔌 Connection Types

| Type         | Use Case                          | Latency           | Security     | Scalability   | Availability      |
| ------------ | --------------------------------- | ----------------- | ------------ | ------------- | ----------------- |
| **InProcess**| Embedded tools, testing           | Lowest (~0.1ms)   | Highest      | Single process| Go package only   |
| **STDIO**    | Local tools, CLI utilities        | Low (~1-10ms)     | High         | Single server | All transports    |
| **HTTP**     | Web APIs, microservices           | Network dependent | Configurable | High          | All transports    |
| **SSE**      | Real-time streams, monitoring     | Variable          | HTTP-level   | Medium        | All transports    |

---

## 🛠️ Popular Use Cases

- **📁 File Management**: Read, write, and organize files
- **🔍 Web Search**: Get current information from the internet
- **💾 Database Operations**: Query and update data stores
- **🔧 DevOps Tools**: Deploy, monitor, and manage infrastructure
- **📊 Analytics**: Generate reports and insights
- **🤝 Integration**: Connect with CRMs, ERPs, and business systems

---

## 🎉 Getting Started

1. **[📖 Quick Start Guide](quickstart/README.md)** - 30-second MCP setup
2. **[🎯 Choose Your Integration](usage/README.md)** - Go package vs HTTP transport
3. **[🏗️ Understand the Architecture](architecture/mcp.md)** - System design deep dive

> **💡 New to Bifrost?** Start with the [main documentation](README.md) to understand Bifrost's core concepts before diving into MCP integration.
