FROM golang:1.20-alpine AS build
WORKDIR /app
COPY go.mod .
RUN go env -w GOPROXY=https://proxy.golang.org
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /service-b ./

FROM alpine:3.18
COPY --from=build /service-b /service-b
EXPOSE 8081
CMD ["/service-b"]
