package main

import (
	"context"
	"os"
	"strconv"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

var (
	token    = os.Getenv("GITHUB_TOKEN")
	owner    = os.Getenv("GITHUB_OWNER")
	name     = os.Getenv("GITHUB_REPO")
	issue, _ = strconv.Atoi(os.Getenv("GITHUB_ISSUE"))
)

func main() {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)
	var getParentIDQuery struct {
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

	variable := map[string]interface{}{
		"repositoryOwner": githubv4.String(owner),
		"repositoryName":  githubv4.String(name),
		"issueNumber":     githubv4.Int(issue),
	}
	if err := client.Query(context.Background(), &getParentIDQuery, variable); err != nil {
		panic(err)
	}

	var getParentIssueQuery struct {
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
	if err := client.Query(context.Background(), &getParentIssueQuery, map[string]interface{}{
		"issueID": githubv4.ID(getParentIDQuery.Repository.Issue.TrackedInIssues.Nodes[0].ID),
	}); err != nil {
		panic(err)
	}

	var mutation struct {
		UpdateIssue struct {
			Issue struct {
				ID githubv4.ID
			}
		} `graphql:"updateIssue(input: $input)"`
	}

	input := githubv4.UpdateIssueInput{
		ID:          getParentIDQuery.Repository.Issue.ID,
		AssigneeIDs: extractIDs(getParentIssueQuery.Node.Issue.Assignees.Nodes),
		LabelIDs:    extractIDs(getParentIssueQuery.Node.Issue.Labels.Nodes),
		MilestoneID: &getParentIssueQuery.Node.Issue.Milestone.ID,
	}
	if err := client.Mutate(context.Background(), &mutation, input, nil); err != nil {
		panic(err)
	}
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
