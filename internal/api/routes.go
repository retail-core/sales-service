package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/retail-core/sales-service/internal/config"
	"github.com/retail-core/sales-service/internal/db"
	"github.com/retail-core/sales-service/internal/handler"
	"github.com/retail-core/sales-service/internal/repository"
	"github.com/retail-core/sales-service/internal/service"
	"go.uber.org/zap"
)

func ConfigureRoutes(config config.Config, _logger *zap.Logger) http.Handler {
	
	database, err := db.InitDB(config)
	if err != nil {
		_logger.Fatal("Database initialization failed", zap.Error(err))
	}

	orderRepo := repository.NewGormOrderRepository(database)
	orderService := service.NewOrderServiceImpl(orderRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Route("/v1", func(v1 chi.Router) {
		v1.Post("/orders", orderHandler.CreateOrder)
	})

	return r
}