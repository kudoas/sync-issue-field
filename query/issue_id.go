package query

import (
	"context"

	"github.com/shurcooL/githubv4"
)

type IssueID struct {
	Repository struct {
		Issue struct {
			ID              githubv4.ID
			TrackedInIssues struct {
				Nodes []struct {
					ID githubv4.ID
				}
			} `graphql:"trackedInIssues(first: 5)"`
		} `graphql:"issue(number: $issueNumber)"`
	} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
}

func NewIssueID() *IssueID {
	return &IssueID{}
}

func (i *IssueID) Query(client *githubv4.Client, ctx context.Context, variables map[string]interface{}) error {
	return client.Query(ctx, &i, variables)
}

func (i *IssueID) GetIssueID() githubv4.ID {
	return i.Repository.Issue.ID
}

func (i *IssueID) GetParentIssueID() githubv4.ID {
	return i.Repository.Issue.TrackedInIssues.Nodes[0].ID
}
