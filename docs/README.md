# Bifrost Documentation

Welcome to Bifrost - the unified AI model gateway that provides seamless integration with multiple AI providers through a single API.

## 🚀 Quick Start

Choose your preferred way to use Bifrost:

| Usage Mode            | Best For                            | Setup Time | Documentation                                           |
| --------------------- | ----------------------------------- | ---------- | ------------------------------------------------------- |
| **🔧 Go Package**     | Direct integration, maximum control | 2 minutes  | [📖 Go Package Guide](quickstart/go-package.md)         |
| **🌐 HTTP Transport** | Language-agnostic, microservices    | 30 seconds | [📖 HTTP Transport Guide](quickstart/http-transport.md) |

**New to Bifrost?** Start with [⚡ Quick Start](quickstart/) to get running with zero configuration in under 30 seconds.

---

## 🔄 Drop-in API Integrations

After you have set up Bifrost's HTTP transport, you can replace your existing AI provider calls with zero code changes. Bifrost provides 100% compatible endpoints for major AI providers:

| Integration                 | Compatible With                | Instant Migration                                                           |
| --------------------------- | ------------------------------ | --------------------------------------------------------------------------- |
| **🤖 OpenAI Compatible**    | OpenAI SDK, LangChain, LiteLLM | ✅ [Setup Guide](usage/http-transport/integrations/openai-compatible.md)    |
| **🧠 Anthropic Compatible** | Anthropic SDK, Claude API      | ✅ [Setup Guide](usage/http-transport/integrations/anthropic-compatible.md) |
| **🔍 GenAI Compatible**     | Google GenAI SDK               | ✅ [Setup Guide](usage/http-transport/integrations/genai-compatible.md)     |

```python
# Before (OpenAI)
client = openai.OpenAI(api_key="sk-...")

# After (Bifrost) - Same code, multi-provider benefits
client = openai.OpenAI(
    base_url="http://localhost:8080/openai",  # Only change needed
    api_key="dummy-api-key" # Handled by Bifrost
)
```

**Migration Benefits:**

- **🔄 Instant Fallbacks** - Never hit rate limits or downtime again
- **🌐 Multi-provider routing** - Route to any available provider, while keeping your existing codebase
- **🚀 Enhanced Features** - [MCP tools](mcp.md), [custom plugins](plugins.md), monitoring

[📖 **Complete Migration Guide**](usage/http-transport/integrations/migration-guide.md)

---

## 🎯 I Want To...

| Task                               | Go Here                                                                               |
| ---------------------------------- | ------------------------------------------------------------------------------------- |
| **Get started in 30 seconds**      | [⚡ Quick Start](quickstart/)                                                         |
| **Replace my OpenAI SDK calls**    | [🔄 OpenAI Integration](usage/http-transport/integrations/openai-compatible.md)       |
| **Replace my Anthropic SDK calls** | [🧠 Anthropic Integration](usage/http-transport/integrations/anthropic-compatible.md) |
| **Replace my GenAI SDK calls**     | [🔍 GenAI Integration](usage/http-transport/integrations/genai-compatible.md)         |
| **Use external tools with AI**     | [🛠️ MCP Integration](mcp.md)                                                          |
| **Add custom middleware**          | [🔌 Plugin System](plugins.md)                                                        |
| **Use Bifrost in my Go app**       | [🔧 Go Package Usage](usage/go-package/)                                              |
| **Configure via HTTP/JSON**        | [🌐 HTTP Transport Usage](usage/http-transport/)                                      |
| **Add fallback providers**         | [🔄 Providers](usage/providers.md)                                                    |
| **Understand the architecture**    | [🏛️ Architecture](architecture/)                                                      |
| **See practical examples**         | [💡 Examples](examples/)                                                              |
| **Deploy to production**           | [🚀 Production Guide](usage/http-transport/configuration/)                            |
| **Contribute to the project**      | [🤝 Contributing](contributing/)                                                      |

---

## 📚 Documentation Sections

### ⚡ [Quick Start](quickstart/)

Get running in under 30 seconds with step-by-step guides for both Go package and HTTP transport usage.

### 📖 [Usage](usage/)

Complete API reference and usage guides:

- **[🔧 Go Package](usage/go-package/)** - Direct Go integration
- **[🌐 HTTP Transport](usage/http-transport/)** - REST API with drop-in integrations

### 🏛️ [Architecture](architecture/)

Deep dive into Bifrost's design, performance, and internals:

- System overview and request flow
- Performance benchmarks and optimization
- Plugin and MCP architecture

### 💡 [Examples](examples/)

Practical, executable examples for common use cases:

- End-to-end tool calling
- MCP integration scenarios
- Production deployment patterns

### 🔧 Core Concepts

Universal concepts that apply to both Go package and HTTP transport:

- **[🔗 Providers](usage/providers.md)** - Multi-provider support and advanced configurations
- **[🔑 Key Management](usage/key-management.md)** - API key rotation and distribution
- **[⚡ Memory Management](usage/memory-management.md)** - Performance optimization
- **[🌐 Networking](usage/networking.md)** - Proxies, timeouts, and retries
- **[❌ Error Handling](usage/errors.md)** - Error types and troubleshooting

### 🤝 [Contributing](contributing/)

Help improve Bifrost for everyone:

- Development setup and guidelines
- Adding new providers and plugins
- Documentation standards

### 📊 Additional Resources

- **[📈 Benchmarks](benchmarks.md)** - Performance metrics and comparisons
- **[🔍 Troubleshooting](troubleshooting.md)** - Common issues and solutions
- **[❓ FAQ](faq.md)** - Frequently asked questions

---

## 🌟 What Makes Bifrost Special

- **🔄 Unified API** - One interface for OpenAI, Anthropic, Bedrock, and more
- **⚡ Intelligent Fallbacks** - Automatic failover between providers and models
- **🛠️ [MCP Integration](mcp.md)** - Enable AI models to use external tools (filesystem, web search, databases) through Model Context Protocol
- **🔌 [Extensible Plugins](plugins.md)** - Lightweight core with endless possibilities through custom middleware and request processing
- **🎯 Drop-in Compatibility** - Replace existing provider APIs without code changes
- **🚀 Production Ready** - Built for scale with comprehensive monitoring

---

## 💡 Quick Links

- **[⚡ 30-Second Setup](quickstart/)** - Get started immediately
- **[🔄 Migration Guide](usage/http-transport/integrations/migration-guide.md)** - Migrate from existing providers
- **[📊 Benchmarks](benchmarks.md)** - Performance benchmarks and optimization
- **[🛠️ Production Deployment](usage/http-transport/configuration/)** - Scale to production

---

**Need help?** Check our [❓ FAQ](faq.md) or [🔧 Troubleshooting](troubleshooting.md).

Built with ❤️ by the Maxim
