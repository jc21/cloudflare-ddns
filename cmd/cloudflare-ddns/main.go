package main

import (
	"github.com/jc21/cloudflare-ddns/internal/config"
	"github.com/jc21/cloudflare-ddns/internal/logger"
	"github.com/jc21/cloudflare-ddns/internal/updater"
)

func main() {
	argConfig := config.GetConfig()
	log := logger.Init(argConfig)
	log.Trace("Args: %+v", argConfig)

	if argConfig.Setup {
		config.SetupConfig()
	}

	cfg := config.GetCloudflareConfig()
	log.Trace("Config: %+v", cfg)
	updater.Process(argConfig, cfg)
}
