// Package semanticcache provides semantic caching integration for Bifrost plugin.
// This plugin caches request body hashes using xxhash and returns cached responses for identical requests.
// It supports configurable caching behavior via the VectorStore abstraction, including success-only caching and custom cache key generation.
package semanticcache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/cespare/xxhash/v2"

	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/framework"
	"github.com/maximhq/bifrost/framework/vectorstore"
)

// Config contains configuration for the semantic cache plugin.
// The VectorStore abstraction handles the underlying storage implementation and its defaults.
// Only specify values you want to override from the semantic cache defaults.
type Config struct {
	CacheKey    string `json:"cache_key"`     // Cache key for context lookup - REQUIRED
	CacheTTLKey string `json:"cache_ttl_key"` // Cache TTL key for context lookup (optional)

	// Plugin behavior settings
	TTL    time.Duration `json:"ttl,omitempty"`    // Time-to-live for cached responses (default: 5min)
	Prefix string        `json:"prefix,omitempty"` // Prefix for cache keys (optional)

	// Advanced caching behavior
	CacheByModel    *bool `json:"cache_by_model,omitempty"`    // Include model in cache key (default: true)
	CacheByProvider *bool `json:"cache_by_provider,omitempty"` // Include provider in cache key (default: true)
}

// UnmarshalJSON implements custom JSON unmarshaling for semantic cache Config.
// It supports TTL parsing from both string durations ("1m", "1hr") and numeric seconds for configurable cache behavior.
func (c *Config) UnmarshalJSON(data []byte) error {
	// Define a temporary struct to avoid infinite recursion
	type TempConfig struct {
		CacheKey        string      `json:"cache_key"`
		CacheTTLKey     string      `json:"cache_ttl_key"`
		TTL             interface{} `json:"ttl,omitempty"`
		Prefix          string      `json:"prefix,omitempty"`
		CacheByModel    *bool       `json:"cache_by_model,omitempty"`
		CacheByProvider *bool       `json:"cache_by_provider,omitempty"`
	}

	var temp TempConfig
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set simple fields
	c.CacheKey = temp.CacheKey
	c.CacheTTLKey = temp.CacheTTLKey
	c.Prefix = temp.Prefix
	c.CacheByModel = temp.CacheByModel
	c.CacheByProvider = temp.CacheByProvider

	// Handle TTL field with custom parsing for VectorStore-backed cache behavior
	if temp.TTL != nil {
		switch v := temp.TTL.(type) {
		case string:
			// Try parsing as duration string (e.g., "1m", "1hr") for semantic cache TTL
			duration, err := time.ParseDuration(v)
			if err != nil {
				return fmt.Errorf("failed to parse TTL duration string '%s': %w", v, err)
			}
			c.TTL = duration
		case int:
			// Handle integer seconds for semantic cache TTL
			c.TTL = time.Duration(v) * time.Second
		default:
			// Try converting to string and parsing as number for semantic cache TTL
			ttlStr := fmt.Sprintf("%v", v)
			if seconds, err := strconv.ParseFloat(ttlStr, 64); err == nil {
				c.TTL = time.Duration(seconds * float64(time.Second))
			} else {
				return fmt.Errorf("unsupported TTL type: %T (value: %v)", v, v)
			}
		}
	}

	return nil
}

// Plugin implements the schemas.Plugin interface for semantic caching.
// It caches responses based on xxhash of normalized requests and returns cached
// responses for identical requests. The plugin supports configurable caching behavior
// via the VectorStore abstraction, including success-only caching and custom cache key generation.
//
// Fields:
//   - store: VectorStore instance for semantic cache operations
//   - config: Plugin configuration including semantic cache and caching settings
//   - logger: Logger instance for plugin operations
type Plugin struct {
	store  vectorstore.VectorStore
	config Config
	logger schemas.Logger
}

// Plugin constants
const (
	PluginName             string        = "semantic_cache"
	PluginLoggerPrefix     string        = "[Semantic Cache]"
	CacheConnectionTimeout time.Duration = 5 * time.Second
	CacheSetTimeout        time.Duration = 30 * time.Second
)

// Dependencies is a list of dependencies that the plugin requires.
var Dependencies []framework.FrameworkDependency = []framework.FrameworkDependency{framework.FrameworkDependencyVectorStore}

