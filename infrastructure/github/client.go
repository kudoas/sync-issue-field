package github

import (
	"context"
	"errors"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type githubClient struct {
	client  *githubv4.Client
	context context.Context
}

func NewGithubClient(ctx context.Context, token string) GitHubClient {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, src)

	return &githubClient{
		client:  githubv4.NewClient(httpClient),
		context: ctx,
	}
}

func (g *githubClient) GetParentIssueNodeID(q *QueryRequest) (githubv4.ID, error) {
	var query struct {
		Repository struct {
			Issue struct {
				Parent struct {
					ID githubv4.ID
				}
				TrackedInIssues struct {
					Nodes []struct {
						ID githubv4.ID
					}
				} `graphql:"trackedInIssues(first: 2)"`
			} `graphql:"issue(number: $issueNumber)"`
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}
	variables := map[string]interface{}{
		"repositoryOwner": githubv4.String(q.RepositoryOwner),
		"repositoryName":  githubv4.String(q.RepositoryName),
		"issueNumber":     githubv4.Int(q.IssueNumber),
	}
	if err := g.client.Query(g.context, &query, variables); err != nil {
		return nil, err
	}

	if id := query.Repository.Issue.Parent.ID; id != nil {
		return id, nil
	}
	// Preserve checklist compatibility only when the tracking parent is unique.
	// Fetching two nodes is sufficient to detect an ambiguous fallback.
	tracked := query.Repository.Issue.TrackedInIssues.Nodes
	switch len(tracked) {
	case 0:
		return nil, nil
	case 1:
		return tracked[0].ID, nil
	default:
		return nil, errors.New("multiple tracking issues found; set an explicit Sub-issue parent")
	}
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
	if err := g.client.Query(g.context, &query, variables); err != nil {
		return nil, err
	}

	return query.Repository.Issue.ID, nil
}

type labelPage struct {
	Nodes    []struct{ ID githubv4.ID }
	PageInfo struct {
		EndCursor   githubv4.String
		HasNextPage bool
	}
}

func (g *githubClient) GetIssueFields(nodeID githubv4.ID) (*IssueFields, error) {
	var query struct {
		Node struct {
			Issue struct {
				Assignees struct {
					Nodes []struct {
						ID githubv4.ID
					}
				} `graphql:"assignees(first: 10)"` // GitHub supports at most 10 assignees per issue.
				Labels    labelPage `graphql:"labels(first: 100)"`
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
	page := query.Node.Issue.Labels
	seenCursors := make(map[githubv4.String]bool)
	for {
		for _, node := range page.Nodes {
			labelIDs = append(labelIDs, node.ID)
		}
		if !page.PageInfo.HasNextPage {
			break
		}
		cursor := page.PageInfo.EndCursor
		if cursor == "" || seenCursors[cursor] {
			return nil, errors.New("labels pagination did not advance")
		}
		seenCursors[cursor] = true
		var next struct {
			Node struct {
				Issue struct {
					Labels labelPage `graphql:"labels(first: 100, after: $cursor)"`
				} `graphql:"... on Issue"`
			} `graphql:"node(id: $nodeID)"`
		}
		if err := g.client.Query(g.context, &next, map[string]interface{}{
			"nodeID": nodeID,
			"cursor": cursor,
		}); err != nil {
			return nil, err
		}
		page = next.Node.Issue.Labels
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

	if err := g.client.Mutate(g.context, &mutation, input, nil); err != nil {
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
