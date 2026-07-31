#!/usr/bin/env bash
set -e

# ---------------------------------------------------------------------------
# Usage: ./start_with_docker.sh <tag>
#   <tag>  Tag for the Docker image, e.g. "1" → owl-service:1 (default: 1)
# ---------------------------------------------------------------------------

TAG="${1:-1}"
IMAGE="owl-service:${TAG}"
NETWORK="owl-network"
MONGO_CONTAINER="owl-mongo"
API_CONTAINER="owl-service"
API_PORT=8080
MONGO_PORT=27017
MAX_RETRIES=30

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
log() { echo "==> $*"; }
wait_for_http() {
  local url="$1"
  local count=0
  log "Waiting for ${url} ..."
  until curl -sf "${url}" >/dev/null 2>&1; do
    count=$((count + 1))
    if [ "${count}" -ge "${MAX_RETRIES}" ]; then
      echo "ERROR: Service at ${url} did not become ready in time."
      exit 1
    fi
    echo "    Still waiting... (${count}/${MAX_RETRIES})"
    sleep 2
  done
}

# ---------------------------------------------------------------------------
# 1. Ensure Docker network exists
# ---------------------------------------------------------------------------
log "Ensuring Docker network '${NETWORK}' exists..."
if ! docker network inspect "${NETWORK}" >/dev/null 2>&1; then
  docker network create "${NETWORK}"
  echo "    Created network: ${NETWORK}"
else
  echo "    Network already exists: ${NETWORK}"
fi

# ---------------------------------------------------------------------------
# 2. Clean up existing containers
# ---------------------------------------------------------------------------
for CONTAINER in "${MONGO_CONTAINER}" "${API_CONTAINER}"; do
  if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER}$"; then
    log "Removing existing container '${CONTAINER}'..."
    docker stop "${CONTAINER}" >/dev/null 2>&1 || true
    docker rm   "${CONTAINER}" >/dev/null 2>&1 || true
  fi
done

# Also free ports if something else is holding them
for PORT in "${API_PORT}" "${MONGO_PORT}"; do
  CONFLICT=$(docker ps --filter "publish=${PORT}" --format '{{.Names}}')
  if [ -n "${CONFLICT}" ]; then
    log "Stopping container occupying port ${PORT}: ${CONFLICT}"
    docker stop "${CONFLICT}" >/dev/null 2>&1 || true
  fi
done

# ---------------------------------------------------------------------------
# 3. Start MongoDB
# ---------------------------------------------------------------------------
log "Starting MongoDB container '${MONGO_CONTAINER}'..."
docker run -d \
  --name "${MONGO_CONTAINER}" \
  --network "${NETWORK}" \
  -p "${MONGO_PORT}:27017" \
  mongo:latest

log "Waiting for MongoDB to be ready..."
count=0
until docker exec "${MONGO_CONTAINER}" mongosh --quiet --eval "db.runCommand({ ping: 1 })" >/dev/null 2>&1; do
  count=$((count + 1))
  if [ "${count}" -ge "${MAX_RETRIES}" ]; then
    echo "ERROR: MongoDB did not become ready in time."
    exit 1
  fi
  echo "    Still waiting... (${count}/${MAX_RETRIES})"
  sleep 1
done
log "MongoDB is ready."

# ---------------------------------------------------------------------------
# 4. Build the API image
# ---------------------------------------------------------------------------
log "Building Docker image '${IMAGE}'..."
docker build -t "${IMAGE}" .

# ---------------------------------------------------------------------------
# 5. Start the API container
# ---------------------------------------------------------------------------
log "Starting API container '${API_CONTAINER}' (image: ${IMAGE})..."
docker run -d \
  --name "${API_CONTAINER}" \
  --network "${NETWORK}" \
  -p "${API_PORT}:8080" \
  -e "MONGODB_URI=mongodb://${MONGO_CONTAINER}:27017" \
  "${IMAGE}"

# ---------------------------------------------------------------------------
# 6. Health check
# ---------------------------------------------------------------------------
wait_for_http "http://localhost:${API_PORT}/health"
log "API is up. Health check response:"
curl -s "http://localhost:${API_PORT}/health" | cat
echo ""
log "All done. owl-service:${TAG} is running on port ${API_PORT}."
