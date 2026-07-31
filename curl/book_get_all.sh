#!/bin/bash
# GET /books
# Returns every book in the system.

BASE_URL="http://localhost:8080"

curl -s -X GET "$BASE_URL/books" \
  -H "Accept: application/json" | jq .
