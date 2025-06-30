# ❌ Error Handling

Understanding Bifrost's structured error format and best practices for error handling.

## 📋 Overview

**Error Handling Features:**

- ✅ **Structured Errors** - Consistent error format across all providers
- ✅ **Error Codes** - Specific error codes for different failure types
- ✅ **Context Information** - Detailed error context and debugging info
- ✅ **Automatic Fallbacks** - Bifrost handles provider fallbacks automatically
- ✅ **Circuit Breaking** - Available via plugins for advanced reliability
- ✅ **Provider Mapping** - Provider-specific errors mapped to common format

**Benefits:**

- 🔍 **Easier Debugging** - Structured error information with context
- 📊 **Better Monitoring** - Categorized errors for alerting and metrics
- 🛡️ **Built-in Reliability** - Automatic fallbacks and recovery
- ⚡ **Simple Integration** - Handle errors without complex retry logic

---

## 🏗️ Error Structure

### BifrostError Schema

<details open>
<summary><strong>🔧 Go Package - BifrostError Structure</strong></summary>

```go
// Bifrost error structure
type BifrostError struct {
    EventID        *string    `json:"event_id,omitempty"`        // Unique error event ID
    Type           *string    `json:"type,omitempty"`            // High-level error category
    IsBifrostError bool       `json:"is_bifrost_error"`          // Always true for Bifrost errors
    StatusCode     *int       `json:"status_code,omitempty"`     // HTTP status code equivalent
    Error          ErrorField `json:"error"`                     // Detailed error information
}

type ErrorField struct {
    Type    *string     `json:"type,omitempty"`    // Specific error type
    Code    *string     `json:"code,omitempty"`    // Error code
    Message string      `json:"message"`           // Human-readable error message
    Error   error       `json:"error,omitempty"`   // Original error (Go only)
    Param   interface{} `json:"param,omitempty"`   // Parameter that caused the error
    EventID *string     `json:"event_id,omitempty"` // Error event ID
}

// Check if error is a BifrostError
func isBifrostError(err error) (*schemas.BifrostError, bool) {
    var bifrostErr *schemas.BifrostError
    if errors.As(err, &bifrostErr) {
        return bifrostErr, true
    }
    return nil, false
}
```

</details>

<details>
<summary><strong>🌐 HTTP Transport - Error Response Format</strong></summary>

```json
{
  "error": {
    "type": "rate_limit_error",
    "code": "rate_limit_exceeded",
    "message": "Rate limit exceeded for model gpt-4o. Please retry after 60 seconds.",
    "param": "model"
  },
  "is_bifrost_error": true,
  "status_code": 429,
  "event_id": "evt_abc123def456"
}
```

**HTTP Status Codes:**

| Error Type              | HTTP Status | Description                    |
| ----------------------- | ----------- | ------------------------------ |
| `authentication_error`  | 401         | Invalid or missing credentials |
| `authorization_error`   | 403         | Insufficient permissions       |
| `rate_limit_error`      | 429         | Rate limit exceeded            |
| `invalid_request_error` | 400         | Malformed request              |
| `api_error`             | 500         | Internal server error          |
| `network_error`         | 502/503     | Network or connectivity issues |

</details>

---

## 🎯 Basic Error Handling

### Simple Error Handling

<details open>
<summary><strong>🔧 Go Package - Basic Error Handling</strong></summary>

```go
func handleBasicErrors(bf *bifrost.Bifrost, request schemas.BifrostRequest) (*schemas.BifrostResponse, error) {
    response, err := bf.ChatCompletion(context.Background(), request)
    if err != nil {
        // Check if it's a structured Bifrost error
        if bifrostErr, ok := isBifrostError(err); ok {
            // Log structured error with context
            logStructuredError(bifrostErr, request)

            // Handle specific error types
            switch bifrostErr.Error.Type {
            case "authentication_error":
                return nil, fmt.Errorf("authentication failed: %s", bifrostErr.Error.Message)
            case "rate_limit_error":
                return nil, fmt.Errorf("rate limited: %s", bifrostErr.Error.Message)
            case "network_error":
                return nil, fmt.Errorf("network error: %s", bifrostErr.Error.Message)
            default:
                return nil, fmt.Errorf("bifrost error: %s", bifrostErr.Error.Message)
            }
        }

        // Handle non-Bifrost errors
        log.WithFields(log.Fields{
            "provider": request.Provider,
            "model":    request.Model,
        }).Error("Unexpected error:", err)

        return nil, err
    }

    return response, nil
}

func logStructuredError(bifrostErr *schemas.BifrostError, request schemas.BifrostRequest) {
    fields := log.Fields{
        "provider":   request.Provider,
        "model":      request.Model,
        "error_type": bifrostErr.Error.Type,
        "error_code": bifrostErr.Error.Code,
    }

    if bifrostErr.EventID != nil {
        fields["event_id"] = *bifrostErr.EventID
    }

    log.WithFields(fields).Error(bifrostErr.Error.Message)
}
```

