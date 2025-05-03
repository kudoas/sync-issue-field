package github

import (
	"github.com/shurcooL/githubv4"
)

type GitHubClient interface {
	GetTrackedIssueNodeIDs(q *QueryRequest) ([]githubv4.ID, error)
	GetIssueNodeID(q *QueryRequest) (githubv4.ID, error)
	GetIssueFields(nodeID githubv4.ID) (*IssueFields, error)
	MutateIssue(input githubv4.UpdateIssueInput) error
	MutateProject(input githubv4.AddProjectV2ItemByIdInput) error
}
