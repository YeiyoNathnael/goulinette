package tools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Run executes the named external tool with the given arguments in the
// process working directory. It is a convenience wrapper around RunInDir
// with an empty dir.
func Run(ctx context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	return RunInDir(ctx, timeout, "", name, args...)
}

// RunInDir executes the named external tool with the given arguments, running
// from dir (or the process working directory when dir is empty). It enforces
// a deadline via timeout, verifies the tool is present in PATH before
// launching, and merges stdout and stderr into a single trimmed string on
// success. On failure it returns an empty string and a descriptive error that
// distinguishes between a missing tool, a deadline exceeded, and a non-zero
// exit status.
func RunInDir(ctx context.Context, timeout time.Duration, dir string, name string, args ...string) (string, error) {
	if _, err := exec.LookPath(name); err != nil {
		return "", fmt.Errorf("tool %q not found in PATH", name)
	}

	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	combined := strings.TrimSpace(stdout.String() + "\n" + stderr.String())
	if err == nil {
		return combined, nil
	}

	if errors.Is(cctx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("command timed out: %s %s", name, strings.Join(args, " "))
	}

	return "", fmt.Errorf("command failed: %s %s: %w", name, strings.Join(args, " "), err)
}
