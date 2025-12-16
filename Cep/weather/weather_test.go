package weather

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCurrentTempC_Success(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"current":{"temp_c":15.2}}`))
	}))
	defer s.Close()

	c := NewClient(s.URL, "KEY", nil)
	temp, err := c.GetCurrentTempC("Rio de Janeiro")
	if err != nil {
		t.Fatal(err)
	}
	if temp != 15.2 {
		t.Fatalf("unexpected temp: %v", temp)
	}
}
