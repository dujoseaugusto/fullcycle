package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type CotacaoResponse struct {
	USD struct {
		Bid string `json:"bid"`
	} `json:"USD"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: received non-200 response status:", resp.Status)
		return
	}

	var cotacao CotacaoResponse
	if err := json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
		fmt.Println("Error decoding response:", err)
		return
	}

	cotacaoTexto := fmt.Sprintf("Dólar: %s", cotacao.USD.Bid)
	if err := ioutil.WriteFile("cotacao.txt", []byte(cotacaoTexto), 0644); err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("Cotação salva com sucesso:", cotacaoTexto)
}