> **💡 Note:** Bifrost automatically handles fallbacks between providers, so you don't need to implement manual fallback logic.

</details>

<details>
<summary><strong>🌐 HTTP Transport - Basic Error Handling</strong></summary>

```python
import requests
import logging
from typing import Dict, Any, Optional

class BifrostClient:
    def __init__(self, base_url: str):
        self.base_url = base_url
        self.logger = logging.getLogger(__name__)

    def chat_completion(self, payload: Dict[Any, Any]) -> Optional[Dict[Any, Any]]:
        try:
            response = requests.post(
                f"{self.base_url}/v1/chat/completions",
                json=payload,
                timeout=30
            )

            if response.status_code == 200:
                return response.json()

            # Handle Bifrost errors
            error_data = response.json()
            if error_data.get("is_bifrost_error"):
                self.log_structured_error(error_data, payload)

                error_type = error_data.get("error", {}).get("type")
                error_message = error_data.get("error", {}).get("message", "Unknown error")

                if error_type == "authentication_error":
                    raise Exception(f"Authentication failed: {error_message}")
                elif error_type == "rate_limit_error":
                    raise Exception(f"Rate limited: {error_message}")
                elif error_type == "network_error":
                    raise Exception(f"Network error: {error_message}")
                else:
                    raise Exception(f"Bifrost error: {error_message}")

            # Handle other HTTP errors
            response.raise_for_status()

        except requests.exceptions.RequestException as e:
            self.logger.error(f"Request failed: {e}")
            raise

    def log_structured_error(self, error_data: Dict[Any, Any], payload: Dict[Any, Any]):
        error_info = error_data.get("error", {})

        self.logger.error(
            "Bifrost error occurred",
            extra={
                "provider": payload.get("provider"),
                "model": payload.get("model"),
                "error_type": error_info.get("type"),
                "error_code": error_info.get("code"),
                "error_message": error_info.get("message"),
                "event_id": error_data.get("event_id")
            }
        )

# Usage
client = BifrostClient("http://localhost:8080")

try:
    response = client.chat_completion({
        "provider": "openai",
        "model": "gpt-4o-mini",
        "messages": [{"role": "user", "content": "Hello!"}]
    })
    print("Success:", response)
except Exception as e:
    print("Error:", e)
```

> **💡 Note:** Bifrost HTTP transport automatically handles retries and fallbacks, so simple error handling is usually sufficient.

</details>

---

## 📋 Common Error Types

### Authentication Errors

| Code                  | Description                     | Status Code | Action                   |
| --------------------- | ------------------------------- | ----------- | ------------------------ |
| `invalid_api_key`     | API key is invalid or malformed | 401         | Check API key format     |
| `api_key_expired`     | API key has expired             | 401         | Rotate API key           |
| `insufficient_quota`  | Account quota exceeded          | 429         | Upgrade plan or wait     |
| `account_deactivated` | Provider account is deactivated | 403         | Contact provider support |
| `unauthorized_model`  | Model access not authorized     | 403         | Check model permissions  |

### Rate Limit Errors

| Code                           | Description                  | Status Code | Action                         |
| ------------------------------ | ---------------------------- | ----------- | ------------------------------ |
| `rate_limit_exceeded`          | General rate limit exceeded  | 429         | Wait and retry with backoff    |
| `concurrent_requests_exceeded` | Too many concurrent requests | 429         | Reduce concurrency             |
| `tokens_per_minute_exceeded`   | Token rate limit exceeded    | 429         | Split requests or wait         |
| `requests_per_day_exceeded`    | Daily request limit exceeded | 429         | Wait until next day or upgrade |

