package query

import (
	"context"
	"fmt"
	"os"

	"github.com/shurcooL/githubv4"
)

type IssueNodeID struct {
	Repository struct {
		Issue struct {
			ID              githubv4.ID
			TrackedInIssues struct {
				Nodes []struct {
					ID githubv4.ID
				}
			} `graphql:"trackedInIssues(first: 5)"`
		} `graphql:"issue(number: $IssueID)"`
	} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
}

func NewIssueNodeID() *IssueNodeID {
	return &IssueNodeID{}
}

func (i *IssueNodeID) Query(client *githubv4.Client, ctx context.Context, variables map[string]interface{}) error {
	return client.Query(ctx, &i, variables)
}

func (i *IssueNodeID) GetIssueNodeID() githubv4.ID {
	return i.Repository.Issue.ID
}

func (i *IssueNodeID) GetParentIssueNodeID() githubv4.ID {
	if len(i.Repository.Issue.TrackedInIssues.Nodes) == 0 {
		os.Exit(0)
	}
	return i.Repository.Issue.TrackedInIssues.Nodes[0].ID
}

type IssueItem struct {
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
					Project struct {
						ID githubv4.ID
					}
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
	} `graphql:"node(id: $issueNodeID)"`
}

func NewIssueItem() *IssueItem {
	return &IssueItem{}
}

func (p *IssueItem) Query(client *githubv4.Client, ctx context.Context, variables map[string]interface{}) error {
	return client.Query(ctx, &p, variables)
}

func (p *IssueItem) GetAssigneeIDs() *[]githubv4.ID {
	return extractIDs(p.Node.Issue.Assignees.Nodes)
}

func (p *IssueItem) GetLabelIDs() *[]githubv4.ID {
	return extractIDs(p.Node.Issue.Labels.Nodes)
}

func (p *IssueItem) GetMilestoneID() *githubv4.ID {
	return &p.Node.Issue.Milestone.ID
}

func (p *IssueItem) GetProjectID() githubv4.ID {
	println(fmt.Sprintf("%+v", p.Node.Issue.ProjectItems.Nodes))
	if len(p.Node.Issue.ProjectItems.Nodes) == 0 {
		os.Exit(0)
	}
	return p.Node.Issue.ProjectItems.Nodes[0].Project.ID
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
