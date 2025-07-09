package handlers

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/fasthttp/router"
	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/transports/bifrost-http/lib"
	"github.com/valyala/fasthttp"
)

// ConfigHandler manages runtime configuration updates for Bifrost.
// It provides an endpoint to hot-reload settings from the configuration file.
type ConfigHandler struct {
	client     *bifrost.Bifrost
	logger     schemas.Logger
	store      *lib.ConfigStore
	configPath string
}

// NewConfigHandler creates a new handler for configuration management.
// It requires the Bifrost client, a logger, and the path to the config file to be reloaded.
func NewConfigHandler(client *bifrost.Bifrost, logger schemas.Logger, store *lib.ConfigStore, configPath string) *ConfigHandler {
	return &ConfigHandler{
		client:     client,
		logger:     logger,
		store:      store,
		configPath: configPath,
	}
}

// RegisterRoutes registers the configuration-related routes.
// It adds the `PUT /api/config` endpoint.
func (h *ConfigHandler) RegisterRoutes(r *router.Router) {
	r.GET("/api/config", h.GetConfig)
	r.PUT("/api/config", h.handleUpdateConfig)
	r.POST("/api/config/save", h.SaveConfig)
}

// GetConfig handles GET /config - Get the current configuration
func (h *ConfigHandler) GetConfig(ctx *fasthttp.RequestCtx) {
	config := h.store.ClientConfig
	SendJSON(ctx, config, h.logger)
}

// handleUpdateConfig updates the core configuration settings.
// Currently, it supports hot-reloading of the `drop_excess_requests` setting.
// Note that settings like `prometheus_labels` cannot be changed at runtime.
func (h *ConfigHandler) handleUpdateConfig(ctx *fasthttp.RequestCtx) {
	var req lib.ClientConfig

	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		SendError(ctx, fasthttp.StatusBadRequest, fmt.Sprintf("Invalid request format: %v", err), h.logger)
		return
	}

	// Get current config with proper locking
	currentConfig := h.store.ClientConfig
	updatedConfig := currentConfig

	if req.DropExcessRequests != currentConfig.DropExcessRequests {
		h.client.UpdateDropExcessRequests(req.DropExcessRequests)
		updatedConfig.DropExcessRequests = req.DropExcessRequests
	}

	if !slices.Equal(req.PrometheusLabels, currentConfig.PrometheusLabels) {
		updatedConfig.PrometheusLabels = req.PrometheusLabels
	}

	if req.InitialPoolSize != currentConfig.InitialPoolSize {
		updatedConfig.InitialPoolSize = req.InitialPoolSize
	}

	// Update the store with the new config
	h.store.ClientConfig = updatedConfig

	ctx.SetStatusCode(fasthttp.StatusOK)
	SendJSON(ctx, map[string]any{
		"status":  "success",
		"message": "Configuration updated successfully",
	}, h.logger)
}

// SaveConfig handles POST /config/save - Persist current configuration to JSON file
func (h *ConfigHandler) SaveConfig(ctx *fasthttp.RequestCtx) {
	// Save current configuration back to the original JSON file
	if err := h.store.SaveConfig(); err != nil {
		h.logger.Warn(fmt.Sprintf("Failed to save configuration: %v", err))
		SendError(ctx, fasthttp.StatusInternalServerError, fmt.Sprintf("Failed to save configuration: %v", err), h.logger)
		return
	}

	SendJSON(ctx, map[string]any{
		"status":  "success",
		"message": "Configuration saved successfully",
	}, h.logger)
}
