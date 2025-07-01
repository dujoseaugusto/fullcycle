package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type OrderHandler struct {
	listOrdersUseCase  *ListOrdersUseCase
	getOrderUseCase    *GetOrderUseCase
	createOrderUseCase *CreateOrderUseCase
	updateOrderUseCase *UpdateOrderUseCase
	deleteOrderUseCase *DeleteOrderUseCase
}

func NewOrderHandler(
	listOrdersUseCase *ListOrdersUseCase,
	getOrderUseCase *GetOrderUseCase,
	createOrderUseCase *CreateOrderUseCase,
	updateOrderUseCase *UpdateOrderUseCase,
	deleteOrderUseCase *DeleteOrderUseCase,
) *OrderHandler {
	return &OrderHandler{
		listOrdersUseCase:  listOrdersUseCase,
		getOrderUseCase:    getOrderUseCase,
		createOrderUseCase: createOrderUseCase,
		updateOrderUseCase: updateOrderUseCase,
		deleteOrderUseCase: deleteOrderUseCase,
	}
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.listOrdersUseCase.Execute()
	if err != nil {
		log.Printf("Error listing orders: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path using Go 1.22+ pattern matching
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	order, err := h.getOrderUseCase.Execute(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		log.Printf("Error getting order: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.createOrderUseCase.Execute(order); err != nil {
		if strings.Contains(err.Error(), "cannot be empty") || strings.Contains(err.Error(), "cannot be negative") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("Error creating order: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Order created successfully"})
}

func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path using Go 1.22+ pattern matching
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Ensure the order ID in the body matches the URL path
	order.ID = id

	if err := h.updateOrderUseCase.Execute(order); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "cannot be empty") || strings.Contains(err.Error(), "cannot be negative") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("Error updating order: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Order updated successfully"})
}

func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path using Go 1.22+ pattern matching
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	if err := h.deleteOrderUseCase.Execute(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		log.Printf("Error deleting order: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Order deleted successfully"})
}

func (h *OrderHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
