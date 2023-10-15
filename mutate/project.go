package mutate

import (
	"context"

	"github.com/shurcooL/githubv4"
)

type MutationProject struct {
	AddProjectV2ItemById struct {
		Item struct {
			ID githubv4.ID `graphql:"id"`
		} `graphql:"item"`
	} `graphql:"addProjectV2ItemById(input: $input)"`
}

func NewMutationProject() *MutationProject {
	return &MutationProject{}
}

func (m *MutationProject) Mutate(client *githubv4.Client, variables githubv4.AddProjectV2ItemByIdInput) error {
	return client.Mutate(context.Background(), &m, variables, nil)
}
