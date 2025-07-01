package main

import (
	migrate "CleanArchitecture/migrations"
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	// Initialize database connection
	db, err := sql.Open("postgres", "postgres://user:pass@db:5432/ordersdb?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Run database migrations
	migrate.RunMigrations(db)

	// Initialize repository
	orderRepo := NewOrderRepository(db)

	// Initialize use cases
	listOrdersUseCase := &ListOrdersUseCase{Repo: orderRepo}
	getOrderUseCase := &GetOrderUseCase{Repo: orderRepo}
	createOrderUseCase := &CreateOrderUseCase{Repo: orderRepo}
	updateOrderUseCase := &UpdateOrderUseCase{Repo: orderRepo}
	deleteOrderUseCase := &DeleteOrderUseCase{Repo: orderRepo}

	// Initialize handler
	orderHandler := NewOrderHandler(
		listOrdersUseCase,
		getOrderUseCase,
		createOrderUseCase,
		updateOrderUseCase,
		deleteOrderUseCase,
	)

	// Initialize Go 1.22+ ServeMux
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /orders", orderHandler.ListOrders)
	mux.HandleFunc("POST /orders", orderHandler.CreateOrder)
	mux.HandleFunc("GET /orders/{id}", orderHandler.GetOrder)
	mux.HandleFunc("PUT /orders/{id}", orderHandler.UpdateOrder)
	mux.HandleFunc("DELETE /orders/{id}", orderHandler.DeleteOrder)
	mux.HandleFunc("GET /health", orderHandler.HealthCheck)

	// Add logging middleware
	loggedMux := loggingMiddleware(mux)

	// Start server
	log.Printf("Server starting on port 8080")
	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// loggingMiddleware adds request logging
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
