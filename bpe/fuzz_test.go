package bpe

import (
	"testing"
	"unicode/utf8"
)

func FuzzEncodeDecode(f *testing.F) {
	// 1. Add "Seed Corpus" (examples to start with)
	f.Add("Hello World")
	f.Add("1234567890")
	f.Add("!@#$%^&*()")
	f.Add("   ") // whitespace

	// 2. Setup a dummy tokenizer (same as unit test)
	// In a real scenario, you might load a small test model using `init()`
	tok := setupTestTokenizer()

	// 3. The Fuzz Loop
	f.Fuzz(func(t *testing.T, orig string) {
		// Verify input is valid UTF-8 (Go strings are usually UTF-8, but fuzzing can generate garbage)
		if !utf8.ValidString(orig) {
			return
		}

		// Step A: Encode
		// We use `false` for parallelism to keep the fuzzer deterministic and simple
		tokens := tok.Encode(orig, false)

		// Step B: Decode
		decoded := tok.Decoder(tokens)

		// Step C: Verify Integrity
		if orig != decoded {
			// This will print the exact string that broke your tokenizer
			t.Fatalf("Mismatch!\nOriginal: %q\nDecoded:  %q\nTokens:   %v", orig, decoded, tokens)
		}
	})
}
