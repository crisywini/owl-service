#!/bin/bash
# GET /reviews/:id
# Retrieves a single review by its ObjectID.
# Usage: ./review_get.sh <review-id>
#   e.g. ./review_get.sh 6849f1a2c3d4e5f6a7b8c9d0

BASE_URL="http://localhost:8080"
REVIEW_ID="${1:?Usage: $0 <review-id>}"

curl -s -X GET "$BASE_URL/reviews/$REVIEW_ID" \
  -H "Accept: application/json" | jq .
