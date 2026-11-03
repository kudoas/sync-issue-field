package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/kudoas/sync-issue-field/infrastructure/github"
	"github.com/shurcooL/githubv4"
)

type SyncIssueFieldsUseCase struct {
	githubClient github.GitHubClient
	options      github.SyncOptions
}

func NewSyncIssueFieldsUseCase(githubClient github.GitHubClient, options github.SyncOptions) *SyncIssueFieldsUseCase {
	return &SyncIssueFieldsUseCase{
		githubClient: githubClient,
		options:      options,
	}
}

func (uc *SyncIssueFieldsUseCase) Execute(ctx context.Context, repoOwner, repoName string, issueNumber int) error {
	q := github.QueryRequest{
		RepositoryOwner: repoOwner,
		RepositoryName:  repoName,
		IssueNumber:     issueNumber,
	}
	targetIssueNodeID, err := uc.githubClient.GetIssueNodeID(&q)
	if err != nil {
		return fmt.Errorf("failed to get target issue node id: %w", err)
	}

	parentID, err := uc.githubClient.GetParentIssueNodeID(&q)
	if err != nil {
		return fmt.Errorf("failed to get parent issue node id: %w", err)
	}
	if parentID == nil {
		log.Println("No parent issue found. Skipping synchronization.")
		return nil
	}

	parentIssueFields, err := uc.githubClient.GetIssueFields(parentID, uc.options)
	if err != nil {
		return fmt.Errorf("failed to get parent issue fields: %w", err)
	}

	updateInput := githubv4.UpdateIssueInput{
		ID:          targetIssueNodeID,
		AssigneeIDs: &parentIssueFields.AssigneeIDs,
		LabelIDs:    &parentIssueFields.LabelIDs,
	}

	if uc.options.Milestone {
		updateInput.MilestoneID = &parentIssueFields.MilestoneID
	}

	if err := uc.githubClient.MutateIssue(updateInput); err != nil {
		return fmt.Errorf("failed to update target issue: %w", err)
	}

	if uc.options.Projects && len(parentIssueFields.ProjectIDs) > 0 {
		if err := uc.githubClient.MutateProject(githubv4.AddProjectV2ItemByIdInput{
			ProjectID: parentIssueFields.ProjectIDs[0],
			ContentID: targetIssueNodeID,
		}); err != nil {
			return fmt.Errorf("issue fields were updated, but failed to add project item: %w", err)
		}
	}

	return nil
}
