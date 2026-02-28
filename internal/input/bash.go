package input

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// BashResult represents the result of a bash command execution
type BashResult struct {
	Command  string
	Output   string
	Error    string
	ExitCode int
	Duration time.Duration
}

// ExecuteBash executes a bash command and returns the result
func ExecuteBash(ctx context.Context, command string) BashResult {
	start := time.Now()

	// Use background context if none provided
	if ctx == nil {
		ctx = context.Background()
	}

	// Create command with bash
	cmd := exec.CommandContext(ctx, "bash", "-c", command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := BashResult{
		Command:  command,
		Output:   stdout.String(),
		Error:    stderr.String(),
		Duration: time.Since(start),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
	}

	return result
}

// FormatBashResult formats a bash result for display
func FormatBashResult(result BashResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("$ %s\n", result.Command))

	if result.Output != "" {
		sb.WriteString(result.Output)
		if !strings.HasSuffix(result.Output, "\n") {
			sb.WriteString("\n")
		}
	}

	if result.Error != "" {
		sb.WriteString(fmt.Sprintf("stderr: %s", result.Error))
		if !strings.HasSuffix(result.Error, "\n") {
			sb.WriteString("\n")
		}
	}

	if result.ExitCode != 0 {
		sb.WriteString(fmt.Sprintf("exit code: %d\n", result.ExitCode))
	}

	sb.WriteString(fmt.Sprintf("(%s)\n", result.Duration.Round(time.Millisecond)))

	return sb.String()
}

// FormatBashResultForContext formats a bash result to add to the conversation context
func FormatBashResultForContext(result BashResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Command: %s\n", result.Command))

	if result.Output != "" {
		sb.WriteString("Output:\n")
		// Truncate very long output
		output := result.Output
		maxLen := 4000
		if len(output) > maxLen {
			output = output[:maxLen] + fmt.Sprintf("\n... (truncated, %d bytes total)", len(result.Output))
		}
		sb.WriteString(output)
		if !strings.HasSuffix(output, "\n") {
			sb.WriteString("\n")
		}
	}

	if result.Error != "" {
		sb.WriteString("Stderr:\n")
		sb.WriteString(result.Error)
		if !strings.HasSuffix(result.Error, "\n") {
			sb.WriteString("\n")
		}
	}

	if result.ExitCode != 0 {
		sb.WriteString(fmt.Sprintf("Exit code: %d\n", result.ExitCode))
	}

	return sb.String()
}
