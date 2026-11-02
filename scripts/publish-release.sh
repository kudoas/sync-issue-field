#!/usr/bin/env bash
set -euo pipefail

tag="${1:?usage: publish-release.sh TAG ASSET_DIRECTORY}"
asset_dir="${2:?usage: publish-release.sh TAG ASSET_DIRECTORY}"
repository="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}"
[[ "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]] || {
  echo '::error::Use a full semantic version tag.' >&2
  exit 1
}
for name in SHA256SUMS source-commit.txt sync-issue-field_linux_amd64 sync-issue-field_linux_arm64; do
  test -f "$asset_dir/$name"
done
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

# Only expose the release after every asset has been uploaded successfully.
# 0.x stable releases are not GitHub prereleases; only suffixed versions are.
prerelease=false
latest=true
if [[ "$tag" == *-* ]]; then
  prerelease=true
  latest=false
fi
gh release edit "$tag" --repo "$repository" --draft=false \
  --prerelease="$prerelease" --latest="$latest"
