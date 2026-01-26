package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/HeLiX-x/tokr/bpe"
)

func main() {
	modelPath := "vocab.model" // Adjust path if needed
	benchFile := "scripts/real_world.txt"

	// 1. Load Model
	fmt.Printf("📦 Loading model from %s...\n", modelPath)
	t, err := bpe.Load(modelPath)
	if err != nil {
		log.Fatalf("Failed to load model: %v", err)
	}

	// 2. Load Real Data
	fmt.Printf("📂 Reading benchmark data from %s...\n", benchFile)
	data, err := os.ReadFile(benchFile)
	if err != nil {
		log.Fatalf("Failed to read file (did you run generate_data.py?): %v", err)
	}
	text := string(data)
	fileSizeMB := float64(len(text)) / (1024 * 1024)
	fmt.Printf("📊 File Size: %.2f MB\n", fileSizeMB)

	// 3. Warmup (Optional: Single run to populate some cache)
	fmt.Println("🔥 Warming up (1 pass)...")
	t.Encode(text[:min(len(text), 5000)], true)

	// 4. The Real Test
	fmt.Println("🚀 Starting Benchmark...")
	start := time.Now()

	// We run it multiple times to simulate high load,
	// but on DIVERSE data, the cache hit rate won't be artificially 100%
	// unless the file itself is small.
	ids := t.Encode(text, true)

	duration := time.Since(start)

	// 5. Report Honest Metrics
	tokenCount := len(ids)
	seconds := duration.Seconds()

	// Access the atomic counters we added in Phase 1
	// Note: Since they are atomic, we just read them directly here
	// or use atomic.LoadInt64 if you want to be pedantic,
	// but direct read is fine after execution stops.
	hits := t.CacheHits
	total := t.TotalChunks
	hitRate := 0.0
	if total > 0 {
		hitRate = (float64(hits) / float64(total)) * 100.0
	}

	fmt.Println("\n--------------------------------")
	fmt.Printf("⏱️  Time Taken:      %.4f s\n", seconds)
	fmt.Printf("🔢 Tokens Generated: %d\n", tokenCount)
	fmt.Printf("🚀 Speed:            %.2f kTokens/sec\n", float64(tokenCount)/1000.0/seconds)
	fmt.Printf("💾 Throughput:       %.2f MB/sec\n", fileSizeMB/seconds)
	fmt.Println("--------------------------------")
	fmt.Printf("🧠 Cache Efficiency: %.2f%%\n", hitRate)
	fmt.Printf("   - Total Chunks:   %d\n", total)
	fmt.Printf("   - Cache Hits:     %d\n", hits)
	fmt.Println("--------------------------------")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
