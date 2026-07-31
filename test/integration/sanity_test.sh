#!/bin/bash
# Integration sanity test for owl-service.
# Runs the full CRUD flow: book + reviews, then cleans up.

BASE_URL="http://localhost:8080"

GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
FAILURES=0
SPINNER_PID=""
STARTED_API=false
START_PID=""

# ─── cleanup ─────────────────────────────────────────────────────────────────

cleanup() {
  stop_spinner
  if [ "$STARTED_API" = "true" ]; then
    printf "\n  🛑  ${BOLD}Shutting down API...${NC}\n"
    # kill start_local.sh children (go run + compiled binary), then the script itself
    pkill -P "$START_PID" 2>/dev/null || true
    kill "$START_PID" 2>/dev/null || true
    # catch any orphaned go run process
    pkill -f "go run ./cmd/main.go" 2>/dev/null || true
    # stop the MongoDB container that start_local.sh launched
    docker stop owl-mongo >/dev/null 2>&1 || true
    printf "     ${GREEN}✓${NC}  API and database stopped\n\n"
  fi
}

trap cleanup EXIT

# ─── spinner ─────────────────────────────────────────────────────────────────

start_spinner() {
  local msg="$1"
  (
    local frames=('⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏')
    local i=0
    while true; do
      printf "\r     ${CYAN}${frames[$i]}${NC}  ${DIM}%s${NC}" "$msg"
      i=$(( (i + 1) % 10 ))
      sleep 0.08
    done
  ) &
  SPINNER_PID=$!
}

stop_spinner() {
  if [ -n "$SPINNER_PID" ]; then
    kill "$SPINNER_PID" 2>/dev/null
    wait "$SPINNER_PID" 2>/dev/null
    printf "\r\033[2K"
    SPINNER_PID=""
  fi
}

# ─── helpers ─────────────────────────────────────────────────────────────────

pass() { printf "     ✅  ${GREEN}[PASSED]${NC} %s\n" "$1"; }
fail() { printf "     ❌  ${RED}[FAILED]${NC} %s\n" "$1"; FAILURES=$((FAILURES + 1)); }

section() {
  sleep 0.4
  printf "\n  ${CYAN}${BOLD}[ %s ]${NC}  %s  ${BOLD}%s${NC}\n" "$1" "$2" "$3"
}

api_status() {
  curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/books" 2>/dev/null
}

# ─── header ──────────────────────────────────────────────────────────────────

printf "\n"
printf "  ${BOLD}🦉  Owl Service — Integration Sanity Check${NC}\n"
printf "  ${DIM}%s${NC}\n" "$BASE_URL"
printf "\n"

# ─── 1. Ensure API is up ─────────────────────────────────────────────────────

printf "  🔌  ${BOLD}Checking if API is running...${NC}\n"

STATUS=$(api_status)
if [[ "$STATUS" =~ ^2 ]]; then
  printf "     ${GREEN}✓${NC}  API already up and running (HTTP %s) — will not shut it down after tests\n" "$STATUS"
else
  printf "     API not reachable (HTTP %s). Starting now...\n" "$STATUS"
  "$SCRIPT_DIR/../../start_local.sh" &
  START_PID=$!
  STARTED_API=true

  start_spinner "Waiting for API to become ready..."
  MAX=60
  COUNT=0
  until [[ "$(api_status)" =~ ^2 ]]; do
    COUNT=$((COUNT + 1))
    if [ "$COUNT" -ge "$MAX" ]; then
      stop_spinner
      fail "API did not start in time"
      exit 1
    fi
    sleep 2
  done
  stop_spinner
  printf "     ${GREEN}✓${NC}  API is ready — will shut it down after tests\n"
fi

if ! command -v jq &>/dev/null; then
  printf "  ${RED}ERROR:${NC} 'jq' is required but not installed.\n"
  exit 1
fi

sleep 0.3
printf "\n  ${DIM}────────────────────────────────────────────────${NC}\n"
sleep 0.4

# ─── 2. Create book ──────────────────────────────────────────────────────────

section "1" "📚" "Create book"
start_spinner "Registering a new book in the library..."
BOOK_RESP=$(curl -s -X POST "$BASE_URL/books" \
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
  }')
sleep 0.7
stop_spinner

BOOK_ID=$(echo "$BOOK_RESP" | jq -r '.id // empty')
if [ -n "$BOOK_ID" ]; then
  pass "Create book — id: $BOOK_ID"
else
  fail "Create book — response: $BOOK_RESP"
fi

