# ⚡ Memory Management & Performance Tuning

Optimizing Bifrost's memory usage and performance for your specific workload.

## 📋 Overview

Bifrost provides three primary knobs for tuning performance and memory consumption:

- **Concurrency (`concurrency`)**: Controls the number of simultaneous requests to each provider.
- **Request Buffering (`buffer_size`)**: Defines the queue size for pending requests for each provider.
- **Object Pooling (`initial_pool_size`)**: Pre-allocates memory for request/response objects to reduce garbage collection overhead.

Understanding how these settings interact is key to configuring Bifrost for high throughput, low latency, or resource-constrained environments.

---

## 1. Concurrency Control (`concurrency`)

Concurrency determines how many worker goroutines are spawned for each provider to process requests in parallel.

- **What it is**: The maximum number of simultaneous requests Bifrost will make to a single provider's API.
- **Impact**: Directly controls the throughput for each provider.
- **Trade-offs**:
  - **Higher Concurrency**: Increases throughput but also increases the risk of hitting API rate limits. Consumes more memory and CPU for in-flight requests.
  - **Lower Concurrency**: Reduces the risk of rate limiting and consumes fewer resources, but may limit throughput.
- **Configuration**: This is configured on a per-provider basis.

<details>
<summary><strong>🔧 Go Package - Concurrency Configuration</strong></summary>

Concurrency is set within the `ProviderConfig` returned by your `Account` implementation.

```go
// In your Account implementation
func (a *MyAccount) GetConfigForProvider(provider schemas.ModelProvider) (*schemas.ProviderConfig, error) {
    // ...
    return &schemas.ProviderConfig{
        ConcurrencyAndBufferSize: schemas.ConcurrencyAndBufferSize{
            Concurrency: 10, // 10 concurrent workers for this provider
            BufferSize:  50,
        },
        // ...
    }, nil
}
```

</details>

<details>
<summary><strong>🌐 HTTP Transport - Concurrency Configuration</strong></summary>

Concurrency is set in your `config.json` under each provider's `concurrency_and_buffer_size`.

```json
{
  "providers": {
    "openai": {
      // ...
      "concurrency_and_buffer_size": {
        "concurrency": 10,
        "buffer_size": 50
      }
    }
  }
}
```

</details>

---

## 2. Request Queuing (`buffer_size`)

The buffer is a queue that holds incoming requests waiting to be processed by the concurrent workers.

- **What it is**: The number of requests that can be queued for a provider before new requests either block or are dropped.
- **Impact**: Helps Bifrost absorb traffic bursts without losing requests.
- **Trade-offs**:
  - **Larger Buffer**: Can handle larger bursts of traffic, preventing blocking. However, it consumes more memory to hold the queued request objects.
  - **Smaller Buffer**: Consumes less memory but may cause requests to block or be dropped during traffic spikes if workers can't keep up.
- **`dropExcessRequests`**: If the buffer is full, the behavior depends on the global `dropExcessRequests` setting (Go package only).
  - `false` (default): New requests will block until space is available in the queue.
  - `true`: New requests are immediately dropped with an error.

<details>
<summary><strong>🔧 Go Package - Buffer Configuration</strong></summary>

The buffer size is set alongside concurrency in `ProviderConfig`.

```go
// In your Account implementation
func (a *MyAccount) GetConfigForProvider(provider schemas.ModelProvider) (*schemas.ProviderConfig, error) {
    // ...
    return &schemas.ProviderConfig{
        ConcurrencyAndBufferSize: schemas.ConcurrencyAndBufferSize{
            Concurrency: 10,
            BufferSize:  50, // Queue up to 50 requests
        },
        // ...
    }, nil
}

// Global config for dropping excess requests
bifrost, err := bifrost.Init(schemas.BifrostConfig{
    //...
    DropExcessRequests: true, // Drop requests when queue is full
})
```

</details>

<details>
<summary><strong>🌐 HTTP Transport - Buffer Configuration</strong></summary>

The buffer size is set in your `config.json`. The `drop_excess_requests` setting can be configured in the `client` section and defaults to `false` (blocking).

