#!/bin/bash
# POST /reviews
# Creates a new review for an existing book.
# book_id must be a valid ObjectID of an existing book.
# rate (> 0) and description are required.
# Returns: {"id": "<hex-object-id>"}
# Usage: ./review_create.sh <book-id>
#   e.g. ./review_create.sh 6849f1a2c3d4e5f6a7b8c9d0

BASE_URL="http://localhost:8080"
BOOK_ID="${1:?Usage: $0 <book-id>}"

curl -s -X POST "$BASE_URL/reviews" \
  -H "Content-Type: application/json" \
  -d "{
    \"book_id\": \"$BOOK_ID\",
    \"rate\": 5,
    \"description\": \"An absolute masterpiece. The world-building is unparalleled.\",
    \"start_date\": \"2024-01-01T00:00:00Z\",
    \"finish_date\": \"2024-01-15T00:00:00Z\",
    \"favorite_phrases\": [
      \"The spice must flow.\",
      \"I must not fear. Fear is the mind-killer.\"
    ]
  }" | jq .