# ─── 3. Get book by id ───────────────────────────────────────────────────────

section "2" "🔍" "Get book by id"
if [ -z "$BOOK_ID" ]; then
  fail "Get book — skipped (no book id)"
else
  start_spinner "Looking up book by id..."
  GET_BOOK_RESP=$(curl -s -X GET "$BASE_URL/books/$BOOK_ID" -H "Accept: application/json")
  sleep 0.6
  stop_spinner
  GET_BOOK_ID=$(echo "$GET_BOOK_RESP" | jq -r '.id // empty')
  if [ "$GET_BOOK_ID" = "$BOOK_ID" ]; then
    pass "Get book — id matches"
  else
    fail "Get book — response: $GET_BOOK_RESP"
  fi
fi

# ─── 4. Get all books ────────────────────────────────────────────────────────

section "3" "📖" "Get all books"
start_spinner "Scanning the full book catalog..."
ALL_BOOKS_RESP=$(curl -s -X GET "$BASE_URL/books" -H "Accept: application/json")
sleep 0.8
stop_spinner
if echo "$ALL_BOOKS_RESP" | jq -e 'type == "array"' >/dev/null 2>&1; then
  BOOK_COUNT=$(echo "$ALL_BOOKS_RESP" | jq 'length')
  pass "Get all books — $BOOK_COUNT book(s) returned"
else
  fail "Get all books — response: $ALL_BOOKS_RESP"
fi

# ─── 5. Update book ──────────────────────────────────────────────────────────

section "4" "✏️ " "Update book"
if [ -z "$BOOK_ID" ]; then
  fail "Update book — skipped (no book id)"
else
  start_spinner "Applying updates to the book record..."
  UPD_BOOK_RESP=$(curl -s -X PUT "$BASE_URL/books/$BOOK_ID" \
    -H "Content-Type: application/json" \
    -d '{
      "title": "Dune Messiah",
      "publisher": "Putnam",
      "published_year": 1969,
      "description": "The second book in the Dune saga.",
      "genre": ["Science Fiction"]
    }')
  sleep 0.7
  stop_spinner
  UPD_BOOK_ID=$(echo "$UPD_BOOK_RESP" | jq -r '.id // empty')
  if [ -n "$UPD_BOOK_ID" ]; then
    pass "Update book — updated successfully"
  else
    fail "Update book — response: $UPD_BOOK_RESP"
  fi
fi

# ─── 6. Create review 1 ──────────────────────────────────────────────────────

section "5" "✍️ " "Create review 1"
REVIEW1_ID=""
if [ -z "$BOOK_ID" ]; then
  fail "Create review 1 — skipped (no book id)"
else
  start_spinner "Submitting first review..."
  REV1_RESP=$(curl -s -X POST "$BASE_URL/reviews" \
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
    }")
  sleep 0.7
  stop_spinner
  REVIEW1_ID=$(echo "$REV1_RESP" | jq -r '.id // empty')
  if [ -n "$REVIEW1_ID" ]; then
    pass "Create review 1 — id: $REVIEW1_ID"
  else
    fail "Create review 1 — response: $REV1_RESP"
  fi
fi

# ─── 7. Get review 1 ─────────────────────────────────────────────────────────

section "6" "🔎" "Get review 1"
if [ -z "$REVIEW1_ID" ]; then
  fail "Get review 1 — skipped (no review id)"
else
  start_spinner "Fetching review by id..."
  GET_REV1_RESP=$(curl -s -X GET "$BASE_URL/reviews/$REVIEW1_ID" -H "Accept: application/json")
  sleep 0.6
  stop_spinner
  GET_REV1_ID=$(echo "$GET_REV1_RESP" | jq -r '.id // empty')
  if [ "$GET_REV1_ID" = "$REVIEW1_ID" ]; then
    pass "Get review 1 — id matches"
  else
    fail "Get review 1 — response: $GET_REV1_RESP"
  fi
fi

# ─── 8. Create review 2 ──────────────────────────────────────────────────────

section "7" "✍️ " "Create review 2"
REVIEW2_ID=""
if [ -z "$BOOK_ID" ]; then
  fail "Create review 2 — skipped (no book id)"
