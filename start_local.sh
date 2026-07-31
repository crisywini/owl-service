#!/usr/bin/env bash
set -e

CONTAINER_NAME="owl-mongo"
MONGO_PORT=27017

echo "==> Stopping any existing MongoDB containers..."
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
  docker stop "${CONTAINER_NAME}" >/dev/null 2>&1 || true
  docker rm "${CONTAINER_NAME}" >/dev/null 2>&1 || true
  echo "    Removed existing container: ${CONTAINER_NAME}"
fi

# Also stop any other container that might be occupying the port
CONFLICTING=$(docker ps --filter "publish=${MONGO_PORT}" --format '{{.Names}}')
if [ -n "$CONFLICTING" ]; then
  echo "    Stopping container occupying port ${MONGO_PORT}: ${CONFLICTING}"
  docker stop $CONFLICTING >/dev/null 2>&1 || true
fi

echo "==> Starting MongoDB container..."
docker run -d \
  --name "${CONTAINER_NAME}" \
  -p "${MONGO_PORT}:27017" \
  mongo:latest

echo "==> Waiting for MongoDB to be ready..."
MAX_RETRIES=30
COUNT=0
until docker exec "${CONTAINER_NAME}" mongosh --quiet --eval "db.runCommand({ ping: 1 })" >/dev/null 2>&1; do
  COUNT=$((COUNT + 1))
  if [ "$COUNT" -ge "$MAX_RETRIES" ]; then
    echo "ERROR: MongoDB did not become ready in time."
    exit 1
  fi
  echo "    Still waiting... (${COUNT}/${MAX_RETRIES})"
  sleep 1
done

echo "==> MongoDB is up and running."

export MONGODB_URI="mongodb://localhost:${MONGO_PORT}"
echo "==> MONGODB_URI set to: ${MONGODB_URI}"

echo "==> Starting owl-service API..."
cd "$(dirname "$0")"
go run ./cmd/main.go
