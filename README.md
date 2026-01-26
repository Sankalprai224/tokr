# tokr ⚡  
**High-Throughput, GPT-4-Compatible Tokenizer in Pure Go**

`tokr` is a production-ready, pure Go implementation of the **Byte Pair Encoding (BPE)** tokenizer used in **GPT-2/GPT-4**. Built for AI infrastructure, it functions as both a high-performance embeddable library and a scalable microservice.

It reproduces OpenAI’s `tiktoken` output **bit-for-bit**, while leveraging Go’s concurrency model to deliver massive throughput—**without the Python GIL**.

---

## 🚀 Key Features

- ⚡ **High Performance**: Processes **~2.17 million tokens/sec** on standard hardware.
- 🧠 **Memory Efficient**: Zero-allocation hot path using `sync.Pool` for merge buffers.
- 🌐 **Microservice Ready**: Built-in HTTP server handles **~8,500 RPS** at **<6ms latency**.
- ✅ **GPT-4 Compatible**: Implements the official regex split pattern (`'s|'t|'re|...`).
- 🔒 **Concurrency Safe**: Fine-grained locking enables safe parallel encoding across goroutines.
- 💾 **Smart Caching**: Bounded pre-tokenization cache leverages **Zipf’s law** for common words.

---

## 📦 Installation

### As a Go Library
```bash
go get github.com/HeLiX-x/tokr
```
Build the CLI Tool from Source
```bash
git clone https://github.com/HeLiX-x/tokr.git
cd tokr
go build -o tokr .
```

## 🛠️ Usage
1. As a Go Library
Embed tokr directly into RAG pipelines, inference servers, or LLM tooling.

```go
package main

import (
    "fmt"
    "github.com/HeLiX-x/tokr/bpe"
)

func main() {
    // Load the BPE model (e.g., vocab.model)
    t, err := bpe.Load("vocab.model")
    if err != nil {
        panic(err)
    }

    // Encode text (thread-safe)
    text := "Hello, world! tokr is fast."
    tokens := t.Encode(text, true) // true = use GPT-4 regex splitting
    fmt.Println("Tokens:", tokens)

    // Decode back to string
    decoded := t.Decoder(tokens)
    fmt.Println("Text:", decoded)
}
```
 2. CLI Tool
The main.go entry point supports three modes: train, inference, and server.
Train a BPE Model
Train a fresh vocabulary on your corpus:
```
./tokr -mode train -input data.txt -model vocab.model -vocab 10000
```

Run Inference (Benchmark)
Test tokenizer speed on a file:
```
./tokr -mode inference -model vocab.model -input test.txt
```

Start HTTP Server
Launch the microservice for remote tokenization:
```
./tokr -mode server -port 8080 -model vocab.model
```

## 🌐 API Reference (Server Mode)
Once running, interact via JSON over HTTP.
Endpoint: POST /encode
Request:

```json
{
  "text": "The quick brown fox jumps over the lazy dog"
}
```

Response:
```json
{
  "tokens": [464, 2068, 7586, 21831, 18045, 625, 262, 16931, 3290],
  "count": 9,
  "time_seconds": 0.000012
}
```

Benchmark with cURL:
```bash
curl -X POST http://localhost:8080/encode \
     -H "Content-Type: application/json" \
     -d '{"text": "Hello world"}'
```

## 📊 Benchmarks
Tests run on an 8-core Linux machine with Go 1.23.
Library Throughput
```
Metric                  Result
----------------------------------------------------
Speed ->                2.17 million tokens/sec      
Throughput ->           ~8.5 MB/sec (Raw Text)   
Cache Hit Rate ->       ~90% (Realistic Workload)) 
----------------------------------------------------
```

Server Throughput:
```
Metric	                Result
---------------------------------------------------------------
RPS	                    8,479 req/sec (50 concurrent clients)
Latency	                5.90 ms (Avg)
Throughput	            1.10 million tokens/sec (over HTTP)
---------------------------------------------------------------
```
## 🧩 Technical Architecture:
```
    Splitter (splitter.go): Uses dlclark/regexp2
     to strictly adhere to the GPT-4 regex pattern, ensuring identical token boundaries.
    Core BPE Engine (api.go):
        Hot Path: Read-locks (RLock) for high concurrency.
        Memory: Reuses integer slices via sync.Pool to minimize GC pressure.
        Cache: Stores precomputed token sequences for frequent substrings (words/prefixes), skipping redundant merges.
```
## 📄 License
```
MIT © HeLiX-x
```
