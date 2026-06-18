package bpe

import (
	"os"
	"strings"
	"testing"
)

// Global sink to prevent compiler optimization
var resultSink []int

// ── helpers ──────────────────────────────────────────────────────────────────

func loadRealTokenizer(b *testing.B) *tokenizer {
	b.Helper()
	t, err := Load("../vocab.model")
	if err != nil {
		b.Fatalf("failed to load vocab.model: %v", err)
	}
	return t
}

func loadCorpus(b *testing.B, path string) string {
	b.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		b.Skipf("corpus file not found (%s), skipping: %v", path, err)
	}
	return string(data)
}

// ── cold path: encodeCore with dummy merges (micro benchmark) ────────────────

func BenchmarkEncodeCore_Micro(b *testing.B) {
	t := &tokenizer{
		cache:     make(map[string][]int),
		merges:    make(map[pair]int),
		vocab:     make(map[int][]byte),
		tokenLens: make(map[int]int),
	}
	t.merges[pair{101, 108}] = 200 // 'e'+'l'
	t.merges[pair{108, 111}] = 201 // 'l'+'o'

	text := "Hello world! This is a test of the BPE tokenizer engine."

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		resultSink, _ = t.encodeCore(text, true)
	}
}

// ── hot path: public Encode with cache (micro benchmark) ─────────────────────

func BenchmarkEncode_Cached(b *testing.B) {
	t := &tokenizer{
		cache:  make(map[string][]int),
		merges: make(map[pair]int),
	}
	text := "Hello world repeated many times to saturate the cache"

	// warm cache
	t.Encode(text, true)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		resultSink, _ = t.Encode(text, true)
	}
}

// ── real model: single-threaded Encode on realistic text sizes ───────────────

func BenchmarkEncode_RealModel_1KB(b *testing.B) {
	tok := loadRealTokenizer(b)
	corpus := loadCorpus(b, "../scripts/real_world.txt")
	text := corpus[:min(1024, len(corpus))]

	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		resultSink, _ = tok.Encode(text, true)
	}
}

func BenchmarkEncode_RealModel_100KB(b *testing.B) {
	tok := loadRealTokenizer(b)
	corpus := loadCorpus(b, "../scripts/real_world.txt")
	text := corpus[:min(100*1024, len(corpus))]

	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		resultSink, _ = tok.Encode(text, true)
	}
}

func BenchmarkEncode_RealModel_500KB(b *testing.B) {
	tok := loadRealTokenizer(b)
	corpus := loadCorpus(b, "../scripts/real_world.txt")
	text := corpus[:min(500*1024, len(corpus))]

	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		resultSink, _ = tok.Encode(text, true)
	}
}

// ── real model: ParallelEncode on 10MB corpus ────────────────────────────────

func BenchmarkParallelEncode_10MB(b *testing.B) {
	tok := loadRealTokenizer(b)
	text := loadCorpus(b, "../10MB.txt")

	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var err error
		resultSink, err = tok.ParallelEncode(text, true)
		if err != nil {
			b.Fatalf("ParallelEncode failed: %v", err)
		}
	}
}

// ── splitter benchmarks ───────────────────────────────────────────────────────

func BenchmarkGPTSplit_100KB(b *testing.B) {
	corpus := loadCorpus(b, "../scripts/real_world.txt")
	text := corpus[:min(100*1024, len(corpus))]

	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GPTSplit(text)
	}
}

func BenchmarkFastSplit_100KB(b *testing.B) {
	corpus := loadCorpus(b, "../scripts/real_world.txt")
	text := corpus[:min(100*1024, len(corpus))]

	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FastSplit(text)
	}
}

// ── decode benchmark ──────────────────────────────────────────────────────────

func BenchmarkDecode_RealModel(b *testing.B) {
	tok := loadRealTokenizer(b)
	corpus := loadCorpus(b, "../scripts/real_world.txt")
	text := corpus[:min(100*1024, len(corpus))]

	tokens, err := tok.Encode(text, true)
	if err != nil {
		b.Fatalf("encode failed: %v", err)
	}

	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tok.Decode(tokens)
	}
}

// ── cache efficiency benchmark ────────────────────────────────────────────────

func BenchmarkEncode_CacheHitRate(b *testing.B) {
	tok := loadRealTokenizer(b)

	// simulate repeated short inputs (high cache hit scenario, like a server)
	sentences := []string{
		"The quick brown fox jumps over the lazy dog",
		"Hello world this is a tokenizer benchmark",
		"BPE encoding is fast in pure Go",
		"GPT-4 compatible tokenization without Python",
		"High throughput tokenizer for AI infrastructure",
	}

	// warm cache
	for _, s := range sentences {
		tok.Encode(s, true)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		resultSink, _ = tok.Encode(sentences[i%len(sentences)], true)
	}

	b.ReportMetric(float64(tok.CacheHits), "cache_hits")
	b.ReportMetric(float64(tok.TotalChunks), "total_chunks")
}

// ── round-trip integrity benchmark ───────────────────────────────────────────

func BenchmarkRoundTrip_RealModel(b *testing.B) {
	tok := loadRealTokenizer(b)
	corpus := loadCorpus(b, "../scripts/real_world.txt")
	text := corpus[:min(50*1024, len(corpus))]

	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokens, _ := tok.Encode(text, true)
		_ = tok.Decode(tokens)
	}
}

// ── parallel vs single-thread comparison ─────────────────────────────────────

func BenchmarkEncode_Single_vs_Parallel(b *testing.B) {
	tok := loadRealTokenizer(b)
	// Multiply by 160 to generate a ~16MB payload
	base := loadCorpus(b, "../scripts/real_world.txt")
	text := strings.Repeat(base[:min(len(base), 100*1024)], 160)

	b.Run("SingleThreaded_Raw", func(b *testing.B) {
		b.SetBytes(int64(len(text)))
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			resultSink, _ = tok.encodeCore(text, true)
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		b.SetBytes(int64(len(text)))
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			resultSink, _ = tok.ParallelEncode(text, true)
		}
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
