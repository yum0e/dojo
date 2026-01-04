package jj

import (
	"context"
)

// DiffOptions configures the diff command.
type DiffOptions struct {
	Revision string // Revision to show diff for (defaults to working copy)
	WorkDir  string // Directory to run diff in (for non-default workspaces)
}

// Diff returns the raw diff output with color codes.
func (c *Client) Diff(ctx context.Context, opts *DiffOptions) (string, error) {
	args := []string{"diff", "--color=always"}

	if opts != nil && opts.Revision != "" {
		args = append(args, "-r", opts.Revision)
	}

	workDir := ""
	if opts != nil && opts.WorkDir != "" {
		workDir = opts.WorkDir
	}

	return c.runInDir(ctx, workDir, args...)
}
