FROM golang:1.20-alpine AS build
WORKDIR /app

# cache deps
COPY go.mod ./
RUN go mod download

# copy sources
COPY . .

# build only the main package into a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /cep .

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /cep /cep
EXPOSE 8080
ENTRYPOINT ["/cep"]