// Init creates a new semantic cache plugin instance with the provided configuration.
// It uses the VectorStore abstraction for cache operations and returns a configured plugin.
//
// The VectorStore handles the underlying storage implementation and its defaults.
// The plugin only sets defaults for its own behavior (TTL, cache key generation, etc.).
//
// Parameters:
//   - config: Semantic cache and plugin configuration (CacheKey is required)
//   - logger: Logger instance for the plugin
//   - store: VectorStore instance for cache operations
//
// Returns:
//   - schemas.Plugin: A configured semantic cache plugin instance
//   - error: Any error that occurred during plugin initialization
func Init(ctx context.Context, config Config, logger schemas.Logger, store vectorstore.VectorStore) (schemas.Plugin, error) {
	if config.CacheKey == "" {
		return nil, fmt.Errorf("cache key is required")
	}

	// Set plugin-specific defaults (not Redis defaults)
	if config.TTL == 0 {
		logger.Debug(PluginLoggerPrefix + " TTL is not set, using default of 5 minutes")
		config.TTL = 5 * time.Minute
	}

	// Set cache behavior defaults
	if config.CacheByModel == nil {
		config.CacheByModel = bifrost.Ptr(true)
	}
	if config.CacheByProvider == nil {
		config.CacheByProvider = bifrost.Ptr(true)
	}

	return &Plugin{
		store:  store,
		config: config,
		logger: logger,
	}, nil
}

