package pulumi

import (
	"errors"
	"fmt"
	"sort"
)

type unsupportedError struct {
	component string
	path      string
	summary   string
	skippable bool
}

func (e *unsupportedError) Error() string {
	return fmt.Sprintf("component %q path %q unsupported construct: %s", e.component, e.path, e.summary)
}

type skippedPathIssue struct {
	component string
	path      string
	reason    string
}

func maybeRecordSkippedPath(ctx *translationContext, err error) bool {
	var unsupportedErr *unsupportedError
	if !errors.As(err, &unsupportedErr) || !unsupportedErr.skippable {
		return false
	}

	ctx.skipped = append(ctx.skipped, skippedPathIssue{
		component: unsupportedErr.component,
		path:      unsupportedErr.path,
		reason:    unsupportedErr.summary,
	})
	return true
}

func sortedSkippedIssues(issues []skippedPathIssue) []skippedPathIssue {
	if len(issues) == 0 {
		return nil
	}

	result := make([]skippedPathIssue, len(issues))
	copy(result, issues)
	sort.Slice(result, func(i, j int) bool {
		if result[i].path != result[j].path {
			return result[i].path < result[j].path
		}
		if result[i].reason != result[j].reason {
			return result[i].reason < result[j].reason
		}
		return result[i].component < result[j].component
	})
	return result
}

func toWarningMessages(issues []skippedPathIssue) []string {
	sorted := sortedSkippedIssues(issues)
	if len(sorted) == 0 {
		return nil
	}

	warnings := make([]string, 0, len(sorted))
	for _, issue := range sorted {
		warnings = append(warnings, fmt.Sprintf(
			"warning: skipped unsupported schema path %q for component %q: %s",
			issue.path,
			issue.component,
			issue.reason,
		))
	}

	return warnings
}

func unsupported(componentToken, path, summary string) error {
	return &unsupportedError{
		component: componentToken,
		path:      path,
		summary:   summary,
		skippable: true,
	}
}

func unsupportedHard(componentToken, path, summary string) error {
	return &unsupportedError{
		component: componentToken,
		path:      path,
		summary:   summary,
		skippable: false,
	}
}
