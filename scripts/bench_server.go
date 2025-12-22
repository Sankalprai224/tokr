package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := "http://localhost:8080/encode"
	concurrency := 50
	numRequests := 10000

	// A realistic 1KB payload
	payload := `{"text": "The quick brown fox jumps over the lazy dog. This is a simulation of a standard API payload. It is not too big, not too small. Just right for testing throughput."}`

	fmt.Printf("🔥 Hammering %s with %d concurrent workers (%d total reqs)...\n", url, concurrency, numRequests)

	var wg sync.WaitGroup
	var successCount int64
	var totalTokens int64
	start := time.Now()

	reqChan := make(chan int, numRequests)

	// Workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range reqChan {
				resp, err := http.Post(url, "application/json", bytes.NewBufferString(payload))
				if err == nil {
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					// Parse to count tokens
					var r struct {
						Count int `json:"count"`
					}
					json.Unmarshal(body, &r)

					if resp.StatusCode == 200 {
						atomic.AddInt64(&successCount, 1)
						atomic.AddInt64(&totalTokens, int64(r.Count))
					}
				}
			}
		}()
	}

	// Feed the workers
	for i := 0; i < numRequests; i++ {
		reqChan <- i
	}
	close(reqChan)
	wg.Wait()

	duration := time.Since(start).Seconds()
	rps := float64(successCount) / duration
	tps := float64(totalTokens) / duration

	fmt.Println("\n--------------------------------")
	fmt.Printf("✅ Requests/sec: %.2f\n", rps)
	fmt.Printf("🚀 System Tokens/sec: %.2f kTokens/s\n", tps/1000.0)
	fmt.Printf("⏱️  Avg Latency:  %.2f ms\n", (duration/float64(successCount))*1000*float64(concurrency))
	fmt.Println("--------------------------------")
}
