package infra

import (
	"context"
	"log"

	"github.com/shurcooL/githubv4"
)

type GithubClient struct {
	client *githubv4.Client
	ctx    context.Context
}

type Option func(*GithubClient)

func NewGithubClient(opt ...Option) *GithubClient {
	g := &GithubClient{}
	for _, o := range opt {
		o(g)
	}
	return g
}

func WithClient(client *githubv4.Client) func(*GithubClient) {
	return func(g *GithubClient) {
		g.client = client
	}
}

func WithContext(ctx context.Context) func(*GithubClient) {
	return func(g *GithubClient) {
		g.ctx = ctx
	}
}

func (g *GithubClient) GetTrackedIssueNodeIDs(repoName string, ownerName string, issueNumber int) []githubv4.ID {
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

	if err := g.client.Query(g.ctx, &query, map[string]interface{}{
		"repositoryOwner": githubv4.String(ownerName),
		"repositoryName":  githubv4.String(repoName),
		"issueNumber":     githubv4.Int(issueNumber),
	}); err != nil {
		log.Fatalf("failed to get tracked issue NodeID: %v", err)
	}

	var ids []githubv4.ID
	for _, node := range query.Repository.Issue.TrackedInIssues.Nodes {
		ids = append(ids, node.ID)
	}

	return ids
}

// TODO: GetIssueNodeID and GetIssueFields can be combined into one function
func (g *GithubClient) GetIssueNodeID(repoName string, ownerName string, issueNumber int) *githubv4.ID {
	var query struct {
		Repository struct {
			Issue struct {
				ID githubv4.ID
			} `graphql:"issue(number: $issueNumber)"`
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}

	if err := g.client.Query(g.ctx, &query, map[string]interface{}{
		"repositoryOwner": githubv4.String(ownerName),
		"repositoryName":  githubv4.String(repoName),
		"issueNumber":     githubv4.Int(issueNumber),
	}); err != nil {
		log.Fatalf("failed to get issue NodeID: %v", err)
	}

	return &query.Repository.Issue.ID
}

type IssueFields struct {
	AssigneeIDs []githubv4.ID
	LabelIDs    []githubv4.ID
	MilestoneID *githubv4.ID
	ProjectIDs  []githubv4.ID
}

func (g *GithubClient) GetIssueFields(nodeID *githubv4.ID) *IssueFields {
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

	if err := g.client.Query(g.ctx, &query, map[string]interface{}{
		"nodeID": *nodeID,
	}); err != nil {
		log.Fatalf("failed to get issue fields: %v", err)
	}

	var assigneeIDs []githubv4.ID
	for _, node := range query.Node.Issue.Assignees.Nodes {
		assigneeIDs = append(assigneeIDs, node.ID)
	}

	var labelIDs []githubv4.ID
	for _, node := range query.Node.Issue.Labels.Nodes {
		labelIDs = append(labelIDs, node.ID)
	}

	var projectIDs []githubv4.ID
	for _, node := range query.Node.Issue.ProjectItems.Nodes {
		projectIDs = append(projectIDs, node.Project.ID)
	}

	return &IssueFields{
		AssigneeIDs: assigneeIDs,
		LabelIDs:    labelIDs,
		MilestoneID: &query.Node.Issue.Milestone.ID,
		ProjectIDs:  projectIDs,
	}
}

func (g *GithubClient) MutateIssue(input githubv4.UpdateIssueInput) error {
	var mutation struct {
		UpdateIssue struct {
			Issue struct {
				ID githubv4.ID
			}
		} `graphql:"updateIssue(input: $input)"`
	}

	if err := g.client.Mutate(g.ctx, &mutation, input, nil); err != nil {
		return err
	}

	return nil
}

func (g *GithubClient) MutateProject(input githubv4.AddProjectV2ItemByIdInput) error {
	var mutation struct {
		AddProjectV2ItemById struct {
			Item struct {
				ID githubv4.ID `graphql:"id"`
			} `graphql:"item"`
		} `graphql:"addProjectV2ItemById(input: $input)"`
	}

	if err := g.client.Mutate(g.ctx, &mutation, input, nil); err != nil {
		return err
	}

	return nil
}
