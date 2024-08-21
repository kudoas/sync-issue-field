package github

import (
	"context"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GithubClient interface {
	GetTrackedIssueNodeIDs(q *QueryRequest) ([]githubv4.ID, error)
	GetIssueNodeID(q *QueryRequest) (githubv4.ID, error)
	GetIssueFields(issueNodeID githubv4.ID) (*IssueFields, error)
	MutateIssue(input githubv4.UpdateIssueInput) error
	MutateProject(input githubv4.AddProjectV2ItemByIdInput) error
}

type githubClient struct {
	client  *githubv4.Client
	context context.Context
}

func NewGithubClient(ctx context.Context, token string) GithubClient {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, src)

	return &githubClient{
		client:  githubv4.NewClient(httpClient),
		context: ctx,
	}
}
