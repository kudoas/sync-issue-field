package main

import (
	"context"
	"log"
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

	i := NewIssueID()
	variable := map[string]interface{}{
		"repositoryOwner": githubv4.String(owner),
		"repositoryName":  githubv4.String(name),
		"issueNumber":     githubv4.Int(issue),
	}
	if err := i.Query(client, context.Background(), variable); err != nil {
		log.Fatalf("failed to get issue id: %v", err)
	}

	p := NewParentIssue()
	if err := p.Query(client, context.Background(), map[string]interface{}{
		"issueID": githubv4.ID(i.GetParentIssueID()),
	}); err != nil {
		log.Fatalf("failed to get parent issue: %v", err)
	}

	m := NewMutationIssue()
	input := githubv4.UpdateIssueInput{
		ID:          i.GetIssueID(),
		AssigneeIDs: p.GetAssigneeIDs(),
		LabelIDs:    p.GetLabelIDs(),
		MilestoneID: p.GetMilestoneID(),
	}
	if err := m.Mutate(client, context.Background(), input); err != nil {
		log.Fatalf("failed to update issue: %v", err)
	}
}
