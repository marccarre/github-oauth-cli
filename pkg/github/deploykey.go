package github

import (
	"context"

	"github.com/google/go-github/v28/github"
)

// AddDeployKey adds the provided deploy key to the provided repository.
func (c Client) AddDeployKey(ctx context.Context, repo string, key *github.Key) {
	c.client.Repositories.CreateKey(ctx, c.username, repo, key)
}
