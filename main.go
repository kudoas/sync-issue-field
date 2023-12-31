package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/kudoas/sync-issue-field/mutate"
	"github.com/kudoas/sync-issue-field/query"
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
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	i := query.NewIssueID()
	variable := map[string]interface{}{
		"repositoryOwner": githubv4.String(owner),
		"repositoryName":  githubv4.String(repository_name),
		"issueNumber":     githubv4.Int(issue),
	}
	if err := i.Query(client, context.Background(), variable); err != nil {
		log.Fatalf("failed to get issue id: %v", err)
	}

	p := query.NewParentIssue()
	if err := p.Query(client, context.Background(), map[string]interface{}{
		"issueID": githubv4.ID(i.GetParentIssueID()),
	}); err != nil {
		log.Fatalf("failed to get parent issue: %v", err)
	}
	mi := mutate.NewMutationIssue()
	input := githubv4.UpdateIssueInput{
		ID:          i.GetIssueID(),
		AssigneeIDs: p.GetAssigneeIDs(),
		LabelIDs:    p.GetLabelIDs(),
		MilestoneID: p.GetMilestoneID(),
	}
	if err := mi.Mutate(client, context.Background(), input); err != nil {
		log.Fatalf("failed to update issue: %v", err)
	}
	mp := mutate.NewMutationProject()
	if err := mp.Mutate(client, githubv4.AddProjectV2ItemByIdInput{
		ProjectID: p.GetProjectID(),
		ContentID: i.GetIssueID(),
	}); err != nil {
		log.Fatalf("failed to add project item: %v", err)
	}
}
