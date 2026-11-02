# Sync Issue Field

[![release-please](https://github.com/kudoas/sync-issue-field/actions/workflows/release.yml/badge.svg)](https://github.com/kudoas/sync-issue-field/actions/workflows/release.yml)
[![CI](https://github.com/kudoas/sync-issue-field/actions/workflows/ci.yml/badge.svg)](https://github.com/kudoas/sync-issue-field/actions/workflows/ci.yml)

This GitHub Action is used to synchronize specific fields (e.g., Assigned, Labels, Milestone, etc.) from a parent Issue to a child Issue.

## Usage

This action downloads a prebuilt binary from GitHub Releases, verifies its SHA-256 checksum and source commit, and runs it. Consumers do not need Go or a build step.

Use a full published version tag (`vX.Y.Z`) or the full 40-character commit SHA of that release. Branches such as `main`, moving aliases such as `v0`, and unreleased commits are not supported. The launcher never falls back to building from source or downloading an unrelated latest release.

Linux x64 and arm64 runners are supported. Self-hosted runners need Bash, Git, curl, sha256sum and access to the action repository and its public release assets on GitHub.com. Windows, macOS and GitHub Enterprise Server are not supported by this distribution mechanism.

### Workflow Example

Here's an example workflow that runs the action when an issue is opened, including the necessary permissions.

```yaml
name: Sync Issue Fields on Open

on:
  issues:
    types: [opened, edited, labeled, unlabeled, assigned]

permissions:
  issues: write
  # repository-projects: write # Uncomment if syncing classic projects
  # contents: read # Needed to checkout code

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - name: Use Sync Issue Field Action
        # Use the action from the kudoas/sync-issue-field repository.
        # Specify a published version tag or its full release commit SHA.
        uses: kudoas/sync-issue-field@<tag_or_commit_hash>
        with:
          # GitHub Token is required.
          token: ${{ secrets.GITHUB_TOKEN }}
          # Specify repository and issue if needed.
          # repository: 'owner/repo'
          # issue: ${{ github.event.issue.number }}
```

Replace `<tag_or_commit_hash>` with a release tag or its full commit SHA. The release must contain the assets listed below. Releases created before this distribution workflow was introduced do not contain those assets; continue using the old action ref until the first release containing this change has finished publishing.

For a SHA pin, the launcher resolves a version tag pointing to that exact commit. For both tag and SHA pins, `source-commit.txt` must match the resolved commit, and the selected binary and source record must match `SHA256SUMS`. Checksums detect transfer corruption and inconsistent assets; they are not independent signatures of the publisher.

### Inputs

The input parameters defined in `action.yml` are as follows:

| Name       | Description                                                                  | Default                     | Required |
|------------|------------------------------------------------------------------------------|-----------------------------|----------|
| `repository` | Repository name with owner (e.g., `kudoas/sync-issue-field`). Defaults to the current repository. | `${{ github.repository }}`  | `false`  |
| `issue`      | The number that identifies the child issue to synchronize with the parent issue. Defaults to the current issue. | `${{ github.event.issue.number }}` | `false`  |
| `token`      | GitHub Token.                                                               |                             | `true`   |

The `token` is required for the action to interact with the GitHub API.

## Developer Information

### How to Build

To build the Go binary locally, run the following command:

```bash
go build -v -o app ./cmd/run
```

To build the release assets locally:

```bash
bash scripts/build-release.sh ./release
RELEASE_ASSET_DIR="$PWD/release" go test ./tests -count=1 -v
```

### Release Process

The `release.yml` workflow runs release-please on pushes to `main`. When it creates a release, the same workflow calls `build.yml` with the release tag and source SHA. This avoids relying on a new workflow event from a tag or release created with `GITHUB_TOKEN` ([release-please documentation](https://github.com/googleapis/release-please-action#other-actions-on-release-please-prs)).

The publishing job checks out the release tag, verifies the source SHA, runs Go tests, builds static Linux amd64/arm64 binaries and tests the packaged assets and launcher. It then attaches these files to the existing GitHub Release:

| Asset | Purpose |
|-------|---------|
| `sync-issue-field_linux_amd64` | Linux x64 executable |
| `sync-issue-field_linux_arm64` | Linux arm64 executable |
| `source-commit.txt` | Full source commit SHA |
| `SHA256SUMS` | Checksums for the three files above |

No binary commits or new signing secrets are required. PR CI builds the same asset format, tests corruption/missing assets/source mismatches, runs the real packaged binary on Linux with invalid input before any GitHub API call, and invokes the composite Action with offline release fixtures. Fixture tests do not publish a release or change Issues.

A release is ready for consumers only after the publishing job succeeds. For a failed or partial upload, run `release.yml` manually with the existing release tag. Only tags that include these scripts can be published this way. Existing identical assets are kept and missing assets are uploaded; differing existing assets cause a failure. Never replace a published binary in place: fix the source and release a new version.

To verify the first real release after merging, check that the four assets are present, run the Action from another repository using both the version tag and full release commit SHA, and confirm that no Go build runs on the consumer. This live publication/consumption check is separate from the offline PR tests. To roll back a consumer, pin a previous working release or legacy action commit; reverting source alone does not change already published assets.


## License

Refer to the [LICENSE](LICENSE) file.
