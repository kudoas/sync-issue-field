package main

import (
	"context"
	"log"

	"github.com/kudoas/sync-issue-field/config"
	"github.com/kudoas/sync-issue-field/infrastructure/github"
	"github.com/kudoas/sync-issue-field/usecase"
)

func main() {
	env, err := config.ProvideEnv()
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	githubClient := github.NewGithubClient(ctx, env.Token())
	syncUseCase := usecase.NewSyncIssueFieldsUseCase(githubClient)
	if err := syncUseCase.Execute(ctx, env.RepoOwner(), env.RepoName(), env.IssueNumber()); err != nil {
		log.Fatalf("failed to execute sync issue fields use case: %v", err)
	}

	log.Println("Successfully synced issue fields.")
}
