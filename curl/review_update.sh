#!/bin/bash
# PUT /reviews/:id
# Updates mutable fields of a review.
# rate (> 0) and description are required.
# book_id cannot be changed after creation.
# Usage: ./review_update.sh <review-id>
#   e.g. ./review_update.sh 6849f1a2c3d4e5f6a7b8c9d0

BASE_URL="http://localhost:8080"
REVIEW_ID="${1:?Usage: $0 <review-id>}"

curl -s -X PUT "$BASE_URL/reviews/$REVIEW_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "rate": 4,
    "description": "Revised opinion: still great, but the pacing drags in part three.",
    "start_date": "2024-01-01T00:00:00Z",
    "finish_date": "2024-01-20T00:00:00Z",
    "favorite_phrases": [
      "The spice must flow."
    ]
  }' | jq .
