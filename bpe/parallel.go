package bpe

import (
	"log"
	"runtime"
	"sync"
)

type Job struct {
	Index int
	Text  string
}

type Result struct {
	Index  int
	Tokens []int
	Err    error
}

func (t *tokenizer) ParallelEncode(text string, useGPT4 bool) ([]int, error) {
	if len(text) == 0 {
		return nil, nil
	}

	var chunks []string
	const chunksize = 1000000 // 1MB chunks

	start := 0
	for start < len(text) {
		end := start + chunksize
		if end >= len(text) {
			chunks = append(chunks, text[start:])
			break
		}

		// Backward scan for natural boundary
		splitPoint := end
		for splitPoint > start {
			if text[splitPoint] == ' ' || text[splitPoint] == '\n' {
				break
			}
			splitPoint--
		}

		// Fallback for missing boundaries, ensuring UTF-8 integrity
		if splitPoint == start {
			splitPoint = end
			for splitPoint > start && (text[splitPoint]&0xC0) == 0x80 {
				splitPoint--
			}
		}

		chunks = append(chunks, text[start:splitPoint])
		start = splitPoint
	}

	numjobs := len(chunks)
	jobs := make(chan Job, numjobs)
	results := make(chan Result, numjobs)
	numworkers := runtime.NumCPU()

	var wg sync.WaitGroup
	for w := 0; w < numworkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(t, useGPT4, jobs, results)
		}()
	}

	for i, chunkText := range chunks {
		jobs <- Job{
			Index: i,
			Text:  chunkText,
		}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	final := make([][]int, numjobs)
	var firstError error
	for res := range results {
		if res.Err != nil && firstError == nil {
			firstError = res.Err
		}
		final[res.Index] = res.Tokens
	}

	if firstError != nil {
		return nil, firstError
	}

	// Zero-allocation assembly path
	totalLen := 0
	for _, chunkTokens := range final {
		totalLen += len(chunkTokens)
	}

	alltokens := make([]int, 0, totalLen)
	for _, chunkTokens := range final {
		alltokens = append(alltokens, chunkTokens...)
	}

	return alltokens, nil
}

func worker(t *tokenizer, useGPT4 bool, jobs <-chan Job, results chan<- Result) {
	for job := range jobs {
		tkns, err := t.encodeCore(job.Text, useGPT4)
		if err != nil {
			log.Printf("error the encoding failed due to regex faliure")
		}

		results <- Result{
			Index:  job.Index,
			Tokens: tkns,
			Err:    err,
		}
	}
}
