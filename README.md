# Sync Issue Field

[![release-please](https://github.com/kudoas/sync-issue-field/actions/workflows/release.yml/badge.svg)](https://github.com/kudoas/sync-issue-field/actions/workflows/release.yml)
[![Build and Commit Binary](https://github.com/kudoas/sync-issue-field/actions/workflows/build-and-commit-binary.yml/badge.svg)](https://github.com/kudoas/sync-issue-field/actions/workflows/build-and-commit-binary.yml)

This GitHub Action is used to synchronize specific fields (e.g., Assigned, Labels, Milestone, etc.) from a parent Issue to a child Issue.

## Usage

This action is used within GitHub Actions workflows. Since this configuration includes the built binary in the repository, workflows using this action must specify the tag or commit hash where the binary has been committed.

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
        # Specify the release tag (e.g., v1.0.0) or commit hash.
        uses: kudoas/sync-issue-field@<tag_or_commit_hash>
        with:
          # GitHub Token is required.
          token: ${{ secrets.GITHUB_TOKEN }}
          # Specify repository and issue if needed.
          # repository: 'owner/repo'
          # issue: ${{ github.event.issue.number }}
```

In `uses: kudoas/sync-issue-field@<tag_or_commit_hash>`, replace `<tag_or_commit_hash>` with the **release tag name (e.g., `v1.0.0`) or commit hash of the `kudoas/sync-issue-field` repository** that contains the built binary in the `dist/` directory.

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

It is recommended to move the generated `app` binary to the `dist/` directory:

```bash
mkdir -p dist
mv app dist/
```

### Release Process

When the `release-please-action` workflow creates a tag, the `Build and Commit Binary` workflow will subsequently run to build and commit the binary to the repository.


## License

Refer to the [LICENSE](LICENSE) file.
