FROM golang:1.21-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

COPY . .

RUN go build -o stress-test .

# Entrypoint para aceitar flags de CLI
ENTRYPOINT ["./stress-test"]
CMD ["--help"]
