package bpe

import (
	"unicode"

	"github.com/dlclark/regexp2"
)

func SplitText(text string, useGPT4 bool) []string {
	if useGPT4 {
		return GPTSplit(text)
	}
	return FastSplit(text)
}

func GPTSplit(text string) []string {
	var tokens []string

	pattern := `'s|'t|'re|'ve|'m|'ll|'d| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+`
	re := regexp2.MustCompile(pattern, regexp2.Compiled)

	match, _ := re.FindStringMatch(text)
	for match != nil {
		if token := match.String(); token != "" {
			tokens = append(tokens, token)
		}
		match, _ = re.FindNextMatch(match)
	}
	return tokens
}

func FastSplit(text string) []string {
	if text == "" {
		return nil
	}

	runes := []rune(text)
	n := len(runes)
	var tokens []string
	i := 0

	for i < n {

		start := i
		r := runes[i]

		if unicode.IsSpace(r) {
			if i+1 < n {
				if unicode.IsLetter(runes[i+1]) {
					i++

					for i < n && (unicode.IsLetter(runes[i])) {
						i++
					}
					tokens = append(tokens, string(runes[start:i]))
					continue
				} else if unicode.IsNumber(runes[i+1]) {
					i++
					for i < n && (unicode.IsNumber(runes[i])) {
						i++
					}
					tokens = append(tokens, string(runes[start:i]))
					continue
				}
			}
			for i < n && unicode.IsSpace(runes[i]) {
				i++
			}
			tokens = append(tokens, string(runes[start:i]))
			continue
		}
		if unicode.IsLetter(r) {

			for i < n && (unicode.IsLetter(runes[i])) {
				i++
			}
			tokens = append(tokens, string(runes[start:i]))
			continue
		}
		if unicode.IsNumber(r) {

			for i < n && (unicode.IsNumber(runes[i])) {
				i++
			}
			tokens = append(tokens, string(runes[start:i]))
			continue
		}
		for i < n && !(unicode.IsLetter(runes[i]) || unicode.IsNumber(runes[i]) || unicode.IsSpace(runes[i])) {
			i++
		}
		tokens = append(tokens, string(runes[start:i]))
	}
	return tokens
}
