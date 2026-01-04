package jj

import (
	"context"
)

// New creates a new empty revision on top of the current working copy.
func (c *Client) New(ctx context.Context) error {
	_, err := c.run(ctx, "new")
	return err
}

// NewInDir creates a new empty revision in a specific directory.
func (c *Client) NewInDir(ctx context.Context, dir string) error {
	_, err := c.runInDir(ctx, dir, "new")
	return err
}

// NewFromRevision creates a new empty revision based on a specific revision.
func (c *Client) NewFromRevision(ctx context.Context, revision string) error {
	_, err := c.run(ctx, "new", revision)
	return err
}

// NewFromRevisionInDir creates a new empty revision based on a specific revision in a specific directory.
func (c *Client) NewFromRevisionInDir(ctx context.Context, dir, revision string) error {
	_, err := c.runInDir(ctx, dir, "new", revision)
	return err
}

// Commit creates a new commit with the given message.
func (c *Client) Commit(ctx context.Context, message string) error {
	_, err := c.run(ctx, "commit", "-m", message)
	return err
}

// Squash squashes the working copy into its parent.
func (c *Client) Squash(ctx context.Context) error {
	_, err := c.run(ctx, "squash")
	return err
}

// SquashInto squashes changes from one revision into another.
func (c *Client) SquashInto(ctx context.Context, from, into string) error {
	_, err := c.run(ctx, "squash", "--from", from, "--into", into)
	return err
}

// Rebase rebases the current revision onto the destination.
func (c *Client) Rebase(ctx context.Context, destination string) error {
	_, err := c.run(ctx, "rebase", "-d", destination)
	return err
}

// Describe sets or updates the description of the working copy.
func (c *Client) Describe(ctx context.Context, message string) error {
	_, err := c.run(ctx, "describe", "-m", message)
	return err
}

// DescribeRevision sets or updates the description of a specific revision.
func (c *Client) DescribeRevision(ctx context.Context, revision, message string) error {
	_, err := c.run(ctx, "describe", revision, "-m", message)
	return err
}

// GitPush pushes all bookmarks to the remote.
func (c *Client) GitPush(ctx context.Context) error {
	_, err := c.run(ctx, "git", "push")
	return err
}

// GitPushBookmark pushes a specific bookmark to the remote.
func (c *Client) GitPushBookmark(ctx context.Context, bookmark string) error {
	_, err := c.run(ctx, "git", "push", "--bookmark", bookmark)
	return err
}
