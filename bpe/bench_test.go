package bpe

import (
	"testing"
)

// Global sink variable to prevent compiler optimization
var resultSink []int

// Run with: go test -bench=. -benchmem ./bpe
func BenchmarkEncodeCore(b *testing.B) {
	// 1. Setup
	t := &tokenizer{
		cache:     make(map[string][]int),
		merges:    make(map[pair]int),
		vocab:     make(map[int][]byte),
		tokenLens: make(map[int]int),
	}

	// Add dummy merges to force the logic to run
	t.merges[pair{101, 108}] = 200 // 'e' + 'l' -> 'el'
	t.merges[pair{108, 111}] = 201 // 'l' + 'o' -> 'lo'

	text := "Hello world! This is a test of the zero-allocation system."

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Assign to global variable so compiler doesn't skip this
		resultSink, _ = t.encodeCore(text, true)
	}
}

func BenchmarkPublicAPI(b *testing.B) {
	t := &tokenizer{
		cache:  make(map[string][]int),
		merges: make(map[pair]int),
		// Note: Public API needs locks initialized if they are pointers,
		// but standard sync.Mutex works fine as zero-value.
	}
	text := "Hello world repeated many times"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resultSink, _ = t.Encode(text, true)
	}
}
