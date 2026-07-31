#!/bin/bash
# PUT /books/:id
# Updates mutable fields of a book. Title must be non-empty.
# Authors cannot be changed after creation.
# Usage: ./book_update.sh <book-id>
#   e.g. ./book_update.sh 6849f1a2c3d4e5f6a7b8c9d0

BASE_URL="http://localhost:8080"
BOOK_ID="${1:?Usage: $0 <book-id>}"

curl -s -X PUT "$BASE_URL/books/$BOOK_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Dune Messiah",
    "publisher": "Putnam",
    "published_year": 1969,
    "description": "The second book in the Dune saga.",
    "genre": ["Science Fiction"]
  }' | jq .
