package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/HeLiX-x/tokr/bpe"
)

func main() {
	inputFile := flag.String("input", "", "this is the input path for training/testing")
	modelFile := flag.String("model", "vocab.model", "this is the path to the model file")
	mode := flag.String("mode", "train", "Mode : train, test or inference")
	vocabSize := flag.Int("vocab", 10000, "target vocab size ( default : 10000)")
	useGPT4 := flag.Bool("gpt4", true, "use gpt4 regex pattern for splitting ( recommended)")
	port := flag.String("port", "8080", "the port of the server connection")

	flag.Parse()

	if *inputFile == "" && *mode == "train" {
		log.Fatal("invalind action : empty input")
	}

	switch *mode {
	case "train":
		runTrain(*inputFile, *modelFile, *vocabSize, *useGPT4)

	case "inference":
		runInference(*modelFile, *inputFile, *useGPT4)

	case "server":
		runServer(*modelFile, *port, *useGPT4)

	default:
		fmt.Println("invalid mode use one from ( train, test, inference,server)")
	}
}

func runTrain(inputPath, modelPath string, size int, gpt4 bool) {
	fmt.Printf("reading.... %s\n", inputPath)

	data, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("failed to read file %v", err)
	}

	text := string(data)

	fmt.Println("training started")
	start := time.Now()

	t := bpe.Train(text, size, gpt4)

	duration := time.Since(start)
	fmt.Printf("training completed %s\n", duration)

	fmt.Println("saving the file...")
	if err := t.Save(modelPath); err != nil {
		log.Fatalf("failed to save %v", err)
	}
}

func runInference(modelPath, inputPath string, gpt4 bool) {
	fmt.Println("Loading the file...")
	t, err := bpe.Load(modelPath)
	if err != nil {
		log.Fatalf("Failed to load the file: %v\n", err)
	}

	// 1. Determine what text to encode
	var text string
	if inputPath != "" {
		fmt.Printf("Reading input file: %s ...\n", inputPath)
		data, err := os.ReadFile(inputPath)
		if err != nil {
			log.Fatalf("Failed to read input file: %v", err)
		}
		text = string(data)
	} else {
		// Default test if no file is provided
		text = "Hello, world! This is a test of the tokenizer."
	}

	fmt.Printf("Encoding %d bytes of text...\n", len(text))

	// 2. Start the Timer (The Benchmark)
	start := time.Now()

	// 3. Run the Hot Path
	ids := t.ParallelEncode(text, gpt4)

	// 4. Stop the Timer
	duration := time.Since(start)

	// 5. Report Results
	fmt.Printf("--------------------------------\n")
	fmt.Printf("Time:       %s\n", duration)
	fmt.Printf("Token Count: %d\n", len(ids))

	seconds := duration.Seconds()
	if seconds > 0 {
		kTokensPerSec := float64(len(ids)) / 1000.0 / seconds
		mbPerSec := (float64(len(text)) / (1024 * 1024)) / seconds
		fmt.Printf("Speed:      %.2f kTokens/sec\n", kTokensPerSec)
		fmt.Printf("Throughput: %.2f MB/sec\n", mbPerSec)
	}
	fmt.Printf("--------------------------------\n")

	// Only print decoded text if it's short (to avoid flooding your terminal)
	if len(text) < 1000 {
		decoded := t.Decode(ids)
		fmt.Printf("Decoded: %q\n", decoded)
		if text == decoded {
			fmt.Println("Round-trip match successful")
		} else {
			fmt.Println("Failed to match")
		}
	}
}

type TokenizeRequest struct {
	Text string `json:"text"`
}

type TokenizeResponse struct {
	Tokens []int   `json:"tokens"`
	Count  int     `json:"count"`
	Time   float64 `json:"time_seconds"`
}

type DecodeRequest struct {
	Tokens []int `json:"tokens"`
}

type DecodeResponse struct {
	Text string `json:"text"`
}

func runServer(modelPath, port string, gpt4 bool) {

	fmt.Printf("loading model from %s\n", modelPath)
	t, err := bpe.Load(modelPath)
	if err != nil {
		log.Fatalf("failed to load to the model %s\n", modelPath)
	}
	fmt.Println("loading done")

	http.HandleFunc("/encode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		var req TokenizeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		start := time.Now()
		tokens, err := t.Encode(req.Text, gpt4)
		if err != nil {
			http.Error(w, " bad request regex failed ", http.StatusBadRequest)
			return
		}
		duration := time.Since(start).Seconds()

		resp := TokenizeResponse{
			Tokens: tokens,
			Count:  len(tokens),
			Time:   duration,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/decode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		var req DecodeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		decodedText := t.Decode(req.Tokens)

		resp := DecodeResponse{
			Text: decodedText,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	fmt.Printf("the tokenizer server running on port : %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
