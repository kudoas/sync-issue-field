package main

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
	return client.Query(context.Background(), &i, variables)
}

func (i *IssueID) GetIssueID() githubv4.ID {
	return i.Repository.Issue.ID
}

func (i *IssueID) GetParentIssueID() githubv4.ID {
	return i.Repository.Issue.TrackedInIssues.Nodes[0].ID
}

type ParentIssue struct {
	Node struct {
		Issue struct {
			Assignees struct {
				Nodes []struct {
					ID githubv4.ID
				}
			} `graphql:"assignees(first: 5)"`
			Labels struct {
				Nodes []struct {
					ID githubv4.ID
				}
			} `graphql:"labels(first: 20)"`
			Milestone struct {
				ID githubv4.ID
			} `graphql:"milestone"`
			ProjectItems struct {
				Nodes []struct {
					FieldValues struct {
						Nodes []struct {
							OnProjectV2ItemFieldSingleSelectValue struct {
								ID       githubv4.ID
								OptionID githubv4.ID
							} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
							OnProjectV2ItemFieldDateValue struct {
								ID   githubv4.ID
								Date githubv4.String
							} `graphql:"... on ProjectV2ItemFieldDateValue"`
						}
					} `graphql:"fieldValues(last: 10)"`
				}
			} `graphql:"projectItems(first: 5)"`
		} `graphql:"... on Issue"`
	} `graphql:"node(id: $issueID)"`
}

func NewParentIssue() *ParentIssue {
	return &ParentIssue{}
}

func (p *ParentIssue) Query(client *githubv4.Client, ctx context.Context, variables map[string]interface{}) error {
	return client.Query(context.Background(), &p, variables)
}

func (p *ParentIssue) GetAssigneeIDs() *[]githubv4.ID {
	return extractIDs(p.Node.Issue.Assignees.Nodes)
}

func (p *ParentIssue) GetLabelIDs() *[]githubv4.ID {
	return extractIDs(p.Node.Issue.Labels.Nodes)
}

func (p *ParentIssue) GetMilestoneID() *githubv4.ID {
	return &p.Node.Issue.Milestone.ID
}

func extractIDs(nodes []struct {
	ID githubv4.ID
}) *[]githubv4.ID {
	ids := make([]githubv4.ID, len(nodes))
	for i := range nodes {
		ids[i] = nodes[i].ID
	}
	return &ids
}
