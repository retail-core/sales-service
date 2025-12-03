package main

import (
	"net/http"

	"github.com/retail-core/sales-service/internal/api"
	"github.com/retail-core/sales-service/internal/config"
	"github.com/retail-core/sales-service/internal/db"
	"github.com/retail-core/sales-service/internal/logger"
	"github.com/retail-core/sales-service/internal/mq"
	"go.uber.org/zap"
)

func main() {

	config := config.LoadConfig()

	logger.Init(config.Environment)

	_logger := logger.L()

	database, err := db.InitDB(config)
	if err != nil {
		_logger.Fatal("Database initialization failed", zap.Error(err))
	}

	publisher, err := mq.InitRabbitMQ(config.RABBITMQ_URL)
	if err != nil {
		_logger.Fatal("Failed to initialize RabbitMQ connection", zap.Error(err))
	}

	err = mq.DeclareExchange(publisher, mq.InventoryExchange)
	if err != nil {
		_logger.Fatal("Failed to declare RabbitMQ exchange", zap.Error(err))
	}

	defer publisher.Close()

	r := api.ConfigureRoutes(database, publisher, config.INVENTORY_SERVICE_URL)

	go func() {
		err = http.ListenAndServe(":"+config.Port, r)
		if err != nil {
			_logger.Fatal("Failed to start server", zap.Error(err))
		}

		_logger.Info("sales service running on port", zap.String("PORT", config.Port))
	}()

	select {}
}
