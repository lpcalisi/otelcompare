package github

import (
	"context"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// Client represents a GitHub client
type Client struct {
	client *github.Client
	ctx    context.Context
}

// NewClient creates a new GitHub client
func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Client{
		client: client,
		ctx:    ctx,
	}
}

// CommentPR adds a comment to a PR with the trace visualization
func (c *Client) CommentPR(owner, repo string, prNumber int, htmlContent string) error {
	_, _, err := c.client.Issues.CreateComment(c.ctx, owner, repo, prNumber, &github.IssueComment{
		Body: &htmlContent,
	})

	return err
}

// CompareTraces compares traces between two versions and generates a comment in the PR
func (c *Client) CompareTraces(owner, repo string, prNumber int, baseHTML, headHTML string) error {
	// TODO: Implement trace comparison
	return nil
}
