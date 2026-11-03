package distribution

import (
	"os"
	"strings"
	"testing"
)

// The workflow guard prevents issuing a write token for other refs. This guard
// also prevents accidental fixture writes when the Go test is invoked directly.
func liveTestSkipReason() string {
	if os.Getenv("LIVE_RELEASE_TEST") != "1" {
		return "live API verification not requested; normal CI (including fork PRs) runs without a PAT"
	}
	if os.Getenv("GITHUB_EVENT_NAME") != "workflow_dispatch" || os.Getenv("GITHUB_REPOSITORY") != "kudoas/sync-issue-field" || os.Getenv("GITHUB_REF") != "refs/heads/main" {
		return "live API verification requires a manual workflow_dispatch on kudoas/sync-issue-field main"
	}
	if os.Getenv("GITHUB_TOKEN") == "" {
		return "live API verification skipped: GITHUB_TOKEN is unavailable"
	}
	return ""
}

func TestLiveVerificationEligibility(t *testing.T) {
	cases := []struct {
		name, optIn, event, repository, ref, token, want string
	}{
		{"fork PR without secrets", "0", "pull_request", "kudoas/sync-issue-field", "refs/pull/42/merge", "", "not requested"},
		{"fork PR cannot opt in", "1", "pull_request", "kudoas/sync-issue-field", "refs/pull/42/merge", "", "requires a manual"},
		{"PR with token still cannot opt in", "1", "pull_request", "kudoas/sync-issue-field", "refs/pull/42/merge", "fixture", "requires a manual"},
		{"pull_request_target rejected", "1", "pull_request_target", "kudoas/sync-issue-field", "refs/heads/main", "fixture", "requires a manual"},
		{"push to main rejected", "1", "push", "kudoas/sync-issue-field", "refs/heads/main", "fixture", "requires a manual"},
		{"manual feature branch rejected", "1", "workflow_dispatch", "kudoas/sync-issue-field", "refs/heads/feature", "fixture", "requires a manual"},
		{"manual fork repository rejected", "1", "workflow_dispatch", "contributor/sync-issue-field", "refs/heads/main", "fixture", "requires a manual"},
		{"manual tag rejected", "1", "workflow_dispatch", "kudoas/sync-issue-field", "refs/tags/v0.3.0", "fixture", "requires a manual"},
		{"missing token skipped", "1", "workflow_dispatch", "kudoas/sync-issue-field", "refs/heads/main", "", "GITHUB_TOKEN is unavailable"},
		{"trusted authenticated manual run", "1", "workflow_dispatch", "kudoas/sync-issue-field", "refs/heads/main", "fixture", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for name, value := range map[string]string{"LIVE_RELEASE_TEST": tc.optIn, "GITHUB_EVENT_NAME": tc.event, "GITHUB_REPOSITORY": tc.repository, "GITHUB_REF": tc.ref, "GITHUB_TOKEN": tc.token} {
				t.Setenv(name, value)
			}
			reason := liveTestSkipReason()
			if tc.want == "" {
				if reason != "" {
					t.Fatalf("trusted live run skipped: %s", reason)
				}
				return
			}
			if !strings.Contains(reason, tc.want) {
				t.Fatalf("skip reason=%q, want %q", reason, tc.want)
			}
		})
	}
}
