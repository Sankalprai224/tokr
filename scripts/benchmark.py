import time
import sys
import os
import tiktoken

if len(sys.argv) < 2:
    print("Usage: python benchmark.py <file.txt>")
    sys.exit(1)

filename = sys.argv[1]

# 1. Load Data
with open(filename, 'r', encoding='utf-8') as f:
    text = f.read()

size_mb = len(text.encode('utf-8')) / (1024 * 1024)
print(f"File size: {size_mb:.2f} MB")

# 2. Prepare Tokenizer (using GPT-4 rules)
enc = tiktoken.get_encoding("cl100k_base")

# 3. Benchmark
print("Running tiktoken (Python/Rust) encode...")
start = time.time()
tokens = enc.encode(text)
end = time.time()

duration = end - start
count = len(tokens)

print("--------------------------------")
print(f"Time:       {duration:.4f} s")
print(f"Throughput: {size_mb / duration:.2f} MB/s")
print(f"Speed:      {count/1000/duration:.2f} kTokens/s")
print("--------------------------------")