package config

import "testing"

func TestOptionalSynchronizationFlags(t *testing.T) {
	for _, milestone := range []bool{false, true} {
		for _, projects := range []bool{false, true} {
			t.Run(boolString(milestone)+"/"+boolString(projects), func(t *testing.T) {
				t.Setenv("GITHUB_TOKEN", "fixture")
				t.Setenv("GITHUB_REPOSITORY", "owner/repo")
				t.Setenv("GITHUB_ISSUE", "42")
				t.Setenv("SYNC_MILESTONE", boolString(milestone))
				t.Setenv("SYNC_PROJECTS", boolString(projects))
				env, err := ProvideEnv()
				if err != nil || env.SyncMilestone() != milestone || env.SyncProjects() != projects {
					t.Fatalf("options not preserved: %+v, %v", env, err)
				}
			})
		}
	}
	for _, name := range []string{"SYNC_MILESTONE", "SYNC_PROJECTS"} {
		for _, value := range []string{"", "TRUE", "1", "yes", "invalid"} {
			t.Run(name+"="+value, func(t *testing.T) {
				t.Setenv("GITHUB_TOKEN", "fixture")
				t.Setenv("GITHUB_REPOSITORY", "owner/repo")
				t.Setenv("GITHUB_ISSUE", "42")
				t.Setenv("SYNC_MILESTONE", "")
				t.Setenv("SYNC_PROJECTS", "")
				t.Setenv(name, value)
				env, err := ProvideEnv()
				if value == "" {
					if err != nil || env.SyncMilestone() || env.SyncProjects() {
						t.Fatal("defaults must be disabled")
					}
					return
				}
				if err == nil {
					t.Fatalf("accepted invalid boolean %q", value)
				}
			})
		}
	}
}
func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
