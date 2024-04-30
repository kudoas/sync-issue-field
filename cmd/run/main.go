package main

import (
	"context"
	"log"
	"os"

	"github.com/kudoas/sync-issue-field/config"
	"github.com/kudoas/sync-issue-field/infra"
	"github.com/shurcooL/githubv4"
)

func main() {
	env, err := config.ProvideEnv()
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	g := infra.NewGithubClient(
		infra.WithClient(ctx, env.Token()),
		infra.WithContext(ctx),
	)

	q := infra.QueryRequest{
		RepositoryOwner: env.RepoOwner(),
		RepositoryName:  env.RepoName(),
		IssueNumber:     env.IssueNumber(),
	}
	trackedIssueNodeIDs := g.GetTrackedIssueNodeIDs(&q)
	targetIssueNodeID := g.GetIssueNodeID(&q)
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
