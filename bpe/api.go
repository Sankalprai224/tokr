package bpe

import (
	"math"
	"strings"
	"sync/atomic"
)

func (t *tokenizer) Encode(text string, useGPT4 bool) []int {

	if len(text) == 0 {
		return []int{}
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
		return append([]int(nil), cached...)

	}

	tokens := t.encodeCore(text, useGPT4)

	//bufptr := t.bufferpool.Get().(*[]int)
	//ids := (*bufptr)[:0]

	/*
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
	*/
	finalTokens := make([]int, len(tokens))
	copy(finalTokens, tokens)

	t.cacheMu.Lock()
	if len(t.cache) > 100000 {
		t.cache = make(map[string][]int)
	}
	t.cache[text] = finalTokens
	t.cacheMu.Unlock()

	atomic.AddInt64(&t.TotalChunks, 1)
	return tokens

	//*bufptr = ids
	//t.bufferpool.Put(bufptr)
}

func (t *tokenizer) encodeCore(text string, useGPT4 bool) []int {

	//	textChunks := SplitText(text, useGPT4)

	allTokens := make([]int, 0, len(text)/3+1)

	ids := make([]int, 0, 1024)

	if useGPT4 {
		match, _ := re.FindStringMatch(text)

		for match != nil {
			chunk := match.String()

			//	for _, chunk := range textChunks {

			ids = ids[:0]

			for i := 0; i < len(chunk); i++ {
				ids = append(ids, int(chunk[i]))
			}

			runMergeLogic(t, &ids)

			allTokens = append(allTokens, ids...)

			match, _ = re.FindNextMatch(match)

		}
	} else {
		textChunks := FastSplit(text)

		for _, chunk := range textChunks {
			ids = ids[:0]
			for i := 0; i < len(chunk); i++ {
				ids = append(ids, int(chunk[i]))
			}
			runMergeLogic(t, &ids)
			allTokens = append(allTokens, ids...)
		}
	}
	return allTokens

}

func runMergeLogic(t *tokenizer, ids *[]int) {

	val := *ids

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

	*ids = val

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