else
  start_spinner "Submitting second review..."
  REV2_RESP=$(curl -s -X POST "$BASE_URL/reviews" \
    -H "Content-Type: application/json" \
    -d "{
      \"book_id\": \"$BOOK_ID\",
      \"rate\": 3,
      \"description\": \"Good but slow in the middle. Worth reading once.\",
      \"start_date\": \"2024-02-01T00:00:00Z\",
      \"finish_date\": \"2024-02-28T00:00:00Z\",
      \"favorite_phrases\": [
        \"He who controls the spice controls the universe.\"
      ]
    }")
  sleep 0.7
  stop_spinner
  REVIEW2_ID=$(echo "$REV2_RESP" | jq -r '.id // empty')
  if [ -n "$REVIEW2_ID" ]; then
    pass "Create review 2 — id: $REVIEW2_ID"
  else
    fail "Create review 2 — response: $REV2_RESP"
  fi
fi

# ─── 9. Get all reviews ──────────────────────────────────────────────────────

section "8" "📋" "Get all reviews"
start_spinner "Loading all reviews from the database..."
ALL_REVS_RESP=$(curl -s -X GET "$BASE_URL/reviews" -H "Accept: application/json")
sleep 0.8
stop_spinner
if echo "$ALL_REVS_RESP" | jq -e 'type == "array"' >/dev/null 2>&1; then
  REV_COUNT=$(echo "$ALL_REVS_RESP" | jq 'length')
  pass "Get all reviews — $REV_COUNT review(s) returned"
else
  fail "Get all reviews — response: $ALL_REVS_RESP"
fi

# ─── 10. Update review 1 ─────────────────────────────────────────────────────

section "9" "✏️ " "Update review 1"
if [ -z "$REVIEW1_ID" ]; then
  fail "Update review 1 — skipped (no review id)"
else
  start_spinner "Applying changes to review..."
  UPD_REV_RESP=$(curl -s -X PUT "$BASE_URL/reviews/$REVIEW1_ID" \
    -H "Content-Type: application/json" \
    -d '{
      "rate": 4,
      "description": "Revised: still great, but pacing drags in part three.",
      "start_date": "2024-01-01T00:00:00Z",
      "finish_date": "2024-01-20T00:00:00Z",
      "favorite_phrases": ["The spice must flow."]
    }')
  sleep 0.7
  stop_spinner
  UPD_REV_ID=$(echo "$UPD_REV_RESP" | jq -r '.id // empty')
  if [ -n "$UPD_REV_ID" ]; then
    pass "Update review 1 — updated successfully"
  else
    fail "Update review 1 — response: $UPD_REV_RESP"
  fi
fi

# ─── 11. Delete review 1 ─────────────────────────────────────────────────────

section "10" "🗑️ " "Delete review 1"
if [ -z "$REVIEW1_ID" ]; then
  fail "Delete review 1 — skipped (no review id)"
else
  start_spinner "Removing review 1..."
  DEL_REV1=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/reviews/$REVIEW1_ID")
  sleep 0.6
  stop_spinner
  if [ "$DEL_REV1" = "204" ]; then
    pass "Delete review 1 — 204 No Content"
  else
    fail "Delete review 1 — HTTP $DEL_REV1"
  fi
fi

# ─── 12. Delete review 2 ─────────────────────────────────────────────────────

section "11" "🗑️ " "Delete review 2"
if [ -z "$REVIEW2_ID" ]; then
  fail "Delete review 2 — skipped (no review id)"
else
  start_spinner "Removing review 2..."
  DEL_REV2=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/reviews/$REVIEW2_ID")
  sleep 0.6
  stop_spinner
  if [ "$DEL_REV2" = "204" ]; then
    pass "Delete review 2 — 204 No Content"
  else
    fail "Delete review 2 — HTTP $DEL_REV2"
  fi
fi

# ─── 13. Delete book ─────────────────────────────────────────────────────────

section "12" "🗑️ " "Delete book"
if [ -z "$BOOK_ID" ]; then
  fail "Delete book — skipped (no book id)"
else
  start_spinner "Removing book from the library..."
  DEL_BOOK=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/books/$BOOK_ID")
  sleep 0.7
  stop_spinner
  if [ "$DEL_BOOK" = "204" ]; then
    pass "Delete book — 204 No Content"
  else
    fail "Delete book — HTTP $DEL_BOOK"
  fi
fi

# ─── summary ─────────────────────────────────────────────────────────────────

sleep 0.3
printf "\n  ${DIM}────────────────────────────────────────────────${NC}\n"
sleep 0.2
printf "\n"
if [ "$FAILURES" -eq 0 ]; then
  printf "  🎉  ${GREEN}${BOLD}All 12 tests passed!${NC}\n\n"
else
  printf "  💥  ${RED}${BOLD}%d / 12 test(s) failed.${NC}\n\n" "$FAILURES"
fi

[ "$FAILURES" -eq 0 ]
