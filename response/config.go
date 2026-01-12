package response

import "sync"

type Config struct {
	MaxTraceSize         int
	ResponseSizeLimit    int // in bytes
	MaxInterceptorAmount int
	DefaultContentType   string
	EnableSizeValidation bool
	DefaultModule        string
	ErrorHandler         ErrorHandler // Custom error handler for FromError
}

// Default configuration values
var defaultConfig = Config{
	MaxTraceSize:         50,
	ResponseSizeLimit:    10 * 1024 * 1024, // 10MB
	MaxInterceptorAmount: 20,
	DefaultContentType:   "application/json",
	EnableSizeValidation: true,
	DefaultModule:        "GoResponse",
	ErrorHandler:         nil, // Will use defaultErrorHandler
}

// Global configuration (thread-safe)
var (
	globalConfig   Config
	globalConfigMu sync.RWMutex
)

// Initialize with default config
func init() {
	globalConfig = defaultConfig
}

// SetConfig updates the global configuration
func SetConfig(config Config) {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()

	// Validate config values
	if config.MaxTraceSize <= 0 {
		config.MaxTraceSize = defaultConfig.MaxTraceSize
	}
	if config.ResponseSizeLimit <= 0 {
		config.ResponseSizeLimit = defaultConfig.ResponseSizeLimit
	}
	if config.MaxInterceptorAmount <= 0 {
		config.MaxInterceptorAmount = defaultConfig.MaxInterceptorAmount
	}
	if config.DefaultContentType == "" {
		config.DefaultContentType = defaultConfig.DefaultContentType
	}

	globalConfig = config

	// Update the error handler if provided
	if config.ErrorHandler != nil {
		errorHandlerMu.Lock()
		errorHandler = config.ErrorHandler
		errorHandlerMu.Unlock()
	}
}

// GetConfig returns a copy of the current global configuration
func GetConfig() Config {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig
}

// getConfig is a helper to get current config (internal use)
func getConfig() Config {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig
}

// getResponseConfig returns the config for this specific response
// Falls back to global config if no specific config is set
func (r *Response) getResponseConfig() Config {
	// Check if this response has a specific config set
	// We detect this by checking if any field differs from zero value
	if r.config.MaxTraceSize > 0 || r.config.ResponseSizeLimit > 0 ||
		r.config.MaxInterceptorAmount > 0 || r.config.DefaultContentType != "" {
		return r.config
	}
	return getConfig()
}
