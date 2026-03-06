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

func Run(ctx context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	if _, err := exec.LookPath(name); err != nil {
		return "", fmt.Errorf("tool %q not found in PATH", name)
	}

	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, name, args...)
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
		return combined, fmt.Errorf("command timed out: %s %s", name, strings.Join(args, " "))
	}

	return combined, fmt.Errorf("command failed: %s %s: %w", name, strings.Join(args, " "), err)
}
