package main

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "log"
    "math"
    "net/http"
    "os"
    "regexp"
    "strconv"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/propagation"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

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
    tracer = otel.Tracer("service-b")
    return tp.Shutdown
}

func main() {
    ctx := context.Background()
    shutdown := initTracer(ctx)
    defer shutdown(ctx)

    mux := http.NewServeMux()
    mux.HandleFunc("/zipcode", zipcodeHandler)

    addr := ":8081"
    log.Printf("service-b listening on %s", addr)
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}

type ZipcodeRequest struct {
    Cep string `json:"cep"`
}

type ZipcodeResponse struct {
    City  string  `json:"city"`
    TempC float64 `json:"temp_C"`
    TempF float64 `json:"temp_F"`
    TempK float64 `json:"temp_K"`
}

var cepRe = regexp.MustCompile(`^\d{8}$`)

func zipcodeHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    var req ZipcodeRequest
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    if err := json.Unmarshal(body, &req); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    if !cepRe.MatchString(req.Cep) {
        http.Error(w, "invalid zipcode", 422)
        return
    }

    ctx, span := tracer.Start(ctx, "lookup-zipcode")
    city, err := lookupCityByCEP(ctx, req.Cep)
    span.End()
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            http.Error(w, "can not find zipcode", http.StatusNotFound)
            return
        }
        http.Error(w, "can not find zipcode", http.StatusNotFound)
        return
    }

    ctx, span2 := tracer.Start(ctx, "lookup-temperature")
    tempC, err := lookupTemperatureByCity(ctx, city)
    span2.End()
    if err != nil {
        http.Error(w, "can not find zipcode", http.StatusNotFound)
        return
    }

    resp := ZipcodeResponse{
        City:  city,
        TempC: round(tempC, 1),
        TempF: round(cToF(tempC), 1),
        TempK: round(cToK(tempC), 1),
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode(resp)
}

func lookupCityByCEP(ctx context.Context, cep string) (string, error) {
    url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    client := http.Client{Transport: http.DefaultTransport}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return "", ErrNotFound
    }
    var data struct {
        Localidade string `json:"localidade"`
        Erro       bool   `json:"erro"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return "", err
    }
    if data.Erro || data.Localidade == "" {
        return "", ErrNotFound
    }
    return data.Localidade, nil
}

func lookupTemperatureByCity(ctx context.Context, city string) (float64, error) {
    // 1) geocode city to lat/lon using Open-Meteo geocoding
    geourl := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=pt", city)
    req, _ := http.NewRequestWithContext(ctx, "GET", geourl, nil)
    client := http.Client{Transport: http.DefaultTransport}
    resp, err := client.Do(req)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return 0, ErrNotFound
    }
    var geo struct {
        Results []struct {
            Latitude  float64 `json:"latitude"`
            Longitude float64 `json:"longitude"`
        } `json:"results"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&geo); err != nil {
        return 0, err
    }
    if len(geo.Results) == 0 {
        return 0, ErrNotFound
    }
    lat := geo.Results[0].Latitude
    lon := geo.Results[0].Longitude

    // 2) call Open-Meteo to get current weather
    weatherURL := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&current_weather=true&temperature_unit=celsius", floatToStr(lat), floatToStr(lon))
    req2, _ := http.NewRequestWithContext(ctx, "GET", weatherURL, nil)
    resp2, err := client.Do(req2)
    if err != nil {
        return 0, err
    }
    defer resp2.Body.Close()
    if resp2.StatusCode != 200 {
        return 0, ErrNotFound
    }
    var wdata struct {
        CurrentWeather struct {
            Temperature float64 `json:"temperature"`
        } `json:"current_weather"`
    }
    if err := json.NewDecoder(resp2.Body).Decode(&wdata); err != nil {
        return 0, err
    }
    return wdata.CurrentWeather.Temperature, nil
}

var ErrNotFound = errors.New("not found")

func cToF(c float64) float64 { return c*1.8 + 32 }
func cToK(c float64) float64 { return c + 273 }
func round(x float64, prec int) float64 {
    pow := math.Pow(10, float64(prec))
    return math.Round(x*pow) / pow
}

func floatToStr(f float64) string {
    return strconv.FormatFloat(f, 'f', 6, 64)
}
