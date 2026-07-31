#!/bin/bash
# POST /books
# Creates a new book. Title and at least one author are required.
# Returns: {"id": "<hex-object-id>"}

BASE_URL="http://localhost:8080"

curl -s -X POST "$BASE_URL/books" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Dune",
    "authors": ["Frank Herbert"],
    "publisher": "Chilton Books",
    "published_year": 1965,
    "isbn10": "0441013597",
    "isbn13": "9780441013593",
    "description": "A science fiction epic set on the desert planet Arrakis.",
    "genre": ["Science Fiction", "Adventure"]
  }' | jq .
