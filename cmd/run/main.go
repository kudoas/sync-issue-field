package main

import (
	"context"
	"log"
	"os"

	"github.com/kudoas/sync-issue-field/config"
	"github.com/kudoas/sync-issue-field/infra/github"
	"github.com/shurcooL/githubv4"
)

func main() {
	env, err := config.ProvideEnv()
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	g := github.NewGithubClient(
		ctx, env.Token(),
	)

	q := github.QueryRequest{
		RepositoryOwner: env.RepoOwner(),
		RepositoryName:  env.RepoName(),
		IssueNumber:     env.IssueNumber(),
	}
	trackedIssueNodeIDs, err := g.GetTrackedIssueNodeIDs(&q)
	if err != nil {
		log.Fatalf("failed to get tracked issue node ids: %v", err)
	}
	targetIssueNodeID, err := g.GetIssueNodeID(&q)
	if err != nil {
		log.Fatalf("failed to get issue node id: %v", err)
	}
	if len(trackedIssueNodeIDs) == 0 {
		os.Exit(0)
	}
	parentID := trackedIssueNodeIDs[0]
	parentIssueFields, err := g.GetIssueFields(parentID)
	if err != nil {
		log.Fatalf("failed to get issue fields: %v", err)
	}

	if err := g.MutateIssue(
		githubv4.UpdateIssueInput{
			ID:          targetIssueNodeID,
			AssigneeIDs: &parentIssueFields.AssigneeIDs,
			LabelIDs:    &parentIssueFields.LabelIDs,
			MilestoneID: &parentIssueFields.MilestoneID,
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
