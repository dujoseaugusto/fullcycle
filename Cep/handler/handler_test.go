package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dujoseaugusto/fullcycle/Cep/viacep"
	"github.com/dujoseaugusto/fullcycle/Cep/weather"
)

func TestZipcodeHandler_Success(t *testing.T) {
	// mock viacep
	vs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"localidade":"São Paulo","uf":"SP"}`))
	}))
	defer vs.Close()

	// mock weather
	ws := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"current":{"temp_c":28.5}}`))
	}))
	defer ws.Close()

	vclient := viacep.NewClient(vs.URL, nil)
	wclient := weather.NewClient(ws.URL, "KEY", nil)

	h := NewHandler(vclient, wclient)

	req := httptest.NewRequest("GET", "/zipcode/01001000", nil)
	rr := httptest.NewRecorder()
	h.ZipcodeHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]float64
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if resp["temp_C"] != 28.5 {
		t.Fatalf("unexpected temp_C: %v", resp["temp_C"])
	}
}

func TestZipcodeHandler_Invalid(t *testing.T) {
	vclient := viacep.NewClient("http://example", nil)
	wclient := weather.NewClient("http://example", "KEY", nil)
	h := NewHandler(vclient, wclient)

	req := httptest.NewRequest("GET", "/zipcode/123", nil)
	rr := httptest.NewRecorder()
	h.ZipcodeHandler(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 got %d", rr.Code)
	}
	if rr.Body.String() != "invalid zipcode" {
		t.Fatalf("expected body 'invalid zipcode' got %q", rr.Body.String())
	}
}

func TestZipcodeHandler_NotFound(t *testing.T) {
	vs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"erro": true}`))
	}))
	defer vs.Close()

	vclient := viacep.NewClient(vs.URL, nil)
	wclient := weather.NewClient("http://example", "KEY", nil)
	h := NewHandler(vclient, wclient)

	req := httptest.NewRequest("GET", "/zipcode/00000000", nil)
	rr := httptest.NewRecorder()
	h.ZipcodeHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", rr.Code)
	}
	if rr.Body.String() != "can not find zipcode" {
		t.Fatalf("expected body 'can not find zipcode' got %q", rr.Body.String())
	}
}
