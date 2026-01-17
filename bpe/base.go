package bpe

import (
	"sync"
)

type pair struct {
	first  int
	second int
}

type tokenizer struct {
	merges        map[pair]int
	vocab         map[int][]byte
	specialTokens map[string]int
	orderedPairs  []pair
	mu            sync.RWMutex
	bufferpool    sync.Pool
	tokenLens     map[int]int
	pattern       string
	cache         map[string][]int
	cacheMu       sync.RWMutex
	CacheHits     int64
	TotalChunks   int64
}

func NewTokenizer() *tokenizer {
	t := &tokenizer{
		merges:        make(map[pair]int),
		vocab:         make(map[int][]byte),
		specialTokens: make(map[string]int),
		cache:         make(map[string][]int),
	}

	t.bufferpool.New = func() interface{} {
		s := make([]int, 0, 1024)
		return &s
	}
	return t

}

func getStats(tokens []int) map[pair]int {

	counts := make(map[pair]int)

	for i := 0; i < len(tokens)-1; i++ {
		p := pair{tokens[i], tokens[i+1]}
		counts[p]++
	}
	return counts
}

func merger(tokens []int, p pair, idx int) []int {

	if len(tokens) < 2 {
		return tokens
	}

	newTokens := make([]int, 0, len(tokens))
	i := 0

	for i < len(tokens) {
		if tokens[i] == p.first && i < len(tokens)-1 && tokens[i+1] == p.second {
			newTokens = append(newTokens, idx)
			i += 2
		} else {
			newTokens = append(newTokens, tokens[i])
			i += 1
		}

	}
	return newTokens
}

func (t *tokenizer) buildVocab() {
	t.vocab = make(map[int][]byte)

	for i := 0; i < 256; i++ {
		t.vocab[i] = []byte{byte(i)}
	}
	for i, pair := range t.orderedPairs {
		idx := 256 + i

		bytes1 := t.vocab[pair.first]
		bytes2 := t.vocab[pair.second]

		t.vocab[idx] = append(append([]byte(nil), bytes1...), bytes2...)
	}
	for token, idx := range t.specialTokens {
		t.vocab[idx] = []byte(token)
	}

	t.tokenLens = make(map[int]int, len(t.vocab))
	for id, b := range t.vocab {
		t.tokenLens[id] = len(b)
	}
}
