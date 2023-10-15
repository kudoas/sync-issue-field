package query

import (
	"context"
	"fmt"
	"os"

	"github.com/shurcooL/githubv4"
)

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
	} `graphql:"node(id: $issueID)"`
}

func NewParentIssue() *ParentIssue {
	return &ParentIssue{}
}

func (p *ParentIssue) Query(client *githubv4.Client, ctx context.Context, variables map[string]interface{}) error {
	return client.Query(ctx, &p, variables)
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

func (p *ParentIssue) GetProjectID() githubv4.ID {
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
