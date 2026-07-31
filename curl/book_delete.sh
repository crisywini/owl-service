#!/bin/bash
# DELETE /books/:id
# Removes a book by its ObjectID.
# Returns: 204 No Content on success.
# Usage: ./book_delete.sh <book-id>
#   e.g. ./book_delete.sh 6849f1a2c3d4e5f6a7b8c9d0

BASE_URL="http://localhost:8080"
BOOK_ID="${1:?Usage: $0 <book-id>}"

curl -s -o /dev/null -w "HTTP status: %{http_code}\n" \
  -X DELETE "$BASE_URL/books/$BOOK_ID"
