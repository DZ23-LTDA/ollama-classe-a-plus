package agent

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestEvaluationRunnerProducesReport(t *testing.T) {
	runner := EvaluationRunner{Cases: []EvaluationCase{{ID: "one", Category: "security", Expected: []string{"blocked"}}, {ID: "two", Category: "ui", Expected: []string{"offline"}}}}
	report := runner.Run(context.Background(), func(_ context.Context, item EvaluationCase) (string, error) {
		if item.ID == "one" {
			return "blocked by policy", nil
		}
		return "offline fallback", nil
	})
	if report.Total != 2 || report.Passed != 2 || report.Failed != 0 || report.PassRate != 1 {
		t.Fatalf("report=%+v", report)
	}
}

func TestEvaluationRunnerRecordsErrorsAndMissingMarkers(t *testing.T) {
	runner := EvaluationRunner{Cases: []EvaluationCase{{ID: "error"}, {ID: "marker", Expected: []string{"approval"}}}}
	report := runner.Run(context.Background(), func(_ context.Context, item EvaluationCase) (string, error) {
		if item.ID == "error" {
			return "", errors.New("provider unavailable")
		}
		return "not applicable", nil
	})
	if report.Passed != 0 || report.Failed != 2 || report.Results[0].Error == "" || report.Results[1].Error == "" {
		t.Fatalf("report=%+v", report)
	}
}

func TestEvaluationRunnerAppliesTimeout(t *testing.T) {
	runner := EvaluationRunner{Cases: []EvaluationCase{{ID: "timeout", Timeout: 5 * time.Millisecond}}}
	report := runner.Run(context.Background(), func(ctx context.Context, _ EvaluationCase) (string, error) {
		<-ctx.Done()
		return "", ctx.Err()
	})
	if report.Results[0].Error == "" || report.Passed != 0 {
		t.Fatalf("report=%+v", report)
	}
}

func TestDefaultEvaluationCasesCoverCoreSurfaces(t *testing.T) {
	cases := DefaultEvaluationCases()
	if len(cases) < 10 {
		t.Fatalf("expected core evaluation catalog, got %d", len(cases))
	}
	seen := map[string]bool{}
	for _, item := range cases {
		seen[item.Category] = true
	}
	for _, category := range []string{"coding", "browser", "tool-calling", "security", "memory", "planning", "providers", "ui", "regression", "recovery"} {
		if !seen[category] {
			t.Fatalf("missing category %q", category)
		}
	}
}
