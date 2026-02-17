import pytest
import requests
from hypothesis import given, settings, strategies as st

# Configuration
GO_SERVER_URL = "http://localhost:8080"

# Helper functions to talk to your Go server
def go_encode(text):
    response = requests.post(f"{GO_SERVER_URL}/encode", json={"text": text})
    response.raise_for_status()
    return response.json()["tokens"]

def go_decode(tokens):
    response = requests.post(f"{GO_SERVER_URL}/decode", json={"tokens": tokens})
    response.raise_for_status()
    return response.json()["text"]

# --- TEST: Property-Based Testing (Round Trip) ---
@settings(max_examples=1000)
@given(st.text())
def test_roundtrip_integrity(text):
    """
    Verifies that Go_Decode(Go_Encode(text)) == text
    """
    try:
        tokens = go_encode(text)
        decoded = go_decode(tokens)
        assert decoded == text, f"Roundtrip failed.\nOriginal: {repr(text)}\nDecoded:  {repr(decoded)}"
    except requests.exceptions.ConnectionError:
        pytest.fail("Could not connect to Go server. Is it running on port 8080?")
