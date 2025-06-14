package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"migrations/migrate"
)


func main() {

	db, err := sql.Open("postgres", "postgres://user:pass@db:5432/ordersdb?sslmode=disable")
	// TODO: Replace this with your actual migration logic or import the correct package.
	// migrate.RunMigrations(db)
		log.Fatal(err)
	}  
	migrate.RunMigrations(db)
	defer db.Close()

	// Inicialize repositório, usecase, etc.
	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// TODO: Replace 'nil' with an actual repository implementation, e.g., NewOrderRepository(db)
		orders, err := ListOrdersUseCase{Repo: nil}.Execute()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(orders)
	})
	http.ListenAndServe(":8080", nil)
}
