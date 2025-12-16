package viacep

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAddress_Success(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"localidade":"Rio de Janeiro","uf":"RJ"}`))
	}))
	defer s.Close()

	c := NewClient(s.URL, nil)
	a, err := c.GetAddress("20040002")
	if err != nil {
		t.Fatal(err)
	}
	if a == nil {
		t.Fatalf("expected address, got nil")
	}
	if a.Localidade != "Rio de Janeiro" {
		t.Fatalf("unexpected localidade: %s", a.Localidade)
	}
}

func TestGetAddress_NotFound(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"erro": true}`))
	}))
	defer s.Close()

	c := NewClient(s.URL, nil)
	a, err := c.GetAddress("00000000")
	if err != nil {
		t.Fatal(err)
	}
	if a != nil {
		t.Fatalf("expected nil address for not found")
	}
}
