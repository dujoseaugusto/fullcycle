package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/dujoseaugusto/fullcycle/Cep/viacep"
	"github.com/dujoseaugusto/fullcycle/Cep/weather"
)

var cepRe = regexp.MustCompile(`^[0-9]{8}$`)

type Handler struct {
	Viacep  *viacep.Client
	Weather *weather.Client
}

func NewHandler(v *viacep.Client, w *weather.Client) *Handler {
	return &Handler{Viacep: v, Weather: w}
}

func (h *Handler) ZipcodeHandler(w http.ResponseWriter, r *http.Request) {
	// path is /zipcode/{cep}
	cep := r.URL.Path[len("/zipcode/"):]

	if !cepRe.MatchString(cep) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("invalid zipcode"))
		return
	}

	addr, err := h.Viacep.GetAddress(cep)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if addr == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("can not find zipcode"))
		return
	}

	query := fmt.Sprintf("%s,%s", addr.Localidade, addr.Uf)
	tempC, err := h.Weather.GetCurrentTempC(query)
	if err != nil {
		http.Error(w, "failed to get weather", http.StatusInternalServerError)
		return
	}

	tempF := tempC*1.8 + 32
	tempK := tempC + 273

	resp := map[string]float64{
		"temp_C": tempC,
		"temp_F": tempF,
		"temp_K": tempK,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// helper to convert string to float used in tests
func parseF(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
