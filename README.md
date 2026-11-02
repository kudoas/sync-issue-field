# Sync Issue Field

[![release-please](https://github.com/kudoas/sync-issue-field/actions/workflows/release.yml/badge.svg)](https://github.com/kudoas/sync-issue-field/actions/workflows/release.yml)
[![CI](https://github.com/kudoas/sync-issue-field/actions/workflows/ci.yml/badge.svg)](https://github.com/kudoas/sync-issue-field/actions/workflows/ci.yml)

This GitHub Action is used to synchronize specific fields (e.g., Assigned, Labels, Milestone, etc.) from a parent Issue to a child Issue.

The parent is the issue's explicit GitHub Sub-issue parent. For compatibility, when no explicit parent exists, a single issue tracking the child in a Markdown checklist is also accepted. An explicit parent takes priority over checklist references. Multiple checklist tracking issues without an explicit parent are ambiguous and cause an error rather than selecting one arbitrarily. If neither relation exists, synchronization is skipped.

The opt-in `Verify parent synchronization` workflow checks both parent relations using isolated test Issues and reads back assignees, labels and milestone. It closes those Issues and deletes its temporary labels/milestones afterward. It requires `issues: write`; normal PR CI remains read-only and skips live tests.

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

Development builds require Go 1.26 or newer; `go.mod` recommends Go 1.27.3. The Docker builder uses the same recommended toolchain. Users of the prebuilt GitHub Action do not need Go installed.

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

The `release.yml` workflow runs release-please on pushes to `main`, using `release-please-config.json` and `.release-please-manifest.json`. Release PRs update the changelog and version manifest. When a release PR is merged, release-please creates a **draft** release and the same workflow calls `build.yml` with its version and source SHA. This avoids relying on a new workflow event from a tag or release created with `GITHUB_TOKEN` ([release-please documentation](https://github.com/googleapis/release-please-action#other-actions-on-release-please-prs)).

The publishing job checks out the exact source SHA, since a draft may not have a Git tag yet. It verifies the draft's target commit (and any existing tag), runs Go tests, builds static Linux amd64/arm64 binaries and tests the packaged assets and launcher. It attaches these files before publishing the release:

| Asset | Purpose |
|-------|---------|
| `sync-issue-field_linux_amd64` | Linux x64 executable |
| `sync-issue-field_linux_arm64` | Linux arm64 executable |
| `source-commit.txt` | Full source commit SHA |
| `SHA256SUMS` | Checksums for the three files above |

No binary commits or new signing secrets are required. PR CI builds the same asset format, tests corruption/missing assets/source mismatches, runs the real packaged binary on Linux with invalid input before any GitHub API call, and invokes the composite Action with offline release fixtures. It also runs `go test ./...` and `go vet ./...` with read-only repository permissions. The GitHub client's HTTP fixture tests exercise query variables, response decoding, mutation inputs, GraphQL errors and HTTP 401 responses without a PAT or shared test Issue. These tests do not publish a release or change Issues; they do not establish compatibility with GitHub's live schema or authentication.

A release is ready for consumers only after every asset is uploaded and the workflow publishes the draft. A suffixed version such as `v0.2.0-rc.1` is marked as a prerelease and is not made latest. `v0.2.0` is a normal release and becomes latest. If an upload fails, the release stays a draft. Retry `release.yml` with the same version and full source SHA recorded in that release. Only source commits containing these scripts can be published this way. Existing identical assets are kept and missing assets are uploaded; differing existing assets cause a failure. Never replace a published binary in place: fix the source and release a new version.

The first rollout is deliberately staged:

1. Request `0.2.0-rc.1` through a conventional commit footer `Release-As: 0.2.0-rc.1`, then review and merge the release PR generated by release-please.
2. Wait for the prerelease and all four assets to be published. From an existing consumer repository, verify both the RC tag and its full commit SHA, confirming the selected architecture, checksum/source verification and execution without Go builds. Record the source SHA and consumer Actions run URL. Distribution verification alone does not prove all Issue synchronization behavior.
3. After verification of the RC or its corrected candidate succeeds and the maintainer approves the stable release, request `0.2.0` with `Release-As: 0.2.0` and review the next release-please PR. This creates a separate release commit and repeats the same draft/build/test/upload/publish process.

If an RC needs a fix, verify the corrected candidate before releasing it. Publish another RC if further prerelease validation is needed, or request the approved stable version after successful checks. Never change the original RC's tag or assets.

During 0.x, compatibility is maintained within a minor version; crossing minor versions may include breaking changes. Full version tags and their assets are never moved or overwritten. Neither `v0` nor `v1` aliases are published in this rollout. The version manifest starts from the actual latest published version, `0.1.3`; the old unpublished `0.2.0` changelog entry is recorded under Unreleased instead.

Release PRs created with `GITHUB_TOKEN` do not automatically trigger pull-request workflows. Run the CI workflow manually against the generated release branch (`gh workflow run ci.yml --ref <release-branch>`) and check the run's exact commit before merging it. The release-please workflow prepares a new release PR only when it did not create a draft release in that run, so it does not prepare another version while the current draft is still being published.

To verify the first real release after merging, check that the four assets are present, run the Action from another repository using both the version tag and full release commit SHA, and confirm that no Go build runs on the consumer. This live publication/consumption check is separate from the offline PR tests. To roll back a consumer, pin a previous working release or legacy action commit; reverting source alone does not change already published assets.


## License

Refer to the [LICENSE](LICENSE) file.

Repository development and release workflows use Node 24 based Actions and require Actions Runner 2.327.1 or newer. Authenticated Git operations from Docker container actions require Runner 2.329.0 or newer with checkout v7; these workflows do not perform such operations. This requirement concerns the repository workflows; the published composite Action uses Bash.
