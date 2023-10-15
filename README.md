# gis (GitHub Issue Syncer)

This action automatically synchronizes child and parent issue information

# Usage

<!-- start usage -->

```yaml
with:
  # Repository name with owner. For example, kudoas/gis
  # Default: ${{ github.repository }}
  repository: ""

  # The number that identifies the issue. Specify the child issue you wish to synchronize with the parent issue
  # Default: ${{ github.event.issue.number }}
  issue: ""

  # Personal access token (PAT) used to fetch the repository.
  token: ""
```

<!-- end usage -->

# Example

```yml
name: Sync issue
on:
  issues:

# If use GITHUB_TOKEN, this permission is required.
permissions:
  issues: write
  repository-projects: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@main
      - name: Sync issue
        uses: kudoas/gis@main
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
```

# Contribution

## Requirement

- Go version 1.21

## Setup

get source code

```sh
git clone git@github.com:kudoas/gis.git
cd /path/to/repository
```

copy .env from .env.sample

```sh
cp .env.sample .env
```

Get values from GitHub and enter them in .env

```.env
INPUT_TOKEN=<YOUR GITHUB TOKEN>
INPUT_REPOSITORY=<TARGET REPOSITORY NAME eg. <owner>/<repository>>
INPUT_ISSUE=<TARGET CHILD ISSUE NUMBER>
```
