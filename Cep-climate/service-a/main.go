package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var svcB string
var tracer trace.Tracer
var cepRe = regexp.MustCompile(`^\d{8}$`)

func initTracer(ctx context.Context) func(context.Context) error {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4318"
	}
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(endpoint), otlptracehttp.WithInsecure())
	if err != nil {
		log.Printf("failed to create otlp exporter: %v", err)
		return func(context.Context) error { return nil }
	}
	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(bsp))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	tracer = otel.Tracer("service-a")
	return tp.Shutdown
}

func main() {
	ctx := context.Background()
	shutdown := initTracer(ctx)
	defer shutdown(ctx)

	svcB = os.Getenv("SERVICE_B_URL")
	if svcB == "" {
		svcB = "http://localhost:8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/zipcode", zipcodeHandler)

	addr := ":8080"
	log.Printf("service-a listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

type ZipcodeRequest struct {
	Cep string `json:"cep"`
}

func zipcodeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "service-a.handle-zipcode")
	defer span.End()
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var req ZipcodeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	// validate: must be string of 8 digits
	if !cepRe.MatchString(req.Cep) {
		http.Error(w, "invalid zipcode", 422)
		return
	}
	// forward to service B with propagated trace context
	client := http.Client{Transport: http.DefaultTransport}
	reqOut, err := http.NewRequestWithContext(ctx, "POST", svcB+"/zipcode", bytes.NewReader(body))
	if err != nil {
		http.Error(w, "service error", http.StatusInternalServerError)
		return
	}
	reqOut.Header.Set("Content-Type", "application/json")
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(reqOut.Header))
	resp, err := client.Do(reqOut)
	if err != nil {
		http.Error(w, "service error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	// relay response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
