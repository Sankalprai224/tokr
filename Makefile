.PHONY: test bench fuzz clean

# Runs the standard unit tests
test:
	go test -v ./...

# Runs the benchmarks with memory profiling
bench:
	go test -bench=. -benchmem ./bpe

# Runs the fuzzer for 10 seconds to ensure no crashes
fuzz:
	go test -fuzz=Fuzz -fuzztime=10s ./bpe

# Cleans up the Go test cache and any generated binaries
clean:
	go clean -testcache
	rm -f tokr