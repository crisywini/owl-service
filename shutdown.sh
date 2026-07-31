#!/usr/bin/env bash
set -e

NETWORK="owl-network"
MONGO_CONTAINER="owl-mongo"
API_CONTAINER="owl-service"
IMAGE_NAME="owl-service"

log() { echo "==> $*"; }

remove_container() {
  local name="$1"
  if docker ps -a --format '{{.Names}}' | grep -q "^${name}$"; then
    log "Stopping and removing container '${name}'..."
    docker stop "${name}" >/dev/null 2>&1 || true
    docker rm   "${name}" >/dev/null 2>&1 || true
    echo "    Done."
  else
    echo "    Container '${name}' not found, skipping."
  fi
}

# ---------------------------------------------------------------------------
# 1. Stop and remove containers
# ---------------------------------------------------------------------------
remove_container "${API_CONTAINER}"
remove_container "${MONGO_CONTAINER}"

# ---------------------------------------------------------------------------
# 2. Remove all local owl-service images
# ---------------------------------------------------------------------------
log "Removing all '${IMAGE_NAME}' images..."
IMAGES=$(docker images --format '{{.Repository}}:{{.Tag}}' | grep "^${IMAGE_NAME}:")
if [ -n "${IMAGES}" ]; then
  echo "${IMAGES}" | xargs docker rmi -f
  echo "    Removed: ${IMAGES}"
else
  echo "    No '${IMAGE_NAME}' images found, skipping."
fi

# ---------------------------------------------------------------------------
# 3. Remove the Docker network
# ---------------------------------------------------------------------------
log "Removing network '${NETWORK}'..."
if docker network inspect "${NETWORK}" >/dev/null 2>&1; then
  docker network rm "${NETWORK}"
  echo "    Done."
else
  echo "    Network '${NETWORK}' not found, skipping."
fi

log "Shutdown complete."
