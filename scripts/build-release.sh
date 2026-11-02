#!/usr/bin/env bash
set -euo pipefail

# Run from the checked-out release commit, never from the workflow's default ref.
output_dir="${1:?usage: build-release.sh OUTPUT_DIRECTORY}"
mkdir -p "$output_dir"
output_dir="$(cd "$output_dir" && pwd)"

git rev-parse HEAD > "$output_dir/source-commit.txt"
for arch in amd64 arm64; do
  CGO_ENABLED=0 GOOS=linux GOARCH="$arch" go build \
    -trimpath -ldflags='-w -s' -o "$output_dir/sync-issue-field_linux_$arch" ./cmd/run
done

(
  cd "$output_dir"
  sha256sum source-commit.txt sync-issue-field_linux_amd64 sync-issue-field_linux_arm64 > SHA256SUMS
)