### Network Errors

| Code                    | Description                  | Status Code | Action                         |
| ----------------------- | ---------------------------- | ----------- | ------------------------------ |
| `connection_timeout`    | Request timed out            | 504         | Retry with exponential backoff |
| `connection_refused`    | Connection refused by server | 502         | Check service availability     |
| `dns_resolution_failed` | DNS lookup failed            | 502         | Check network configuration    |
| `proxy_error`           | Proxy connection failed      | 502         | Check proxy settings           |

---

## 📊 Error Monitoring

### Metrics and Alerting

<details>
<summary><strong>📈 Error Tracking</strong></summary>

**Go Package - Error Metrics:**

```go
func trackErrorMetrics(bifrostErr *schemas.BifrostError, provider schemas.ModelProvider) {
    if bifrostErr.Error.Type != nil && bifrostErr.Error.Code != nil {
        // Track error counts by type and provider
        errorCounter.WithLabelValues(
            string(provider),
            *bifrostErr.Error.Type,
            *bifrostErr.Error.Code,
        ).Inc()

        // Track error rates for alerting
        if *bifrostErr.Error.Type == "authentication_error" {
            authErrorRate.WithLabelValues(string(provider)).Inc()
        }
    }
}
```

**HTTP Transport - Prometheus Metrics:**

Bifrost automatically exposes error metrics at `/metrics`:

```bash
# Check error metrics
curl http://localhost:8080/metrics | grep -E "error"

# Example metrics:
# bifrost_errors_total{provider="openai",type="rate_limit_error",code="rate_limit_exceeded"} 5
# bifrost_error_rate{provider="openai"} 0.02
```

</details>

---

## 🛠️ Best Practices

### Error Handling Guidelines

<details>
<summary><strong>📋 Best Practices</strong></summary>

**1. Always Check for Bifrost Errors:**

```go
response, err := bf.ChatCompletion(ctx, request)
if err != nil {
    if bifrostErr, ok := isBifrostError(err); ok {
        // Handle structured Bifrost error
        handleStructuredError(bifrostErr)
    } else {
        // Handle other errors
        handleGenericError(err)
    }
}
```

**2. Log Errors with Context:**

```go
func logError(err error, request schemas.BifrostRequest) {
    if bifrostErr, ok := isBifrostError(err); ok {
        log.WithFields(log.Fields{
            "error_type": bifrostErr.Error.Type,
            "error_code": bifrostErr.Error.Code,
            "provider":   request.Provider,
            "model":      request.Model,
            "event_id":   bifrostErr.EventID,
        }).Error(bifrostErr.Error.Message)
    } else {
        log.WithFields(log.Fields{
            "provider": request.Provider,
            "model":    request.Model,
        }).Error(err.Error())
    }
}
```

**3. Monitor Error Patterns:**

```go
// Set up alerts for high error rates
if errorRate > 0.1 { // 10% error rate
    alertManager.Send("High error rate detected")
}

// Track specific error types
authErrors := getErrorCount("authentication_error")
if authErrors > 5 {
    alertManager.Send("Multiple authentication failures")
}
```

**4. Don't Implement Manual Fallbacks:**

```go
// ❌ Don't do this - Bifrost handles fallbacks automatically
providers := []schemas.ModelProvider{schemas.OpenAI, schemas.Anthropic}
for _, provider := range providers {
    // Manual fallback logic
}

// ✅ Do this - Let Bifrost handle it
response, err := bf.ChatCompletion(ctx, request)
if err != nil {
    // Just handle the final error
    logError(err, request)
    return nil, err
}
```

</details>

---

## 🎯 Next Steps

| **Task**                    | **Documentation**                         |
| --------------------------- | ----------------------------------------- |
| **🔗 Configure providers**  | [Providers](providers.md)                 |
| **🔑 Manage API keys**      | [Key Management](key-management.md)       |
| **🌐 Set up networking**    | [Networking](networking.md)               |
| **⚡ Optimize performance** | [Memory Management](memory-management.md) |

> **💡 Tip:** Bifrost handles complex error recovery automatically. Focus on understanding error types for monitoring and debugging rather than implementing retry logic.
