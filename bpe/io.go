package bpe

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func (t *tokenizer) Save(filename string) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	w := bufio.NewWriter(f)

	if _, err := fmt.Fprintln(w, "minbpe v1"); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, t.pattern); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, len(t.specialTokens)); err != nil {
		return err
	}
	for token, id := range t.specialTokens {
		if _, err := fmt.Fprintf(w, "%s %d\n", token, id); err != nil {
			return err
		}
	}

	for _, p := range t.orderedPairs {
		if _, err := fmt.Fprintf(w, "%d %d\n", p.first, p.second); err != nil {
			return err
		}
	}
	return w.Flush()
}

func Load(filename string) (*tokenizer, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	t := NewTokenizer()
	scanner := bufio.NewScanner(f)

	if !scanner.Scan() {
		return nil, fmt.Errorf("empty file")
	}

	if scanner.Text() != "minbpe v1" {
		return nil, fmt.Errorf("unknown format %s\n", scanner.Text())
	}

	if !scanner.Scan() {
		return nil, fmt.Errorf("missing pattern")
	}

	t.pattern = scanner.Text()

	if !scanner.Scan() {
		return nil, fmt.Errorf("missing special token count")
	}

	numSpecial, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return nil, fmt.Errorf("missing special token count")
	}

	for i := 0; i < numSpecial; i++ {
		if !scanner.Scan() {
			return nil, fmt.Errorf("end of file")
		}

		parts := strings.Fields(scanner.Text())
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid tokens ")
		}

		name := parts[0]
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("undable to fetch the id of specialToken")
		}
		t.specialTokens[name] = id
	}

	currentId := 256

	for scanner.Scan() {

		lines := scanner.Text()
		parts := strings.Fields(lines)

		if len(parts) != 2 {
			return nil, fmt.Errorf("not enough pairs")
		}

		p1, err1 := strconv.Atoi(parts[0])
		p2, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("error in fetching the code")
		}

		p := pair{p1, p2}
		t.merges[p] = currentId
		t.orderedPairs = append(t.orderedPairs, p)

		currentId++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	t.buildVocab()
	return t, nil
}