// generateRequestHash creates an xxhash of the request for semantic cache key generation.
// It normalizes the request by including only the relevant fields based on VectorStore configuration:
// - Provider (if CacheByProvider is true)
// - Model (if CacheByModel is true)
// - Input (chat completion or text completion)
// - Parameters (all parameters are included)
//
// Note: Fallbacks are excluded as they only affect error handling, not the actual response.
//
// Parameters:
//   - req: The Bifrost request to hash for semantic cache key generation
//   - cacheKey: The cache key prefix from context
//
// Returns:
//   - string: Hexadecimal representation of the xxhash for semantic cache storage
//   - error: Any error that occurred during request normalization or hashing
func (plugin *Plugin) generateRequestHash(req *schemas.BifrostRequest, cacheKey string) (string, error) {
	// Create a normalized request for hashing
	// Note: Fallbacks are excluded as they only affect error handling, not the actual response
	normalizedReq := struct {
		Provider schemas.ModelProvider    `json:"provider,omitempty"`
		Model    string                   `json:"model,omitempty"`
		Input    schemas.RequestInput     `json:"input"`
		Params   *schemas.ModelParameters `json:"params,omitempty"`
	}{
		Input: req.Input,
	}

	// Include provider and model based on configuration
	if plugin.config.CacheByProvider != nil && *plugin.config.CacheByProvider {
		normalizedReq.Provider = req.Provider
	}
	if plugin.config.CacheByModel != nil && *plugin.config.CacheByModel {
		normalizedReq.Model = req.Model
	}

	// Include all parameters in cache key
	normalizedReq.Params = req.Params

	// Marshal to JSON for consistent hashing
	jsonData, err := json.Marshal(normalizedReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Generate hash based on configured algorithm
	hash := xxhash.Sum64(jsonData)
	return fmt.Sprintf("%s_%x", cacheKey, hash), nil
}

// ContextKey is a custom type for context keys to prevent key collisions
type ContextKey string

const (
	requestHashKey ContextKey = "semantic_cache_request_hash"
	isCacheHitKey  ContextKey = "semantic_cache_is_cache_hit"
)

// GetName returns the canonical name of the semantic cache plugin.
// This name is used for plugin identification and logging purposes.
//
// Returns:
//   - string: The plugin name for semantic cache
func (plugin *Plugin) GetName() string {
	return PluginName
}

// PreHook is called before a request is processed by Bifrost.
// It checks if a cached response exists for the request hash and returns it if found.
//
// Parameters:
//   - ctx: Pointer to the context.Context
//   - req: The incoming Bifrost request
//
// Returns:
//   - *schemas.BifrostRequest: The original request
//   - *schemas.BifrostResponse: Cached response if found, nil otherwise
//   - error: Any error that occurred during cache lookup
func (plugin *Plugin) PreHook(ctx *context.Context, req *schemas.BifrostRequest) (*schemas.BifrostRequest, *schemas.PluginShortCircuit, error) {
	// Get the cache key from the context
	var cacheKey string
	var ok bool
	if ctx != nil {
		cacheKey, ok = (*ctx).Value(ContextKey(plugin.config.CacheKey)).(string)
		if !ok || cacheKey == "" {
			plugin.logger.Debug(PluginLoggerPrefix + " No cache key found in context key: " + plugin.config.CacheKey + ", continuing without caching")
			return req, nil, nil
		}
	} else {
		return req, nil, nil
	}

	// Generate hash for the request
	hash, err := plugin.generateRequestHash(req, cacheKey)
	if err != nil {
		// If we can't generate hash, just continue without caching
		plugin.logger.Debug(PluginLoggerPrefix + " Failed to generate request hash, continuing without caching")
		return req, nil, nil
	}

	// Store hash in context for PostHook
	*ctx = context.WithValue(*ctx, requestHashKey, hash)

	requestTypeValue := (*ctx).Value(bifrost.BifrostContextKeyRequestType)
	if requestTypeValue == nil {
		plugin.logger.Debug(PluginLoggerPrefix + " No request type found in context, continuing without caching")
		return req, nil, nil
	}
	requestType, ok := requestTypeValue.(bifrost.RequestType)
	if !ok {
		plugin.logger.Debug(PluginLoggerPrefix + " Request type is not a bifrost.RequestType, continuing without caching")
		return req, nil, nil
	}

	// Create cache key
	cacheKey = plugin.config.Prefix + hash

	if plugin.isStreamingRequest(requestType) {
		// For streaming requests, find all chunks and create a stream
		chunkPattern := cacheKey + "_chunk_*"

		// Get all chunk keys matching the pattern using SCAN
		var chunkKeys []string
		var cursor *string
		for {
			batch, c, err := plugin.store.GetAll(*ctx, chunkPattern, cursor, 1000)
			if err != nil {
				plugin.logger.Warn(PluginLoggerPrefix + " Failed to scan cached chunks, continuing with request")
				return req, nil, nil
			}
			chunkKeys = append(chunkKeys, batch...)
			cursor = c
			if cursor == nil {
				break
			}
		}

		if len(chunkKeys) == 0 {
			plugin.logger.Debug(PluginLoggerPrefix + " No cached chunks found, continuing with request")
			return req, nil, nil
		}

		plugin.logger.Info(fmt.Sprintf("%s Found %d cached chunks for request %s, returning stream", PluginLoggerPrefix, len(chunkKeys), cacheKey))

		// Create stream channel
		streamChan := make(chan *schemas.BifrostStream)

		go func() {
			defer close(streamChan)

			// Get all chunk data
			chunkData, err := plugin.store.GetChunks(*ctx, chunkKeys)
			if err != nil {
				if !errors.Is(err, vectorstore.ErrNotFound) {
					plugin.logger.Debug(PluginLoggerPrefix + " No cached chunks found, continuing with request")
					return
				}
				plugin.logger.Warn(PluginLoggerPrefix + " Failed to retrieve cached chunks")
				return
			}

			var chunks []schemas.BifrostResponse
			for _, data := range chunkData {
				if data == nil {
					continue
				}

				// Unmarshal cached response
				var cachedResponse schemas.BifrostResponse
				if err := json.Unmarshal([]byte(data.(string)), &cachedResponse); err != nil {
					plugin.logger.Warn(PluginLoggerPrefix + " Failed to unmarshal cached chunk, skipping")
					continue
				}

				chunks = append(chunks, cachedResponse)
			}

			// Sort chunks by index
			sort.Slice(chunks, func(i, j int) bool {
				return chunks[i].ExtraFields.ChunkIndex < chunks[j].ExtraFields.ChunkIndex
			})

			// Send chunks in order
			for _, chunk := range chunks {
				if chunk.ExtraFields.RawResponse == nil {
					chunk.ExtraFields.RawResponse = make(map[string]interface{})
				}
				if rawResponseMap, ok := chunk.ExtraFields.RawResponse.(map[string]interface{}); ok {
					rawResponseMap["bifrost_cached"] = true
					rawResponseMap["bifrost_cache_key"] = fmt.Sprintf("%s_chunk_%d", cacheKey, chunk.ExtraFields.ChunkIndex)
				}

				chunk.ExtraFields.Provider = req.Provider

				streamChan <- &schemas.BifrostStream{
					BifrostResponse: &chunk,
				}
			}
		}()

		*ctx = context.WithValue(*ctx, isCacheHitKey, true)

		// Return short-circuit with stream
		return req, &schemas.PluginShortCircuit{
			Stream: streamChan,
		}, nil

	} else {
		// Check if cached response exists
		cachedData, err := plugin.store.GetChunk(*ctx, cacheKey)
		if err != nil {
			if errors.Is(err, vectorstore.ErrNotFound) {
				plugin.logger.Debug(PluginLoggerPrefix + " No cached response found, continuing with request")
				// No cached response found, continue with normal processing
				return req, nil, nil
			}
			// Log error but continue processing
			plugin.logger.Warn(PluginLoggerPrefix + " Failed to get cached response, continuing without caching")
			return req, nil, nil
		}

		// Unmarshal cached response
		var cachedResponse schemas.BifrostResponse
		if err := json.Unmarshal([]byte(cachedData), &cachedResponse); err != nil {
			// If we can't unmarshal, just continue without cached response
			plugin.logger.Warn(PluginLoggerPrefix + " Failed to unmarshal cached response, continuing without caching")
			return req, nil, nil
		}

		plugin.logger.Debug(fmt.Sprintf("%s Found cached response for request %s, returning it", PluginLoggerPrefix, cacheKey))

		// Mark response as cached in extra fields
		if cachedResponse.ExtraFields.RawResponse == nil {
			cachedResponse.ExtraFields.RawResponse = make(map[string]interface{})
		}
		if rawResponseMap, ok := cachedResponse.ExtraFields.RawResponse.(map[string]interface{}); ok {
			rawResponseMap["bifrost_cached"] = true
			rawResponseMap["bifrost_cache_key"] = cacheKey
		}
		cachedResponse.ExtraFields.Provider = req.Provider

		*ctx = context.WithValue(*ctx, isCacheHitKey, true)

		// Return cached response
		return req, &schemas.PluginShortCircuit{
			Response: &cachedResponse,
		}, nil
	}

}

// PostHook is called after a response is received from a provider.
// It caches the response using the request hash as the key via the VectorStore abstraction, with optional filtering
// based on configurable caching behavior.
//
// The function performs the following operations:
// 1. Checks configurable caching behavior and skips caching for unsuccessful responses if configured
// 2. Retrieves the request hash from the context (set during PreHook)
// 3. Marshals the response for storage
// 4. Stores the response in the VectorStore-backed cache asynchronously (non-blocking)
//
// The VectorStore Add operation runs in a separate goroutine to avoid blocking the response.
// The function gracefully handles errors and continues without caching if any step fails,
// ensuring that response processing is never interrupted by caching issues.
//
// Parameters:
//   - ctx: Pointer to the context.Context containing the request hash
//   - res: The response from the provider to be cached
//   - bifrostErr: The error from the provider, if any (used for success determination)
//
// Returns:
//   - *schemas.BifrostResponse: The original response, unmodified
//   - *schemas.BifrostError: The original error, unmodified
//   - error: Any error that occurred during caching preparation (always nil as errors are handled gracefully)
func (plugin *Plugin) PostHook(ctx *context.Context, res *schemas.BifrostResponse, bifrostErr *schemas.BifrostError) (*schemas.BifrostResponse, *schemas.BifrostError, error) {
	if bifrostErr != nil {
		return res, bifrostErr, nil
	}

	isCacheHit := (*ctx).Value(isCacheHitKey)
	if isCacheHit != nil {
		isCacheHitValue, ok := isCacheHit.(bool)
		if ok && isCacheHitValue {
			// If the cache hit is true, we should not cache
			return res, nil, nil
		}
	}

	// Get the request type from context
	requestTypeValue := (*ctx).Value(bifrost.BifrostContextKeyRequestType)
	if requestTypeValue == nil {
		plugin.logger.Debug(PluginLoggerPrefix + " No request type found in context, continuing without caching")
		return res, nil, nil
	}

	requestType, ok := requestTypeValue.(bifrost.RequestType)
	if !ok {
		plugin.logger.Debug(PluginLoggerPrefix + " Request type is not a bifrost.RequestType, continuing without caching")
		return res, nil, nil
	}

	// Get the hash from context
	hashValue := (*ctx).Value(requestHashKey)
	if hashValue == nil {
		// If we don't have the hash, we can't cache (expected when cache key is not present)
		return res, nil, nil
	}

	hash, ok := hashValue.(string)
	if !ok {
		plugin.logger.Debug(PluginLoggerPrefix + " Hash is not a string, continuing without caching")
		return res, nil, nil
	}

	cacheTTL := plugin.config.TTL

	// Get the request TTL from the context
	ttlValue := (*ctx).Value(ContextKey(plugin.config.CacheTTLKey))
	if ttlValue != nil {
		ttl, ok := ttlValue.(time.Duration)
		if !ok {
			plugin.logger.Debug(PluginLoggerPrefix + " TTL is not a time.Duration, using default TTL")
		} else {
			cacheTTL = ttl
		}
	}

	// Create cache key
	cacheKey := plugin.config.Prefix + hash

	// Add "chunk_{index}" to the cache key for streaming responses
	if plugin.isStreamingRequest(requestType) {
		cacheKey = fmt.Sprintf("%s_chunk_%d", cacheKey, res.ExtraFields.ChunkIndex)
	}

	// Cache the response asynchronously to avoid blocking the response
	go func() {
		// Create a background context with timeout for the cache operation
		// This ensures the cache operation doesn't run indefinitely
		cacheCtx, cancel := context.WithTimeout(context.Background(), CacheSetTimeout)
		defer cancel()

		// Marshal response for caching
		responseData, err := json.Marshal(res)
		if err != nil {
			// If we can't marshal, just return the response without caching
			plugin.logger.Warn(PluginLoggerPrefix + " Failed to marshal response, continuing without caching")
			return
		}

		// Perform the VectorStore Add operation for semantic cache storage
		err = plugin.store.Add(cacheCtx, cacheKey, string(responseData), cacheTTL)
		if err != nil {
			plugin.logger.Warn(PluginLoggerPrefix + " Failed to cache response asynchronously: " + err.Error())
		} else {
			plugin.logger.Debug(fmt.Sprintf("%s Cached response for request %s", PluginLoggerPrefix, cacheKey))
		}
	}()

	return res, nil, nil
}

// Cleanup performs cleanup operations for the semantic cache plugin.
// It removes all cached entries with the configured prefix from the VectorStore-backed cache.
//
// The function performs the following operations:
// 1. Retrieves all cache keys matching the configured prefix pattern
// 2. Deletes all matching cache entries from the VectorStore-backed cache
//
// This method should be called when shutting down the application to ensure
// proper resource cleanup.
//
// Returns:
//   - error: Any error that occurred during cleanup operations
func (plugin *Plugin) Cleanup() error {
	// Get all keys matching the prefix using SCAN
	var keys []string
	var cursor *string
	pattern := plugin.config.Prefix + "*"
	for {
		batch, c, err := plugin.store.GetAll(context.Background(), pattern, cursor, 1000)
		if err != nil {
			return fmt.Errorf("failed to scan keys for cleanup: %w", err)
		}
		keys = append(keys, batch...)
		cursor = c
		if cursor == nil {
			break
		}
	}

	if len(keys) > 0 {
		if err := plugin.store.Delete(context.Background(), keys); err != nil {
			return fmt.Errorf("failed to delete cache keys: %w", err)
		}
		plugin.logger.Debug(fmt.Sprintf("%s Cleaned up %d cache entries", PluginLoggerPrefix, len(keys)))
	}

	return nil
}

// ClearCacheForKey deletes a specific cache key from the VectorStore-backed semantic cache.
// It is used to clear a specific cache key when needed.
//
// Parameters:
//   - key: The cache key to delete
//
// Returns:
//   - error: Any error that occurred during cache key deletion
func (plugin *Plugin) ClearCacheForKey(key string) error {
	var keys []string
	keys = append(keys, key)

	// For streaming requests, we need to delete all chunks for the key
	chunkPattern := key + "_chunk_*"

	// Get all chunk keys matching the pattern using SCAN
	var chunkKeys []string
	var cursor *string
	for {
		batch, c, err := plugin.store.GetAll(context.Background(), chunkPattern, cursor, 1000)
		if err != nil {
			plugin.logger.Warn(PluginLoggerPrefix + " Failed to scan cached chunks, continuing with request")
			return err
		}
		chunkKeys = append(chunkKeys, batch...)
		cursor = c
		if cursor == nil {
			break
		}
	}

	keys = append(keys, chunkKeys...)

	if err := plugin.store.Delete(context.Background(), keys); err != nil {
		plugin.logger.Warn(PluginLoggerPrefix + " Failed to get cached chunks, continuing with request")
		return err
	}

	return nil
}

// UTILS FUNCTIONS

func (plugin *Plugin) isStreamingRequest(requestType bifrost.RequestType) bool {
	return requestType == bifrost.ChatCompletionStreamRequest ||
		requestType == bifrost.SpeechStreamRequest ||
		requestType == bifrost.TranscriptionStreamRequest
}
