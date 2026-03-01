package bpe

func Train(text string, vocabSize int, useGPT4 bool) *tokenizer {

	t := NewTokenizer()
	chunks := SplitText(text, useGPT4)

	if useGPT4 {
		t.pattern = `'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+`
	} else {
		t.pattern = ""
	}

	var ids [][]int
	for _, chunk := range chunks {
		b := []byte(chunk)
		tokens := make([]int, len(b))
		for i, c := range b {
			tokens[i] = int(c)
		}
		ids = append(ids, tokens)
	}
	numMerges := vocabSize - 256
	for i := 0; i < numMerges; i++ {

		stats := make(map[pair]int)
		for _, chunk := range ids {
			chunkStats := getStats(chunk)
			for p, count := range chunkStats {
				stats[p] += count
			}
		}
		if len(stats) == 0 {
			break
		}
		var bestpair pair
		bestcount := 0

		for p, count := range stats {
			if count > bestcount || (count == bestcount && (p.first < bestpair.first || (p.first == bestpair.first && p.second < bestpair.second))) {
				bestcount = count
				bestpair = p
			}

		}
		newId := 256 + i
		t.merges[bestpair] = newId

		t.orderedPairs = append(t.orderedPairs, bestpair)

		for j, _ := range ids {
			ids[j] = merger(ids[j], bestpair, newId)
		}
	}
	t.buildVocab()
	return t
}
