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

func ProvideEnv() (*Env, error) {
	token := os.Getenv("GITHUB_TOKEN")
	repository := os.Getenv("GITHUB_REPOSITORY")
	issueStr := os.Getenv("GITHUB_ISSUE")

	fmt.Println("GITHUB_REPOSITORY", repository)
	fmt.Println("GITHUB_ISSUE", issueStr)

	r := strings.Split(repository, "/")
	if len(r) != 2 || r[0] == "" || r[1] == "" {
		return nil, &EnvInvalidError{"Repository format is incorrect"}
	}
	var (
		repoOwner = r[0]
		repoName  = r[1]
	)

	if token == "" {
		return nil, &EnvInvalidError{"Token is not set"}
	}

	issue, err := strconv.Atoi(issueStr)
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
