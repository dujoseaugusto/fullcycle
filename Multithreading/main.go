package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"
)

type Endereco struct {
    Cep         string `json:"cep"`
    Logradouro  string `json:"logradouro"`
    Bairro      string `json:"bairro"`
    Localidade  string `json:"localidade"`
    Uf          string `json:"uf"`
    Origem      string
}

func buscaBrasilAPI(ctx context.Context, cep string, ch chan<- Endereco) {
    url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    var e Endereco
    if err := json.Unmarshal(body, &e); err == nil {
        e.Origem = "BrasilAPI"
        ch <- e
    }
}

func buscaViaCEP(ctx context.Context, cep string, ch chan<- Endereco) {
    url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    var e Endereco
    if err := json.Unmarshal(body, &e); err == nil {
        e.Origem = "ViaCEP"
        ch <- e
    }
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Uso: go run busca_cep.go <CEP>")
        return
    }
    cep := os.Args[1]
    ch := make(chan Endereco, 2)
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    go buscaBrasilAPI(ctx, cep, ch)
    go buscaViaCEP(ctx, cep, ch)

    select {
    case res := <-ch:
        fmt.Printf("Origem: %s\nCEP: %s\nLogradouro: %s\nBairro: %s\nCidade: %s\nUF: %s\n",
            res.Origem, res.Cep, res.Logradouro, res.Bairro, res.Localidade, res.Uf)
    case <-ctx.Done():
        fmt.Println("Timeout: nenhuma API respondeu em 1 segundo.")
    }
}