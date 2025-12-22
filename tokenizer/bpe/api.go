package bpe

import (
	"math"
	"strings"
)

func (t *tokenizer) Encode(text string, useGPT4 bool) []int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	textChunks := SplitText(text, useGPT4)

	allTokens := make([]int, 0, len(text)/3+1)

	for _, chunk := range textChunks {

		t.cacheMu.RLock()
		cached, found := t.cache[chunk]
		t.cacheMu.RUnlock()

		if found {
			allTokens = append(allTokens, cached...)
			continue
		}

		bufptr := t.bufferpool.Get().(*[]int)
		ids := (*bufptr)[:0]

		for i := 0; i < len(chunk); i++ {
			ids = append(ids, int(chunk[i]))
		}

		for {
			if len(ids) < 2 {
				break
			}
			bestIdx := -1
			minrank := math.MaxInt

			for i := 0; i < len(ids)-1; i++ {
				p := pair{ids[i], ids[i+1]}
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

			ids[bestIdx] = minrank

			copy(ids[bestIdx+1:], ids[bestIdx+2:])
			ids = ids[:len(ids)-1]
		}

		allTokens = append(allTokens, ids...)

		finalTokens := make([]int, len(ids))
		copy(finalTokens, ids)

		t.cacheMu.Lock()
		t.cache[chunk] = finalTokens
		t.cacheMu.Unlock()

		*bufptr = ids
		t.bufferpool.Put(bufptr)
	}
	return allTokens
}

func (t *tokenizer) Decoder(ids []int) string {
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
