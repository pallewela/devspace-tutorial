#!/usr/bin/env bash
set -e
IMAGE="$1"
TAG="${2:-latest}"
FULL_IMAGE="${IMAGE}:${TAG}"

buildah bud -t "$FULL_IMAGE" -f "${DOCKERFILE:-./Dockerfile}" "${CONTEXT:-.}"
buildah push "$FULL_IMAGE"
