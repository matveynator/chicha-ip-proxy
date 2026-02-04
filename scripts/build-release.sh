#!/usr/bin/env bash

# This script mirrors the CI matrix so local builds stay consistent and cgo-free.
set -euo pipefail

version="${1:-dev}"
project_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist_dir="${project_root}/dist"

# Keeping the targets explicit makes it easy to reason about which binaries we ship.
targets=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
  "freebsd/amd64"
  "freebsd/arm64"
  "openbsd/amd64"
  "openbsd/arm64"
)

mkdir -p "${dist_dir}"

# Build sequentially to keep output logs readable while avoiding shared state.
for target in "${targets[@]}"; do
  os="${target%/*}"
  arch="${target#*/}"
  ext=""
  if [ "${os}" = "windows" ]; then
    ext=".exe"
  fi

  echo "Building ${os}/${arch}..."
  CGO_ENABLED=0 GOOS="${os}" GOARCH="${arch}" \
    go build -tags netgo \
    -ldflags "-X github.com/matveynator/chicha-ip-proxy/pkg/version.Number=${version}" \
    -o "${dist_dir}/chicha-ip-proxy-${version}-${os}-${arch}${ext}" \
    "${project_root}"
done
