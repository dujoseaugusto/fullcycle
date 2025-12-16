package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dujoseaugusto/fullcycle/Cep/handler"
	"github.com/dujoseaugusto/fullcycle/Cep/viacep"
	"github.com/dujoseaugusto/fullcycle/Cep/weather"
)

func main() {
	// create clients
	httpClient := &http.Client{Timeout: 5 * time.Second}

	viacepClient := viacep.NewClient("https://viacep.com.br/ws", httpClient)
	weatherKey := os.Getenv("WEATHER_API_KEY")
	weatherClient := weather.NewClient("http://api.weatherapi.com/v1", weatherKey, httpClient)

	h := handler.NewHandler(viacepClient, weatherClient)

	http.HandleFunc("/zipcode/", h.ZipcodeHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
