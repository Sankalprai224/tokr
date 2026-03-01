package bpe

import (
	"testing"
)

// setupTestTokenizer creates a tokenizer with a complete ASCII vocabulary
func setupTestTokenizer() *tokenizer {
	t := &tokenizer{
		cache:  make(map[string][]int),
		merges: make(map[pair]int),
		// The map stores []byte, not string
		vocab:     make(map[int][]byte),
		tokenLens: make(map[int]int),
	}

	// AUTOMATIC FIX: Populate vocab with all 256 ASCII characters
	for i := 0; i < 256; i++ {
		// FIX: Use []byte{byte(i)} directly. Do NOT convert to string().
		t.vocab[i] = []byte{byte(i)}
	}

	return t
}

func TestEncodeDecode(t *testing.T) {
	tok := setupTestTokenizer()

	original := "Hello World"

	// Test Encode
	tokens, _ := tok.Encode(original, false)
	if len(tokens) == 0 {
		t.Fatalf("Expected tokens, got empty slice")
	}

	// Test Round-Trip Integrity
	// Note: Make sure to use the correct method name (Decode vs Decoder)
	decoded := tok.Decoder(tokens)
	if decoded != original {
		t.Errorf("Mismatch!\nExpected: %q\nGot:      %q\nTokens:   %v", original, decoded, tokens)
	}
}
