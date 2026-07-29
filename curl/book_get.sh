#!/bin/bash
# GET /books/:id
# Retrieves a single book by its ObjectID.
# Usage: ./book_get.sh <book-id>
#   e.g. ./book_get.sh 6849f1a2c3d4e5f6a7b8c9d0

BASE_URL="http://localhost:8080"
BOOK_ID="${1:?Usage: $0 <book-id>}"

curl -s -X GET "$BASE_URL/books/$BOOK_ID" \
  -H "Accept: application/json" | jq .
