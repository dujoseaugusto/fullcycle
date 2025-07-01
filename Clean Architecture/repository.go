package main

import (
	"database/sql"
	"fmt"
	"log"
)

type OrderRepositoryImpl struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &OrderRepositoryImpl{db: db}
}

func (r *OrderRepositoryImpl) List() ([]Order, error) {
	query := "SELECT id, value FROM orders ORDER BY id"
	rows, err := r.db.Query(query)
	if err != nil {
		log.Printf("Error querying orders: %v", err)
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.ID, &order.Value); err != nil {
			log.Printf("Error scanning order: %v", err)
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating orders: %v", err)
		return nil, fmt.Errorf("failed to iterate orders: %w", err)
	}

	return orders, nil
}

func (r *OrderRepositoryImpl) GetByID(id string) (*Order, error) {
	query := "SELECT id, value FROM orders WHERE id = $1"
	var order Order
	err := r.db.QueryRow(query, id).Scan(&order.ID, &order.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found: %s", id)
		}
		log.Printf("Error querying order by ID: %v", err)
		return nil, fmt.Errorf("failed to query order: %w", err)
	}
	return &order, nil
}

func (r *OrderRepositoryImpl) Create(order Order) error {
	query := "INSERT INTO orders (id, value) VALUES ($1, $2)"
	_, err := r.db.Exec(query, order.ID, order.Value)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

func (r *OrderRepositoryImpl) Update(order Order) error {
	query := "UPDATE orders SET value = $2 WHERE id = $1"
	result, err := r.db.Exec(query, order.ID, order.Value)
	if err != nil {
		log.Printf("Error updating order: %v", err)
		return fmt.Errorf("failed to update order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found: %s", order.ID)
	}

	return nil
}

func (r *OrderRepositoryImpl) Delete(id string) error {
	query := "DELETE FROM orders WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting order: %v", err)
		return fmt.Errorf("failed to delete order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found: %s", id)
	}

	return nil
}
