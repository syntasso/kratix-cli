package pulumi

import (
	"errors"
	"testing"
)

func TestMaybeRecordSkippedPath(t *testing.T) {
	t.Parallel()

	ctx := &translationContext{}
	if !maybeRecordSkippedPath(ctx, unsupported("pkg:index:Thing", "spec.alpha", `keyword "oneOf"`)) {
		t.Fatal("expected skippable unsupported error to be recorded")
	}
	if len(ctx.skipped) != 1 {
		t.Fatalf("expected one skipped issue, got %d", len(ctx.skipped))
	}

	if maybeRecordSkippedPath(ctx, unsupportedHard("pkg:index:Thing", "spec.alpha", "hard")) {
		t.Fatal("did not expect hard unsupported error to be recorded as skipped")
	}
	if maybeRecordSkippedPath(ctx, errors.New("boom")) {
		t.Fatal("did not expect generic error to be recorded as skipped")
	}
}

func TestToWarningMessages(t *testing.T) {
	t.Parallel()

	got := toWarningMessages([]skippedPathIssue{
		{component: "b", path: "spec.b", reason: "r2"},
		{component: "a", path: "spec.a", reason: "r2"},
		{component: "a", path: "spec.a", reason: "r1"},
	})

	want := []string{
		`warning: skipped unsupported schema path "spec.a" for component "a": r1`,
		`warning: skipped unsupported schema path "spec.a" for component "a": r2`,
		`warning: skipped unsupported schema path "spec.b" for component "b": r2`,
	}
	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("mismatch at %d: got %q want %q", i, got[i], want[i])
		}
	}
}
