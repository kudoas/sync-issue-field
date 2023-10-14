# gis (GitHub Issue Syncer)

Tool that automatically synchronizes child and parent issue information

# Requirement

- Go version 1.21

# Setup

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
INPUT_GITHUB_TOKEN=<YOUR GITHUB TOKEN>
INPUT_GITHUB_REPO=<TARGET REPOSITORY NAME>
INPUT_GITHUB_ISSUE=<TARGET CHILD ISSUE NUMBER>
```

# Usage

run go command

```sh
go run .
```
