package FUN

import (
	"sync"
)

type Config struct {
	MaxTraceSize         int
	ResponseSizeLimit    int // in bytes
	MaxInterceptorAmount int
	DefaultContentType   string
	EnableSizeValidation bool
	DefaultModule        string
	IsDevelopment        bool
}

var defaultConfig = Config{
	MaxTraceSize:         50,
	ResponseSizeLimit:    10 * 1024 * 1024, // 10MB
	MaxInterceptorAmount: 20,
	DefaultContentType:   "application/json",
	EnableSizeValidation: true,
	DefaultModule:        "GoResponse",
}

var (
	globalConfig   Config
	globalConfigMu sync.RWMutex
)

func init() {
	globalConfig = defaultConfig
}

func SetConfig(config Config) {
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

	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	globalConfig = config
}

func GetConfig() Config {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig
}

func getConfig() Config {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig
}
