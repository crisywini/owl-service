#!/bin/bash
# DELETE /reviews/:id
# Removes a review by its ObjectID.
# Returns: 204 No Content on success.
# Usage: ./review_delete.sh <review-id>
#   e.g. ./review_delete.sh 6849f1a2c3d4e5f6a7b8c9d0

BASE_URL="http://localhost:8080"
REVIEW_ID="${1:?Usage: $0 <review-id>}"

curl -s -o /dev/null -w "HTTP status: %{http_code}\n" \
  -X DELETE "$BASE_URL/reviews/$REVIEW_ID"
