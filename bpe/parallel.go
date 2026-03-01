package bpe

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"runtime"
	"strings"
	"sync"
)

type Job struct {
	Index int
	Text  string
}

type Result struct {
	Index  int
	Tokens []int
}

func (t *tokenizer) ParallelEncode(text string, useGPT4 bool) []int {
	r := bufio.NewReader(strings.NewReader(text))
	chunkidx := 0
	var sb strings.Builder
	chunks := []string{}
	totalRead := 0
	const chunksize = 999000
	//const chunksize = 50000
	for {
		lines, err := r.ReadString('\n')
		totalRead += len(lines)
		sb.WriteString(lines)

		if totalRead >= chunksize {
			fmt.Printf("the size of file is %d\n", totalRead)
			chunkidx += 1
			totalRead = 0
			chunks = append(chunks, sb.String())
			sb.Reset()

		}

		if err == io.EOF {
			fmt.Printf("EOF reached, the size of file is %d\n", totalRead)
			break
		}

		if err != nil {
			fmt.Println("error reading file:", err)
			break
		}
	}
	if sb.Len() > 0 {
		chunks = append(chunks, sb.String())
		chunkidx += 1
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
	for i, text := range chunks {
		jobs <- Job{
			Index: i,
			Text:  text,
		}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	final := make([][]int, numjobs)
	for res := range results {
		final[res.Index] = res.Tokens
	}

	var alltokens []int
	for _, chunkalltokens := range final {
		alltokens = append(alltokens, chunkalltokens...)
	}
	return alltokens

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
		}
	}
}
