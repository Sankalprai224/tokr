package bpe

import (
	"log"
	"math"
	"strings"
	"sync/atomic"
)

func (t *tokenizer) Encode(text string, useGPT4 bool) ([]int, error) {

	if len(text) == 0 {
		return []int{}, nil
	}

	//t.mu.RLock()
	//defer t.mu.RUnlock()

	//textChunks := SplitText(text, useGPT4)

	//allTokens := make([]int, 0, len(text)/3+1)

	//for _, chunk := range textChunks {

	//	atomic.AddInt64(&t.TotalChunks, 1)

	t.cacheMu.RLock()
	cached, found := t.cache[text]
	t.cacheMu.RUnlock()

	if found {
		atomic.AddInt64(&t.CacheHits, 1)
		//allTokens = append(allTokens, cached...)
		return append([]int(nil), cached...), nil

	}

	tokens, err := t.encodeCore(text, useGPT4)
	if err != nil {
		log.Printf("error : regex failed")
		return nil, err
	}

	finalTokens := make([]int, len(tokens))
	copy(finalTokens, tokens)

	t.cacheMu.Lock()
	if len(t.cache) > 100000 {
		t.cache = make(map[string][]int)
	}
	t.cache[text] = finalTokens
	t.cacheMu.Unlock()

	atomic.AddInt64(&t.TotalChunks, 1)
	return tokens, nil

}

func (t *tokenizer) encodeCore(text string, useGPT4 bool) ([]int, error) {

	allTokens := make([]int, 0, len(text)/3+1)

	ids := make([]int, 0, 1024)

	if useGPT4 {
		match, err := re.FindStringMatch(text)
		if err != nil {
			log.Printf("regex failed")
			return nil, err
		}

		for match != nil {
			chunk := match.String()

			//	for _, chunk := range textChunks {

			ids = ids[:0]

			for i := 0; i < len(chunk); i++ {
				ids = append(ids, int(chunk[i]))
			}

			ids = runMergeLogic(t, ids)

			allTokens = append(allTokens, ids...)

			match, err = re.FindNextMatch(match)
			if err != nil {
				log.Printf("error regex failed")
				return nil, err
			}

		}
	} else {
		textChunks := FastSplit(text)

		for _, chunk := range textChunks {
			ids = ids[:0]
			for i := 0; i < len(chunk); i++ {
				ids = append(ids, int(chunk[i]))
			}
			ids = runMergeLogic(t, ids)
			allTokens = append(allTokens, ids...)
		}
	}
	return allTokens, nil

}

func runMergeLogic(t *tokenizer, val []int) []int {

	for {
		if len(val) < 2 {
			break
		}
		bestIdx := -1
		minrank := math.MaxInt

		for i := 0; i < len(val)-1; i++ {
			p := pair{val[i], val[i+1]}
			if rank, ok := t.merges[p]; ok {
				if rank < minrank {
					minrank = rank
					bestIdx = i
				}
			}
		}
		if bestIdx == -1 {
			break
		}

		val[bestIdx] = minrank

		copy(val[bestIdx+1:], val[bestIdx+2:])
		val = val[:len(val)-1]
	}

	return val

}

func (t *tokenizer) Decode(ids []int) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	totalLen := 0
	for _, id := range ids {
		if l, ok := t.tokenLens[id]; ok {
			totalLen += l
		}
	}

	var builder strings.Builder
	builder.Grow(totalLen)

	for _, id := range ids {
		if b, ok := t.vocab[id]; ok {
			builder.Write(b)
		}
	}

	return builder.String()
}
