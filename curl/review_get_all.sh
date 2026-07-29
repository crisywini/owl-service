#!/bin/bash
# GET /reviews
# Returns every review in the system.

BASE_URL="http://localhost:8080"

curl -s -X GET "$BASE_URL/reviews" \
  -H "Accept: application/json" | jq .
