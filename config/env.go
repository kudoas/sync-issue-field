package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Env struct {
	token         string
	repoOwner     string
	repoName      string
	issueNumber   int
	syncMilestone bool
	syncProjects  bool
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

	syncMilestone, err := optionalBool("SYNC_MILESTONE")
	if err != nil {
		return nil, err
	}
	syncProjects, err := optionalBool("SYNC_PROJECTS")
	if err != nil {
		return nil, err
	}

	return &Env{
		token:         token,
		repoOwner:     repoOwner,
		repoName:      repoName,
		issueNumber:   issue,
		syncMilestone: syncMilestone,
		syncProjects:  syncProjects,
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

func optionalBool(name string) (bool, error) {
	switch os.Getenv(name) {
	case "", "false":
		return false, nil
	case "true":
		return true, nil
	default:
		return false, &EnvInvalidError{name + " must be true or false"}
	}
}
func (e *Env) SyncMilestone() bool { return e.syncMilestone }
func (e *Env) SyncProjects() bool  { return e.syncProjects }
