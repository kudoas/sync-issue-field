package main

import (
	"context"

	"github.com/shurcooL/githubv4"
)

type MutationIssue struct {
	UpdateIssue struct {
		Issue struct {
			ID githubv4.ID
		}
	} `graphql:"updateIssue(input: $input)"`
}

func NewMutationIssue() *MutationIssue {
	return &MutationIssue{}
}

func (m *MutationIssue) Mutate(client *githubv4.Client, ctx context.Context, variables githubv4.UpdateIssueInput) error {
	return client.Mutate(ctx, &m, variables, nil)
}
