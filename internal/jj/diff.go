package jj

import (
	"context"
	"strings"
)

// DiffOptions configures the diff command.
type DiffOptions struct {
	Revision string // Revision to show diff for (defaults to working copy)
	WorkDir  string // Directory to run diff in (for non-default workspaces)
}

// Diff returns the raw diff output with color codes.
// Automatically handles stale workspaces by updating them first.
func (c *Client) Diff(ctx context.Context, opts *DiffOptions) (string, error) {
	args := []string{"diff", "--color=always"}

	if opts != nil && opts.Revision != "" {
		args = append(args, "-r", opts.Revision)
	}

	workDir := ""
	if opts != nil && opts.WorkDir != "" {
		workDir = opts.WorkDir
	}

	result, err := c.runInDir(ctx, workDir, args...)
	if err != nil {
		// Check if this is a stale workspace error
		if cmdErr, ok := err.(*CommandError); ok && strings.Contains(cmdErr.Stderr, "working copy is stale") {
			// Try to update stale workspace and retry
			if updateErr := c.WorkspaceUpdateStale(ctx, workDir); updateErr == nil {
				// Retry the diff
				return c.runInDir(ctx, workDir, args...)
			}
		}
	}

	return result, err
}