```json
{
  "client": {
    "drop_excess_requests": false
  },
  "providers": {
    "openai": {
      // ...
      "concurrency_and_buffer_size": {
        "concurrency": 10,
        "buffer_size": 50
      }
    }
  }
}
```

</details>

---

## 3. Object Pooling (`initial_pool_size`)

Bifrost uses object pools to reuse request and response objects, reducing the load on the garbage collector and improving latency.

- **What it is**: A global setting that pre-allocates a specified number of objects for requests, responses, and errors.
- **Impact**: Significantly reduces memory allocation and GC pressure during high-traffic scenarios.
- **Trade-offs**:
  - **Larger Pool**: Improves performance under heavy load by minimizing allocations. Increases the initial memory footprint of Bifrost.
  - **Smaller Pool**: Lower initial memory usage, but may lead to more GC activity and higher latency under load.
- **Configuration**: This is a global setting. For the Go package, it is set in `BifrostConfig`. For the HTTP transport, it's configured in the `client` section of `config.json` or via the web UI.

<details open>
<summary><strong>🔧 Go Package - Object Pool Configuration</strong></summary>

Set `InitialPoolSize` in the `BifrostConfig` during initialization.

```go
// Global config for object pooling
bifrost, err := bifrost.Init(schemas.BifrostConfig{
    Account:        myAccount,
    InitialPoolSize: 1000, // Pre-allocate 1000 objects of each type
    // ...
})
```

</details>

<details>
<summary><strong>🌐 HTTP Transport - Object Pool Configuration</strong></summary>

The pool size for the HTTP transport is configured in the `config.json` file under the `client` section or via the web UI.

**Using Config File**

```json
{
  "client": {
    "initial_pool_size": 1000,
    "drop_excess_requests": false
  },
  "providers": {
    // ... provider configurations
  }
}
```

**Using Web UI**

1. Start Bifrost: `docker run -p 8080:8080 maximhq/bifrost`
2. Open `http://localhost:8080`
3. Navigate to the "Configuration" section
4. Set "Initial Pool Size" and other client settings
5. Save configuration

</details>

---

## ✨ Future Development

### Dynamic Scaling

> **Note:** This feature is under active development.

A planned feature for Bifrost is dynamic scaling, which will allow `concurrency` and `buffer_size` to adjust automatically based on real-time request load and provider feedback (like rate-limit headers). This will enable Bifrost to smartly self-tune for optimal performance and cost-efficiency.

---

## ⚙️ Configuration Recommendations

Tune these settings based on your application's traffic patterns and performance goals.

| Use Case                    | Concurrency (per provider) | Buffer Size (per provider) | Initial Pool Size (global) | Goal                                                                |
| --------------------------- | -------------------------- | -------------------------- | -------------------------- | ------------------------------------------------------------------- |
| **🚀 High-Throughput**      | 50-200                     | 500-1000                   | 1000-5000                  | Maximize RPS, assuming provider rate limits are high.               |
| **⚖️ Balanced** (Default)   | 10-50                      | 100-500                    | 500-1000                   | Good for most production workloads with moderate traffic.           |
| **💧 Burst-Resistant**      | 10-20                      | 1000-5000                  | 500-1000                   | Handles sudden traffic spikes without dropping requests.            |
| **🌱 Resource-Constrained** | 2-5                        | 10-50                      | 50-100                     | Minimizes memory footprint for development or low-traffic services. |

---

## 📊 Monitoring Memory

Monitor your Bifrost instance to ensure your configuration is optimal.

- **Prometheus Metrics**: The HTTP transport exposes metrics at the `/metrics` endpoint. While there are no specific memory metrics, you can monitor `go_memstats_*` to observe memory usage.
- **Go Profiling (pprof)**: For detailed memory analysis when using the Go package, use the standard `net/http/pprof` tool to inspect heap allocations and goroutine counts.

> **💡 Tip:** Start with the **Balanced** configuration and adjust based on observed performance and resource utilization. For example, if you see requests blocking frequently, increase `buffer_size`. If your provider rate limits are being hit, decrease `concurrency`.
