// Package http provides an HTTP service using FastHTTP that exposes endpoints
// for text and chat completions using various AI model providers (OpenAI, Anthropic, Bedrock, Mistral, Ollama, etc.).
//
// The HTTP service provides the following main endpoints:
//   - /v1/text/completions: For text completion requests
//   - /v1/chat/completions: For chat completion requests
//   - /v1/mcp/tool/execute: For MCP tool execution requests
//   - /providers/*: For provider configuration management
//
// Configuration is handled through a JSON config file, high-performance ConfigStore, and environment variables:
//   - Use -config flag to specify the config file location
//   - Use -port flag to specify the server port (default: 8080)
//   - Use -pool-size flag to specify the initial connection pool size (default: 300)
//
// ConfigStore Features:
//   - Pure in-memory storage for ultra-fast config access
//   - Environment variable processing for secure configuration management
//   - Real-time configuration updates via HTTP API
//   - Explicit persistence control via POST /config/save endpoint
//   - Provider-specific meta config support (Azure, Bedrock, Vertex)
//   - Thread-safe operations with concurrent request handling
//   - Statistics and monitoring endpoints for operational insights
//
// Performance Optimizations:
//   - Configuration data is processed once during startup and stored in memory
//   - Ultra-fast memory access eliminates I/O overhead on every request
//   - All environment variable processing done upfront during configuration loading
//   - Thread-safe concurrent access with read-write mutex protection
//
// Example usage:
//
//	go run main.go -config config.example.json -port 8080 -pool-size 300
//	after setting the environment variables present in config.example.json in the environment.
//
// Integration Support:
// Bifrost supports multiple AI provider integrations through dedicated HTTP endpoints.
// Each integration exposes API-compatible endpoints that accept the provider's native request format,
// automatically convert it to Bifrost's unified format, process it, and return the expected response format.
//
// Integration endpoints follow the pattern: /{provider}/{provider_api_path}
// Examples:
//   - OpenAI: POST /openai/v1/chat/completions (accepts OpenAI ChatCompletion requests)
//   - GenAI:  POST /genai/v1beta/models/{model} (accepts Google GenAI requests)
//   - Anthropic: POST /anthropic/v1/messages (accepts Anthropic Messages requests)
//
// This allows clients to use their existing integration code without modification while benefiting
// from Bifrost's unified model routing, fallbacks, monitoring capabilities, and high-performance configuration management.
//
// NOTE: Streaming is not supported yet so all the flags related to streaming are ignored. (in both bifrost and its integrations)
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fasthttp/router"
	bifrost "github.com/maximhq/bifrost/core"
	schemas "github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/plugins/maxim"
	"github.com/maximhq/bifrost/transports/bifrost-http/handlers"
	"github.com/maximhq/bifrost/transports/bifrost-http/lib"
	"github.com/maximhq/bifrost/transports/bifrost-http/plugins/logging"
	"github.com/maximhq/bifrost/transports/bifrost-http/plugins/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// Command line flags
var (
	initialPoolSize int      // Initial size of the connection pool
	port            string   // Port to run the server on
	configPath      string   // Path to the config file
	pluginsToLoad   []string // Path to the plugins
)

// init initializes command line flags and validates required configuration.
// It sets up the following flags:
//   - pool-size: Initial connection pool size (default: 300)
//   - port: Server port (default: 8080)
//   - config: Path to config file (required)
func init() {
	pluginString := ""

	flag.IntVar(&initialPoolSize, "pool-size", 300, "Initial pool size for Bifrost")
	flag.StringVar(&port, "port", "8080", "Port to run the server on")
	flag.StringVar(&configPath, "config", "", "Path to the config file")
	flag.StringVar(&pluginString, "plugins", "", "Comma separated list of plugins to load")
	flag.Parse()

	pluginsToLoad = strings.Split(pluginString, ",")

	if configPath == "" {
		log.Fatalf("config path is required")
	}
}

// registerCollectorSafely attempts to register a Prometheus collector,
// handling the case where it may already be registered.
// It logs any errors that occur during registration, except for AlreadyRegisteredError.
func registerCollectorSafely(collector prometheus.Collector) {
	if err := prometheus.Register(collector); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			log.Printf("Failed to register collector: %v", err)
		}
	}
}

