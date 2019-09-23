package github

import (
	"context"

	gh "github.com/google/go-github/v28/github"
	"github.com/marccarre/github-oauth-cli/pkg/oauth"
	"golang.org/x/oauth2"
)

// Client is a high-level GitHub client, wrapping around a google/go-github/github#Client.
type Client struct {
	username string
	client   *gh.Client
}

// NewClient returns an authenticated GitHub client.
func NewClient(ctx context.Context, username string) (*Client, error) {
	token, err := oauth.GetToken(ctx, username)
	if err != nil {
		return nil, err
	}
	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	return &Client{
		username: username,
		client:   gh.NewClient(oauthClient),
	}, nil
}
