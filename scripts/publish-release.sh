#!/usr/bin/env bash
set -euo pipefail

tag="${1:?usage: publish-release.sh TAG ASSET_DIRECTORY}"
asset_dir="${2:?usage: publish-release.sh TAG ASSET_DIRECTORY}"
repository="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}"
existing="$(gh release view "$tag" --repo "$repository" --json assets --jq '.assets[].name')"
check_dir="$(mktemp -d)"
trap 'rm -rf "$check_dir"' EXIT

# Recover partial uploads without changing assets consumers may already have used.
# Check all existing files before uploading any missing files.
for path in "$asset_dir/"*; do
  name="${path##*/}"
  if printf '%s\n' "$existing" | grep -Fxq "$name"; then
    gh release download "$tag" --repo "$repository" --pattern "$name" --dir "$check_dir"
    cmp "$path" "$check_dir/$name" || {
      echo "::error::Published asset $name differs; publish a new version instead of overwriting it." >&2
      exit 1
    }
  fi
done
for path in "$asset_dir/"*; do
  if ! printf '%s\n' "$existing" | grep -Fxq "${path##*/}"; then
    gh release upload "$tag" --repo "$repository" "$path"
  fi
done
