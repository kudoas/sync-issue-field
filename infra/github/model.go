package github

import "github.com/shurcooL/githubv4"

type QueryRequest struct {
	RepositoryOwner string
	RepositoryName  string
	IssueNumber     int
}

type IssueFields struct {
	AssigneeIDs []githubv4.ID
	LabelIDs    []githubv4.ID
	MilestoneID githubv4.ID
	ProjectIDs  []githubv4.ID
}
