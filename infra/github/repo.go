package github

import (
	"github.com/shurcooL/githubv4"
)

func (g *githubClient) GetTrackedIssueNodeIDs(q *QueryRequest) ([]githubv4.ID, error) {
	var query struct {
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

	variables := map[string]interface{}{
		"repositoryOwner": githubv4.String(q.RepositoryOwner),
		"repositoryName":  githubv4.String(q.RepositoryName),
		"issueNumber":     githubv4.Int(q.IssueNumber),
	}

	err := g.client.Query(g.context, &query, variables)
	if err != nil {
		return nil, err
	}

	ids := make([]githubv4.ID, 0, len(query.Repository.Issue.TrackedInIssues.Nodes))
	for _, node := range query.Repository.Issue.TrackedInIssues.Nodes {
		ids = append(ids, node.ID)
	}
	return ids, nil
}

func (g *githubClient) GetIssueNodeID(q *QueryRequest) (githubv4.ID, error) {
	var query struct {
		Repository struct {
			Issue struct {
				ID githubv4.ID
			} `graphql:"issue(number: $issueNumber)"`
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}

	variables := map[string]interface{}{
		"repositoryOwner": githubv4.String(q.RepositoryOwner),
		"repositoryName":  githubv4.String(q.RepositoryName),
		"issueNumber":     githubv4.Int(q.IssueNumber),
	}

	err := g.client.Query(g.context, &query, variables)
	if err != nil {
		return nil, err
	}

	return query.Repository.Issue.ID, nil
}

func (g *githubClient) GetIssueFields(nodeID githubv4.ID) (*IssueFields, error) {
	var query struct {
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
		} `graphql:"node(id: $nodeID)"`
	}
	if err := g.client.Query(g.context, &query, map[string]interface{}{
		"nodeID": nodeID,
	}); err != nil {
		return nil, err
	}

	assigneeIDs := make([]githubv4.ID, 0, len(query.Node.Issue.Assignees.Nodes))
	for _, node := range query.Node.Issue.Assignees.Nodes {
		assigneeIDs = append(assigneeIDs, node.ID)
	}

	labelIDs := make([]githubv4.ID, 0, len(query.Node.Issue.Labels.Nodes))
	for _, node := range query.Node.Issue.Labels.Nodes {
		labelIDs = append(labelIDs, node.ID)
	}

	projectIDs := make([]githubv4.ID, 0, len(query.Node.Issue.ProjectItems.Nodes))
	for _, node := range query.Node.Issue.ProjectItems.Nodes {
		projectIDs = append(projectIDs, node.Project.ID)
	}

	return &IssueFields{
		AssigneeIDs: assigneeIDs,
		LabelIDs:    labelIDs,
		MilestoneID: &query.Node.Issue.Milestone.ID,
		ProjectIDs:  projectIDs,
	}, nil
}

func (g *githubClient) MutateIssue(input githubv4.UpdateIssueInput) error {
	var mutation struct {
		UpdateIssue struct {
			Issue struct {
				ID githubv4.ID
			}
		} `graphql:"updateIssue(input: $input)"`
	}

	err := g.client.Mutate(g.context, &mutation, input, nil)
	if err != nil {
		return err
	}

	return nil
}

func (g *githubClient) MutateProject(input githubv4.AddProjectV2ItemByIdInput) error {
	var mutation struct {
		AddProjectV2ItemById struct {
			Item struct {
				ID githubv4.ID `graphql:"id"`
			} `graphql:"item"`
		} `graphql:"addProjectV2ItemById(input: $input)"`
	}

	if err := g.client.Mutate(g.context, &mutation, input, nil); err != nil {
		return err
	}

	return nil
}
