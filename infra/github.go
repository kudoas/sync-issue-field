package infra

import (
	"context"
	"log"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
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

type GitHubToken string

func WithClient(ctx context.Context, token string) func(*GithubClient) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client := githubv4.NewClient(oauth2.NewClient(ctx, src))

	return func(g *GithubClient) {
		g.client = client
	}
}

func WithContext(ctx context.Context) func(*GithubClient) {
	return func(g *GithubClient) {
		g.ctx = ctx
	}
}

var issueQuery struct {
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

type QueryRequest struct {
	RepositoryOwner string
	RepositoryName  string
	IssueNumber     int
}

func (g *GithubClient) GetTrackedIssueNodeIDs(q *QueryRequest) []githubv4.ID {
	if err := g.client.Query(g.ctx, &issueQuery, map[string]interface{}{
		"repositoryOwner": githubv4.String(q.RepositoryOwner),
		"repositoryName":  githubv4.String(q.RepositoryName),
		"issueNumber":     githubv4.Int(q.IssueNumber),
	}); err != nil {
		log.Fatalf("failed to get tracked issue NodeID: %v", err)
	}

	var ids []githubv4.ID

	for _, node := range issueQuery.Repository.Issue.TrackedInIssues.Nodes {
		ids = append(ids, node.ID)
	}

	return ids
}

func (g *GithubClient) GetIssueNodeID(q *QueryRequest) *githubv4.ID {
	// If the issue is already tracked, return the issue NodeID directly
	// TODO(bug): If QueryResult is not same, it should request to get issue NodeID
	if issueQuery.Repository.Issue.ID != "" {
		return &issueQuery.Repository.Issue.ID
	}

	if err := g.client.Query(g.ctx, &issueQuery, map[string]interface{}{
		"repositoryOwner": githubv4.String(q.RepositoryOwner),
		"repositoryName":  githubv4.String(q.RepositoryName),
		"issueNumber":     githubv4.Int(q.IssueNumber),
	}); err != nil {
		log.Fatalf("failed to get issue NodeID: %v", err)
	}

	return &issueQuery.Repository.Issue.ID
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
