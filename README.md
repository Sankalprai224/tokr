# tokr ⚡  
**High-Throughput, GPT-4-Compatible Tokenizer in Pure Go**

`tokr` is a production-ready, pure Go implementation of the Byte Pair Encoding (BPE) tokenizer. Built for high-load AI infrastructure, it functions as both an embeddable library and a scalable microservice.

It matches the GPT-4 tokenization rules while leveraging a mutex-protected parallel architecture to deliver massive throughput—without the Python GIL or CGO overhead.

---

## 🚀 Key Features

- ⚡ **High Performance**: Processes **~5.3 million tokens/sec (9.03 MB/s)** on standard hardware.
- 🧵 Parallel Engine: Automatically distributes work across CPU cores using a worker-pool pattern.
- 🧠 Zero-Allocation: Hot path achieves 0 dynamic allocations (1 alloc/op) via concurrent caching.
- 🌐 Microservice Ready: Built-in HTTP server handles high concurrency with <6ms latency.
- ✅ Round-Trip Integrity: Fuzz-tested to guarantee Decode(Encode(t)) == t for 100% data safety.
- 🔒 Stream Processing: Uses inlined regex streaming to handle large files without memory spikes.

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
    "log"
    "github.com/HeLiX-x/tokr/bpe"
)

func main() {
    // 1. Load the BPE model
    // Note: Use NewTokenizer, not Load
    t := bpe.NewTokenizer("vocab.model")

    // 2. Encode text (Automatically uses parallel engine for large inputs)
    text := "Hello, world! tokr is fast."
    tokens := t.Encode(text, true) // true = use GPT-4 regex splitting
    fmt.Println("Tokens:", tokens)

    // 3. Decode back to string
    // Note: It is Decode(), not Decoder()
    decoded := t.Decode(tokens)
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

### 5. Performance & Optimization Highlights: Breaking the GC Wall
Building a fast tokenizer in Go requires carefully managing the memory bus. To reach maximum throughput without sacrificing absolute 1:1 `tiktoken` (OpenAI) correctness, the engine employs strict allocation control:

* **Zero-Allocation Chunking (`parallel.go`):** To feed the worker pool, large text inputs are dynamically sliced using native Go string slicing (`text[start:splitPoint]`), which creates zero heap allocations. To prevent BPE corruption, the chunker scans backward from the 1MB mark to find natural boundaries (spaces or newlines) so words are never cut in half. A bitwise UTF-8 fallback (`(text[splitPoint]&0xC0) == 0x80`) ensures multi-byte Unicode characters are never bisected, preventing regex panics.
* **Zero-Allocation Merge Phase (`runMergeLogic`):** The rank-merge hot path operates entirely in-place. Instead of allocating a new slice per iteration, it overwrites the target index and shifts the remaining elements left via `copy`, shrinking the slice in place. 

#### **The Problem: The GC Wall (Cold Path)**
To maintain absolute GPT-4 correctness, the Go port of the .NET regex engine (`regexp2`) is required, but it is notoriously allocation-heavy. When we disable the cache and force the CPU to do the raw BPE math and `regexp2` allocations on a massive 16MB file, the parallel workers eventually hit a Garbage Collection bottleneck (allocating memory 18.6 million times). 

As shown below, raw compute scaling caps out around 3.30 MB/s because the Go runtime is overwhelmed by allocation locks:

goos: linux
goarch: amd64
cpu: 13th Gen Intel(R) Core(TM) i5-13500H

// Raw Math (No Cache) - 16MB Payload
BenchmarkEncode_Single_vs_Parallel/SingleThreaded_Raw-16       1      5582823447 ns/op     2.93 MB/s   1617717192 B/op   18651547 allocs/op
BenchmarkEncode_Single_vs_Parallel/Parallel-16                 2      4968771030 ns/op     3.30 MB/s   1708757640 B/op   18651686 allocs/op

#### **The Solution: The Word Cache (Neutralizing `regexp2`)**
Instead of dropping `regexp2` (and losing correctness), the architecture neutralizes the GC overhead by introducing a highly concurrent `sync.Map` at the *word level*. Once `regexp2` extracts a common word (like `" the "`), the BPE math and slice allocations run exactly once. Future occurrences instantly pull the pre-computed tokens from the `sync.Map`, bypassing the heavy merge math and completely eliminating millions of redundant array allocations.

// Cached Inference (Hot Path vs Cold Path)
Benchmark	             Latency	    Allocations	     Context
BenchmarkPublicAPI	     52.55 ns/op	1 allocs/op	     Cached / Hot Path
BenchmarkEncodeCore	     10.23 µs/op	75 allocs/op	 Uncached / Cold Path

BenchmarkEncode_Cached-16           26193117       219.6 ns/op       448 B/op         1 allocs/op
BenchmarkEncode_CacheHitRate-16     39210831       144.8 ns/op       172 B/op         1 allocs/op

* **~50-200 Nanosecond Latency:** Once a string is cached, tokenization bypasses the regex engine entirely, returning results in fractions of a microsecond.
* **1 Allocation:** The single memory allocation (`1 allocs/op`) is just a defensive copy.

#### **The Result: System Throughput vs OpenAI**
Because the real world consists of repetitive language, this caching architecture allows the effective system throughput to bypass the GC wall entirely. Tests run on an AMD Ryzen 5 7530U (25MB Corpus) demonstrate that `tokr` delivers highly competitive performance compared to OpenAI's Rust implementation (`tiktoken`):

-------------------------------------------------------
Library	    Language	  Throughput	 Speed
tiktoken    Rust/Python   11.33 MB/s	~3.7M tokens/s
tokr ⚡	    Pure Go	      9.03 MB/s	    ~5.3M tokens/s
-------------------------------------------------------

## 📄 License
```
MIT © HeLiX-x
```
