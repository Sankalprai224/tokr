import random

# diverse_corpus.py
SIZES = {
    "prose": 5000,   # 5000 paragraphs of text
    "code": 2000,    # 2000 snippets of code
    "unicode": 1000, # 1000 lines of mixed scripts
}

SAMPLES = {
    "prose": [
        "The quick brown fox jumps over the lazy dog.",
        "Generative pre-trained transformers are a type of large language model.",
        "In computer science, a B-tree is a self-balancing tree data structure.",
        "The economy of the 21st century is driven by information technology."
    ],
    "code": [
        "def main():\n    print('hello world')",
        "func (t *tokenizer) Encode(text string) []int { return nil }",
        "console.log(`User ${user.id} logged in at ${Date.now()}`);",
        "SELECT * FROM users WHERE last_login > NOW() - INTERVAL '1 day';"
    ],
    "unicode": [
        "こんにちは世界 (Hello World)", 
        "😊 😂 🥺 😉", 
        "Café, Naïve, Entrepôt, Façade", 
        "The price is £10.50 or €12.00 depending on exchange rates."
    ]
}

def generate_file(filename="realistic_bench.txt"):
    print(f"Generating diverse benchmark data into {filename}...")
    with open(filename, "w", encoding="utf-8") as f:
        # 1. Standard Prose (The bulk)
        for _ in range(SIZES["prose"]):
            # Combine random sentences to make paragraphs
            para = " ".join(random.choices(SAMPLES["prose"], k=random.randint(3, 10)))
            f.write(para + "\n\n")
        
        # 2. Code Blocks (Harder for regex splitters)
        for _ in range(SIZES["code"]):
            code = random.choice(SAMPLES["code"])
            f.write(f"```\n{code}\n```\n")

        # 3. Unicode/Edge Cases (Stress test byte handling)
        for _ in range(SIZES["unicode"]):
            uni = random.choice(SAMPLES["unicode"])
            f.write(uni + "\n")

    print("Done! File created.")

if __name__ == "__main__":
    generate_file()