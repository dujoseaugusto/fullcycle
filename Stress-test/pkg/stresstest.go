package stresstest

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics coleta métricas da execução dos testes
type Metrics struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	StatusCodeCount    map[int]int64
	StartTime          time.Time
	EndTime            time.Time
	mu                 sync.RWMutex
}

// TestConfig contém configurações para os testes
type TestConfig struct {
	URL           string
	TotalRequests int
	Concurrency   int
	Timeout       time.Duration
}

// StressTest executa os testes de carga
type StressTest struct {
	config  *TestConfig
	metrics *Metrics
	client  *http.Client
}

// NewStressTest cria uma nova instância de StressTest
func NewStressTest(config *TestConfig) *StressTest {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &StressTest{
		config: config,
		metrics: &Metrics{
			StatusCodeCount: make(map[int]int64),
			StartTime:       time.Now(),
		},
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Run executa os testes de carga
func (st *StressTest) Run(ctx context.Context) *Metrics {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, st.config.Concurrency)

	requestsPerWorker := st.config.TotalRequests / st.config.Concurrency
	remainingRequests := st.config.TotalRequests % st.config.Concurrency

	for i := 0; i < st.config.Concurrency; i++ {
		wg.Add(1)

		// Distribuir requests adicionais entre os workers
		requests := requestsPerWorker
		if i < remainingRequests {
			requests++
		}

		go func(numRequests int) {
			defer wg.Done()

			for j := 0; j < numRequests; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					semaphore <- struct{}{}
					st.executeRequest()
					<-semaphore
				}
			}
		}(requests)
	}

	wg.Wait()
	st.metrics.EndTime = time.Now()
	st.metrics.SuccessfulRequests = int64(st.config.TotalRequests) - st.metrics.FailedRequests

	return st.metrics
}

// executeRequest executa uma requisição HTTP
func (st *StressTest) executeRequest() {
	atomic.AddInt64(&st.metrics.TotalRequests, 1)

	ctx, cancel := context.WithTimeout(context.Background(), st.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", st.config.URL, nil)
	if err != nil {
		atomic.AddInt64(&st.metrics.FailedRequests, 1)
		return
	}

	resp, err := st.client.Do(req)
	if err != nil {
		atomic.AddInt64(&st.metrics.FailedRequests, 1)
		return
	}
	defer resp.Body.Close()

	st.metrics.mu.Lock()
	st.metrics.StatusCodeCount[resp.StatusCode]++
	st.metrics.mu.Unlock()
}

// GetDuration retorna a duração total dos testes
func (m *Metrics) GetDuration() time.Duration {
	return m.EndTime.Sub(m.StartTime)
}

// GetRequestsPerSecond retorna o número de requests por segundo
func (m *Metrics) GetRequestsPerSecond() float64 {
	duration := m.GetDuration().Seconds()
	if duration == 0 {
		return 0
	}
	return float64(m.TotalRequests) / duration
}
