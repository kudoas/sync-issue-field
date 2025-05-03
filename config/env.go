package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Env struct {
	token       string
	repoOwner   string
	repoName    string
	issueNumber int
}

var (
	token      = os.Getenv("GITHUB_TOKEN")
	repository = os.Getenv("GITHUB_REPOSITORY")
	issue      = os.Getenv("GITHUB_ISSUE")
)

func ProvideEnv() (*Env, error) {
	// print env
	fmt.Println("GITHUB_REPOSITORY", repository)
	fmt.Println("GITHUB_ISSUE", issue)

	token := os.Getenv("GITHUB_TOKEN")
	r := strings.Split(repository, "/")
	var (
		repoOwner = r[0]
		repoName  = r[1]
	)
	if (repoOwner == "" || repoName == "") || len(r) != 2 {
		return nil, &EnvInvalidError{"Repository format is incorrect"}
	}

	if token == "" {
		return nil, &EnvInvalidError{"Token is not set"}
	}

	issue, err := strconv.Atoi(issue)
	if err != nil {
		return nil, &EnvInvalidError{"Issue number can't parse to int"}
	}
	if issue == 0 {
		return nil, &EnvInvalidError{"Issue number is not set"}
	}

	return &Env{
		token:       token,
		repoOwner:   repoOwner,
		repoName:    repoName,
		issueNumber: issue,
	}, nil
}

func (e *Env) Token() string {
	return e.token
}

func (e *Env) RepoOwner() string {
	return e.repoOwner
}

func (e *Env) RepoName() string {
	return e.repoName
}

func (e *Env) IssueNumber() int {
	return e.issueNumber
}

type EnvInvalidError struct {
	message string
}

func (e *EnvInvalidError) Error() string {
	return fmt.Sprintf("environment variable invalid error: %v", e.message)
}

func (e *EnvInvalidError) Unwrap() error {
	return e
}
