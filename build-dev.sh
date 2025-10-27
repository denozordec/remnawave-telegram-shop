#!/usr/bin/env bash

set -euo pipefail

IMAGE_BASE="ghcr.io/jolymmiels/remnawave-telegram-shop-bot"
DEV_TAG="dev"

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg VERSION="${DEV_TAG}" \
  -t "${IMAGE_BASE}:${DEV_TAG}" \
  --push \
  .
