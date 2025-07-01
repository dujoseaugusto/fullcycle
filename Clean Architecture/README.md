# Clean Architecture Go API

A RESTful API built with Go 1.23+ following Clean Architecture principles, using the standard library's `net/http` package with Go 1.22+ ServeMux for routing.

## Features

- **Clean Architecture**: Separation of concerns with Domain, Use Cases, Repository, and Handlers
- **Go 1.22+ ServeMux**: Modern routing with pattern matching and method-based routing
- **PostgreSQL**: Database with migrations using Goose
- **Docker**: Containerized application with Docker Compose
- **RESTful API**: Full CRUD operations for Orders
- **Input Validation**: Proper validation and error handling
- **Logging**: Request logging middleware

## Project Structure

```
Clean Architecture/
├── main.go              # Application entry point
├── domain.go            # Domain entities
├── usecase.go           # Business logic use cases
├── repository.go        # Data access layer
├── handlers.go          # HTTP handlers
├── migrations/          # Database migrations
├── docker-compose.yml   # Docker services
├── dockerfile          # Application container
└── test.http           # API test requests
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/orders` | List all orders |
| POST | `/orders` | Create a new order |
| GET | `/orders/{id}` | Get order by ID |
| PUT | `/orders/{id}` | Update an order |
| DELETE | `/orders/{id}` | Delete an order |

## Running the Application

### Prerequisites
- Docker and Docker Compose
- Go 1.23+ (for local development)

### Using Docker Compose
```bash
cd "Clean Architecture"
docker-compose up --build
```

The API will be available at `http://localhost:8081`

### Local Development
```bash
cd "Clean Architecture"
go mod tidy
go run .
```

## Testing the API

Use the provided `test.http` file or curl commands:

```bash
# Health check
curl http://localhost:8081/health

# List orders
curl http://localhost:8081/orders

# Create order
curl -X POST http://localhost:8081/orders \
  -H "Content-Type: application/json" \
  -d '{"id": "order-123", "value": 99.99}'

# Get order
curl http://localhost:8081/orders/order-123

# Update order
curl -X PUT http://localhost:8081/orders/order-123 \
  -H "Content-Type: application/json" \
  -d '{"value": 149.99}'

# Delete order
curl -X DELETE http://localhost:8081/orders/order-123
```

## Architecture Layers

### Domain Layer (`domain.go`)
- Contains business entities (Order)
- Pure business logic, no external dependencies

### Use Case Layer (`usecase.go`)
- Implements business rules and orchestration
- Depends on repository interface
- Input validation and business logic

### Repository Layer (`repository.go`)
- Data access implementation
- Database operations
- Implements repository interface

### Handler Layer (`handlers.go`)
- HTTP request/response handling
- JSON serialization/deserialization
- Error handling and status codes

## Error Handling

The API returns appropriate HTTP status codes:
- `200 OK`: Successful operations
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid input data
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server errors

## Database

PostgreSQL database with automatic migrations using Goose. The database schema includes:
- `orders` table with `id` and `value` columns

```sh
docker compose up