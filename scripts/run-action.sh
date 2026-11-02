#!/usr/bin/env bash
set -euo pipefail

fail() { echo "::error::$*" >&2; exit 1; }

[[ "${RUNNER_OS:-}" == Linux ]] || fail 'This action supports Linux runners only.'
case "${RUNNER_ARCH:-}" in
  X64) arch=amd64 ;;
  ARM64) arch=arm64 ;;
  *) fail "Unsupported runner architecture: ${RUNNER_ARCH:-unset}" ;;
esac

repository="${ACTION_REPOSITORY:?ACTION_REPOSITORY is required}"
ref="${ACTION_REF:-}"
# Local `uses: ./` has no action_ref; use that checkout, not the caller's event SHA.
if [[ -z "$ref" ]]; then
  ref="$(git -C "${ACTION_PATH:?ACTION_PATH is required for a local action}" rev-parse HEAD)"
fi
[[ "$repository" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] || fail 'Invalid action repository.'
version_pattern='v[0-9]+[.][0-9]+[.][0-9]+(-[0-9A-Za-z.-]+)?'

# Resolve only version tags; never silently run the latest release for a branch.
# Peeling annotated tags also allows a full release commit SHA in `uses:`.
if [[ "$ref" =~ ^${version_pattern}$ ]]; then
  refs="$(git ls-remote --tags "https://github.com/$repository.git" "refs/tags/$ref" "refs/tags/$ref^{}")"
elif [[ "$ref" =~ ^[0-9a-f]{40}$ ]]; then
  refs="$(git ls-remote --tags "https://github.com/$repository.git")"
else
  fail 'Use a full release version tag (vX.Y.Z) or its 40-character commit SHA; branches are not supported.'
fi

selection="$(printf '%s\n' "$refs" | awk -v ref="$ref" -v pattern="^refs/tags/$version_pattern$" '
  $2 ~ /\^\{\}$/ { name=$2; sub(/\^\{\}$/, "", name); peeled[name]=$1; next }
  $2 ~ pattern { commits[$2]=$1 }
  END {
    for (name in commits) {
      sha=(name in peeled) ? peeled[name] : commits[name]
      tag=name; sub(/^refs\/tags\//, "", tag)
      if (tag == ref || sha == ref) print tag, sha
    }
  }' | sort)"
[[ -n "$selection" ]] || fail 'No release version tag matches this action ref.'
# Several release tags can point to the same commit; choose one deterministically.
read -r tag source_commit <<< "${selection%%$'\n'*}"
[[ "$source_commit" =~ ^[0-9a-f]{40}$ ]] || fail 'Invalid release source commit.'

work_dir="$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/sync-issue-field.XXXXXX")"
trap 'rm -rf "$work_dir"' EXIT
base_url="https://github.com/$repository/releases/download/$tag"
binary="sync-issue-field_linux_$arch"
for name in SHA256SUMS source-commit.txt "$binary"; do
  curl --fail --silent --show-error --location --retry 3 \
    --connect-timeout 15 --max-time 120 \
    "$base_url/$name" --output "$work_dir/$name"
done

# Select exact filenames, require one checksum per file, and never evaluate names
# from the remote checksum file as shell commands or filesystem paths.
for name in source-commit.txt "$binary"; do
  checksum="$(awk -v name="$name" '$2 == name {print $1}' "$work_dir/SHA256SUMS")"
  [[ "$checksum" =~ ^[0-9a-f]{64}$ ]] || fail "Missing, duplicate or invalid checksum for $name."
  printf '%s  %s\n' "$checksum" "$work_dir/$name" | sha256sum -c - >/dev/null \
    || fail "Checksum mismatch for $name."
done
[[ "$(cat "$work_dir/source-commit.txt")" == "$source_commit" ]] \
  || fail 'The release binary does not match the action source commit.'

chmod +x "$work_dir/$binary"
"$work_dir/$binary"