// main is the entry point of the application.
// It:
// 1. Initializes Prometheus collectors for monitoring
// 2. Reads and parses configuration from the specified config file
// 3. Initializes the Bifrost client with the configuration
// 4. Sets up HTTP routes for text and chat completions
// 5. Starts the HTTP server on the specified port
//
// The server exposes the following endpoints:
//   - POST /v1/text/completions: For text completion requests
//   - POST /v1/chat/completions: For chat completion requests
//   - GET /metrics: For Prometheus metrics
func main() {
	// Register Prometheus collectors
	registerCollectorSafely(collectors.NewGoCollector())
	registerCollectorSafely(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	logger := bifrost.NewDefaultLogger(schemas.LogLevelInfo)

	// Define a struct to unmarshal the entire config file
	var config struct {
		Client struct {
			DropExcessRequests bool     `json:"drop_excess_requests"`
			PrometheusLabels   []string `json:"prometheus_labels"`
		} `json:"client"`
		Providers json.RawMessage    `json:"providers"`
		MCP       *schemas.MCPConfig `json:"mcp"`
	}

	// Read and parse config
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("failed to parse config JSON: %v", err)
	}

	telemetry.InitPrometheusMetrics(config.Client.PrometheusLabels)
	log.Println("Prometheus Go/Process collectors registered.")

	// Initialize high-performance configuration store with caching
	store, err := lib.NewConfigStore(logger)
	if err != nil {
		log.Fatalf("failed to initialize config store: %v", err)
	}

	// Load configuration from JSON file into the store with full preprocessing
	// This processes environment variables and stores all configurations in memory for ultra-fast access
	if err := store.LoadFromConfig(configPath); err != nil {
		log.Fatalf("failed to load config into store: %v", err)
	}

	// Create account backed by the high-performance store (all processing is done in LoadFromConfig)
	// The account interface now benefits from ultra-fast config access times via in-memory storage
	account := lib.NewBaseAccount(store)

	// Get the processed MCP configuration from the store
	// All environment variable processing is already done during LoadFromConfig
	mcpConfig := store.GetMCPConfig()

	loadedPlugins := []schemas.Plugin{}

	for _, plugin := range pluginsToLoad {
		switch strings.ToLower(plugin) {
		case "maxim":
			if os.Getenv("MAXIM_LOG_REPO_ID") == "" {
				log.Println("warning: maxim log repo id is required to initialize maxim plugin")
				continue
			}
			if os.Getenv("MAXIM_API_KEY") == "" {
				log.Println("warning: maxim api key is required in environment variable MAXIM_API_KEY to initialize maxim plugin")
				continue
			}

			maximPlugin, err := maxim.NewMaximLoggerPlugin(os.Getenv("MAXIM_API_KEY"), os.Getenv("MAXIM_LOG_REPO_ID"))
			if err != nil {
				log.Printf("warning: failed to initialize maxim plugin: %v", err)
				continue
			}

			loadedPlugins = append(loadedPlugins, maximPlugin)
		}
	}

	promPlugin := telemetry.NewPrometheusPlugin()
	loggingPlugin, err := logging.NewLoggerPlugin(nil)
	if err != nil {
		log.Fatalf("failed to initialize logging plugin: %v", err)
	}

	loadedPlugins = append(loadedPlugins, promPlugin, loggingPlugin)

	client, err := bifrost.Init(schemas.BifrostConfig{
		Account:            account,
		InitialPoolSize:    initialPoolSize,
		DropExcessRequests: config.Client.DropExcessRequests,
		Plugins:            loadedPlugins,
		MCPConfig:          mcpConfig,
		Logger:             logger,
	})
	if err != nil {
		log.Fatalf("failed to initialize bifrost: %v", err)
	}

	store.SetBifrostClient(client)

	// Initialize handlers
	providerHandler := handlers.NewProviderHandler(store, client, logger)
	completionHandler := handlers.NewCompletionHandler(client, logger)
	mcpHandler := handlers.NewMCPHandler(client, logger, store)
	integrationHandler := handlers.NewIntegrationHandler(client)
	configHandler := handlers.NewConfigHandler(client, logger, configPath)
	loggingHandler := handlers.NewLoggingHandler(loggingPlugin.GetPluginLogManager(), logger)
	wsHandler := handlers.NewWebSocketHandler(loggingPlugin.GetPluginLogManager(), logger)

	// Set up WebSocket callback for real-time log updates
	loggingPlugin.SetLogCallback(wsHandler.BroadcastLogUpdate)

	// Start WebSocket heartbeat
	wsHandler.StartHeartbeat()

	r := router.New()

	// Register all handler routes
	providerHandler.RegisterRoutes(r)
	completionHandler.RegisterRoutes(r)
	mcpHandler.RegisterRoutes(r)
	integrationHandler.RegisterRoutes(r)
	configHandler.RegisterRoutes(r)
	loggingHandler.RegisterRoutes(r)
	wsHandler.RegisterRoutes(r)

	// Add Prometheus /metrics endpoint
	r.GET("/metrics", fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()))

	r.NotFound = func(ctx *fasthttp.RequestCtx) {
		handlers.SendError(ctx, fasthttp.StatusNotFound, "Route not found: "+string(ctx.Path()), logger)
	}

	server := &fasthttp.Server{
		// A custom handler that excludes middleware from /metrics
		Handler: func(ctx *fasthttp.RequestCtx) {
			if string(ctx.Path()) == "/metrics" {
				r.Handler(ctx)
				return
			}
			telemetry.PrometheusMiddleware(r.Handler)(ctx)
		},
	}

	log.Println("Started Bifrost HTTP server on port", port)
	if err := server.ListenAndServe(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	wsHandler.Stop()
	client.Cleanup()
}
