package stresstest

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Report contém o relatório de execução dos testes
type Report struct {
	Metrics *Metrics
}

// NewReport cria um novo relatório
func NewReport(metrics *Metrics) *Report {
	return &Report{
		Metrics: metrics,
	}
}

// Print imprime o relatório na saída padrão
func (r *Report) Print() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                    STRESS TEST REPORT")
	fmt.Println(strings.Repeat("=", 80))

	// Informações gerais
	fmt.Printf("\nExecution Summary:\n")
	fmt.Printf("  Total Duration:      %v\n", r.Metrics.GetDuration())
	fmt.Printf("  Total Requests:      %d\n", r.Metrics.TotalRequests)
	fmt.Printf("  Successful:          %d\n", r.Metrics.SuccessfulRequests)
	fmt.Printf("  Failed:              %d\n", r.Metrics.FailedRequests)
	fmt.Printf("  Requests/Second:     %.2f\n", r.Metrics.GetRequestsPerSecond())

	// Distribuição de status codes
	fmt.Printf("\nHTTP Status Code Distribution:\n")
	r.printStatusCodeDistribution()

	// Tempo de execução detalhado
	fmt.Printf("\nTiming:\n")
	fmt.Printf("  Start Time:          %s\n", r.Metrics.StartTime.Format(time.RFC3339))
	fmt.Printf("  End Time:            %s\n", r.Metrics.EndTime.Format(time.RFC3339))
	fmt.Printf("  Duration:            %.2f seconds\n", r.Metrics.GetDuration().Seconds())

	fmt.Println("\n" + strings.Repeat("=", 80) + "\n")
}

// printStatusCodeDistribution imprime a distribuição de status codes
func (r *Report) printStatusCodeDistribution() {
	if len(r.Metrics.StatusCodeCount) == 0 {
		fmt.Println("  No responses received")
		return
	}

	// Ordenar status codes
	var statusCodes []int
	for code := range r.Metrics.StatusCodeCount {
		statusCodes = append(statusCodes, code)
	}
	sort.Ints(statusCodes)

	// Imprimir status codes
	for _, code := range statusCodes {
		count := r.Metrics.StatusCodeCount[code]
		percentage := float64(count) * 100 / float64(r.Metrics.TotalRequests)
		statusText := fmt.Sprintf("%s (%d)", getStatusText(code), code)
		fmt.Printf("  %-40s: %d (%.2f%%)\n", statusText, count, percentage)
	}
}

// Remove unused import warning by using strings in print
var _ = strings.Repeat

// getStatusText retorna a descrição de um status code
func getStatusText(code int) string {
	switch code {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 204:
		return "No Content"
	case 301:
		return "Moved Permanently"
	case 302:
		return "Found"
	case 304:
		return "Not Modified"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 429:
		return "Too Many Requests"
	case 500:
		return "Internal Server Error"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	case 504:
		return "Gateway Timeout"
	default:
		if code >= 400 && code < 500 {
			return "Client Error"
		} else if code >= 500 && code < 600 {
			return "Server Error"
		}
		return "Unknown"
	}
}

// String retorna uma representação em string do relatório
func (r *Report) String() string {
	output := fmt.Sprintf(`
╔════════════════════════════════════════════════════════════════════════════╗
║                        STRESS TEST REPORT                                  ║
╚════════════════════════════════════════════════════════════════════════════╝

📊 Execution Summary:
  • Total Duration:       %v
  • Total Requests:       %d
  • Successful:           %d
  • Failed:               %d
  • Requests/Second:      %.2f

📈 HTTP Status Code Distribution:
`,
		r.Metrics.GetDuration(),
		r.Metrics.TotalRequests,
		r.Metrics.SuccessfulRequests,
		r.Metrics.FailedRequests,
		r.Metrics.GetRequestsPerSecond(),
	)

	// Ordenar status codes
	var statusCodes []int
	for code := range r.Metrics.StatusCodeCount {
		statusCodes = append(statusCodes, code)
	}
	sort.Ints(statusCodes)

	// Adicionar status codes
	for _, code := range statusCodes {
		count := r.Metrics.StatusCodeCount[code]
		percentage := float64(count) * 100 / float64(r.Metrics.TotalRequests)
		statusText := getStatusText(code)
		output += fmt.Sprintf("  • %s (%d):          %d (%.2f%%)\n", statusText, code, count, percentage)
	}

	output += fmt.Sprintf(`
⏱️  Timing:
  • Start Time:           %s
  • End Time:             %s
  • Duration:             %.2f seconds

╚════════════════════════════════════════════════════════════════════════════╝
`,
		r.Metrics.StartTime.Format(time.RFC3339),
		r.Metrics.EndTime.Format(time.RFC3339),
		r.Metrics.GetDuration().Seconds(),
	)

	return output
}
