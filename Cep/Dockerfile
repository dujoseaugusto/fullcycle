FROM golang:1.20-alpine AS build
WORKDIR /app
COPY go.mod .
COPY . .
RUN go build -o /cep ./...

FROM alpine:3.18
COPY --from=build /cep /cep
EXPOSE 8080
ENTRYPOINT ["/cep"]
