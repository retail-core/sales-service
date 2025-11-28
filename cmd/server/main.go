package main

import (
	"net/http"
	"github.com/retail-core/sales-service/internal/api"
	"github.com/retail-core/sales-service/internal/config"
	"github.com/retail-core/sales-service/internal/logger"
	"go.uber.org/zap"
)

func main() {

	config := config.LoadConfig()

	logger.Init(config.Environment)

	_logger := logger.L()

    r := api.ConfigureRoutes(config, _logger)

	_logger.Info("Sales service starting on port", zap.String("PORT", config.Port))

	err := http.ListenAndServe(":"+config.Port, r)
	if err != nil {
		_logger.Fatal("Failed to start server", zap.Error(err))
	}
}