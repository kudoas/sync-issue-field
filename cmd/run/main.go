package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/kudoas/sync-issue-field/infra"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

var (
	token                  = os.Getenv("INPUT_TOKEN")
	repository             = strings.Split(os.Getenv("INPUT_REPOSITORY"), "/")
	owner, repository_name = repository[0], repository[1]
	issue, _               = strconv.Atoi(os.Getenv("INPUT_ISSUE"))
)

func main() {
	ctx := context.Background()
	client := getGithubClient(token, ctx)
	g := infra.NewGithubClient(
		infra.WithClient(client),
		infra.WithContext(ctx),
	)

	trackedIssueNodeIDs := g.GetTrackedIssueNodeIDs(repository_name, owner, issue)
	targetIssueNodeID := g.GetIssueNodeID(repository_name, owner, issue)
	parentIssueFields := g.GetIssueFields(&trackedIssueNodeIDs[0])

	if err := g.MutateIssue(
		githubv4.UpdateIssueInput{
			ID:          targetIssueNodeID,
			AssigneeIDs: &parentIssueFields.AssigneeIDs,
			LabelIDs:    &parentIssueFields.LabelIDs,
			MilestoneID: parentIssueFields.MilestoneID,
		},
	); err != nil {
		log.Fatalf("failed to update issue: %v", err)
	}

	if len(parentIssueFields.ProjectIDs) == 0 {
		os.Exit(0)
	}
	if err := g.MutateProject(githubv4.AddProjectV2ItemByIdInput{
		ProjectID: parentIssueFields.ProjectIDs[0],
		ContentID: targetIssueNodeID,
	}); err != nil {
		log.Fatalf("failed to add project item: %v", err)
	}
}

func getGithubClient(token string, ctx context.Context) *githubv4.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	client := oauth2.NewClient(ctx, src)
	return githubv4.NewClient(client)
}
