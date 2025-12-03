package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	stresstest "github.com/dujoseaugusto/stress-test/pkg"
)

func main() {
	// Definir flags CLI
	url := flag.String("url", "", "URL do serviço a ser testado (obrigatório)")
	requests := flag.Int("requests", 100, "Número total de requests a serem realizados")
	concurrency := flag.Int("concurrency", 10, "Número de chamadas simultâneas")

	flag.Parse()

	// Validar argumentos obrigatórios
	if *url == "" {
		fmt.Fprintf(os.Stderr, "erro: --url é obrigatório\n")
		fmt.Fprintf(os.Stderr, "uso: %s --url=<url> [--requests=<num>] [--concurrency=<num>]\n", os.Args[0])
		os.Exit(1)
	}

	if *requests <= 0 {
		fmt.Fprintf(os.Stderr, "erro: --requests deve ser um número positivo\n")
		os.Exit(1)
	}

	if *concurrency <= 0 {
		fmt.Fprintf(os.Stderr, "erro: --concurrency deve ser um número positivo\n")
		os.Exit(1)
	}

	// Criar configuração de teste
	config := &stresstest.TestConfig{
		URL:           *url,
		TotalRequests: *requests,
		Concurrency:   *concurrency,
	}

	// Exibir informações de inicio
	log.Printf("🚀 Iniciando Stress Test")
	log.Printf("   URL: %s", config.URL)
	log.Printf("   Total de Requests: %d", config.TotalRequests)
	log.Printf("   Concorrência: %d", config.Concurrency)
	log.Println()

	// Criar e executar stress test
	st := stresstest.NewStressTest(config)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metrics := st.Run(ctx)

	// Gerar e exibir relatório
	report := stresstest.NewReport(metrics)
	report.Print()

	// Se houver erro ou status não-200, retornar código de saída não-zero
	if metrics.FailedRequests > 0 {
		os.Exit(1)
	}
}
