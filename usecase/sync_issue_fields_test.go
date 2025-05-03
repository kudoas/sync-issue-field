package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/kudoas/sync-issue-field/infrastructure/github"
	mock_github "github.com/kudoas/sync-issue-field/infrastructure/github"
	"github.com/shurcooL/githubv4"
	"go.uber.org/mock/gomock"
)

func TestSyncIssueFieldsUseCase_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_github.NewMockGitHubClient(ctrl)

	useCase := NewSyncIssueFieldsUseCase(mockClient)

	ctx := context.Background()
	repoOwner := "test_owner"
	repoName := "test_repo"
	issueNumber := 123

	t.Run("successfully syncs fields", func(t *testing.T) {
		// Mockの設定
		queryReq := &github.QueryRequest{
			RepositoryOwner: repoOwner,
			RepositoryName:  repoName,
			IssueNumber:     issueNumber,
		}
		targetNodeID := githubv4.ID("targetNodeID")
		parentID := githubv4.ID("parentID")

		mockClient.EXPECT().GetIssueNodeID(queryReq).Return(targetNodeID, nil)
		mockClient.EXPECT().GetTrackedIssueNodeIDs(queryReq).Return([]githubv4.ID{parentID}, nil)

		parentFields := &github.IssueFields{
			AssigneeIDs: []githubv4.ID{"assignee1", "assignee2"},
			LabelIDs:    []githubv4.ID{"label1", "label2"},
			MilestoneID: githubv4.ID("milestone1"),
			ProjectIDs:  []githubv4.ID{"project1"},
		}
		mockClient.EXPECT().GetIssueFields(parentID).Return(parentFields, nil)

		updateInput := githubv4.UpdateIssueInput{
			ID:          targetNodeID,
			AssigneeIDs: &parentFields.AssigneeIDs,
			LabelIDs:    &parentFields.LabelIDs,
		}
		milestoneID := parentFields.MilestoneID
		if milestoneID != "" {
			updateInput.MilestoneID = &milestoneID
		}
		mockClient.EXPECT().MutateIssue(updateInput).Return(nil)

		addProjectInput := githubv4.AddProjectV2ItemByIdInput{
			ProjectID: parentFields.ProjectIDs[0],
			ContentID: targetNodeID,
		}
		mockClient.EXPECT().MutateProject(addProjectInput).Return(nil)

		err := useCase.Execute(ctx, repoOwner, repoName, issueNumber)

		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	})

	t.Run("no tracked issues", func(t *testing.T) {
		queryReq := &github.QueryRequest{
			RepositoryOwner: repoOwner,
			RepositoryName:  repoName,
			IssueNumber:     issueNumber,
		}
		targetNodeID := githubv4.ID("targetNodeID")

		mockClient.EXPECT().GetIssueNodeID(queryReq).Return(targetNodeID, nil)
		mockClient.EXPECT().GetTrackedIssueNodeIDs(queryReq).Return([]githubv4.ID{}, nil) // tracked issueがないケース

		exited := false
		exitCode := 0
		customExiter := func(code int) {
			exited = true
			exitCode = code
			panic("os.Exit called")
		}
		useCase.exiter = customExiter

		defer func() {
			if r := recover(); r != nil {
				if r != "os.Exit called" {
					t.Errorf("Expected os.Exit panic, but got %v", r)
				}
			}
			if !exited || exitCode != 0 {
				t.Errorf("Expected os.Exit(0) to be called, but it was not")
			}
		}()

		err := useCase.Execute(ctx, repoOwner, repoName, issueNumber)

		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	})

	t.Run("GetIssueNodeID returns error", func(t *testing.T) {
		queryReq := &github.QueryRequest{
			RepositoryOwner: repoOwner,
			RepositoryName:  repoName,
			IssueNumber:     issueNumber,
		}
		expectedErr := errors.New("get issue node id error")
		mockClient.EXPECT().GetIssueNodeID(queryReq).Return(githubv4.ID(""), expectedErr)

		err := useCase.Execute(ctx, repoOwner, repoName, issueNumber)

		if err == nil || !errors.Is(err, expectedErr) {
			t.Errorf("Expected error %v, but got %v", expectedErr, err)
		}
	})
}
