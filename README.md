# Sync Issue

This action automatically synchronizes child and parent issue information.
If you create child issue from parent issue, the parent information is copied to the child.

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

# Example Workflow

Create a [workflow](https://docs.github.com/en/actions/using-workflows) and save it as a `.yml` file in the `.github/workflows/` directory of your target repository.

```yml
name: Sync issue
on:
  issues:
    types: [opened]

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
        uses: kudoas/sync-issue@main
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
