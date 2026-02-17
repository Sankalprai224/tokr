import random
import os

FILENAME = "scripts/heavy_test.txt"
TARGET_SIZE = 25 * 1024 * 1024  # 25 MB

# diverse sources to trick the CPU branch predictor
SOURCES = [
    # Prose (English)
    "The quick brown fox jumps over the lazy dog.",
    "Generative pre-trained transformers are a type of large language model.",
    "To be, or not to be, that is the question.",
    "It was the best of times, it was the worst of times.",
    
    # Code (Go/Python)
    "func main() { fmt.Println('Hello World') }",
    "def train(self, data): return [self.encode(x) for x in data]",
    "if (x > 10 && y < 20) { return true; }",
    
    # Structured Data (JSON/Logs)
    '{"id": 123, "event": "login", "timestamp": 16789000}',
    "[INFO] 2024-02-15 12:00:00 Connection established from 192.168.1.1",
    
    # Noise (Hard for Regex)
    "       ",  # Just spaces
    "1234567890",
    "!@#$%^&*()_+",
]

def generate():
    print(f"Generating {TARGET_SIZE / 1024 / 1024:.2f} MB of realistic chaos...")
    
    with open(FILENAME, "w", encoding="utf-8") as f:
        current_size = 0
        while current_size < TARGET_SIZE:
            # 1. Pick a random base text
            text = random.choice(SOURCES)
            
            # 2. Randomly combine it to make variable line lengths
            # (Real files don't have constant line lengths)
            repeat = random.randint(1, 5)
            line = (text + " ") * repeat
            
            # 3. Write it with a newline (CRITICAL for your parallel splitter)
            line += "\n"
            
            f.write(line)
            current_size += len(line)

    print(f"Done! Created {FILENAME}")

if __name__ == "__main__":
    generate()