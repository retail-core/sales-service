package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/retail-core/sales-service/internal/client"
	"github.com/retail-core/sales-service/internal/handler"
	"github.com/retail-core/sales-service/internal/mq"
	"github.com/retail-core/sales-service/internal/redis_client"
	"github.com/retail-core/sales-service/internal/repository"
	"github.com/retail-core/sales-service/internal/service"
	"gorm.io/gorm"
)

func ConfigureRoutes(database *gorm.DB, redisClient *redis.Client, mqPublisher *mq.RabbitMQConnection, invServiceUrl string) http.Handler {

	orderRepo := repository.NewGormOrderRepository(database)

	inventoryClient := client.NewHttpInventoryClient(invServiceUrl)

	queueStore := redis_client.NewRedisQueueStore(redisClient)

	orderService := service.NewOrderServiceImpl(orderRepo, inventoryClient, mqPublisher, queueStore)
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
		v1.Post("/stores/{store_id}/orders", orderHandler.CreateOrder)
		v1.Get("/stores/{store_id}/orders/{order_id}", orderHandler.GetOrder)
		v1.Get("/stores/{store_id}/orders", orderHandler.GetOrdersByStoreID)
		v1.Get("/stores/{store_id}/orders/dashboard", orderHandler.GetDashboardByStoreID)
		v1.Get("/stores/{store_id}/sales/report", orderHandler.GetSalesReport)
	})

	return r
}