#!/usr/bin/env bash

set -euo pipefail

IMAGE_BASE="ghcr.io/jolymmiels/remnawave-telegram-shop-bot"

read -rp "Enter release version (e.g. 3.1.3): " VERSION
if [[ -z "${VERSION}" ]]; then
  echo "Version must not be empty" >&2
  exit 1
fi

MAJOR_VERSION="${VERSION%%.*}"
if [[ -z "${MAJOR_VERSION}" ]]; then
  echo "Failed to derive major version from '${VERSION}'" >&2
  exit 1
fi

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg VERSION="${VERSION}" \
  -t "${IMAGE_BASE}:${VERSION}" \
  -t "${IMAGE_BASE}:${MAJOR_VERSION}" \
  -t "${IMAGE_BASE}:latest" \
  --push \
  .
