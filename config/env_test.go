package config

import (
	"os"
	"testing"
)

func TestProvideEnv(t *testing.T) {
	tests := []struct {
		name                string
		envVars             map[string]string
		expectError         bool
		expectedRepoOwner   string
		expectedRepoName    string
		expectedIssueNumber int
	}{
		{
			name: "valid environment variables",
			envVars: map[string]string{
				"GITHUB_TOKEN":      "test_token",
				"GITHUB_REPOSITORY": "test_owner/test_repo",
				"GITHUB_ISSUE":      "123",
			},
			expectError:         false,
			expectedRepoOwner:   "test_owner",
			expectedRepoName:    "test_repo",
			expectedIssueNumber: 123,
		},
		{
			name: "missing GITHUB_TOKEN",
			envVars: map[string]string{
				"GITHUB_REPOSITORY": "test_owner/test_repo",
				"GITHUB_ISSUE":      "123",
			},
			expectError:         true,
			expectedRepoOwner:   "",
			expectedRepoName:    "",
			expectedIssueNumber: 0,
		},
		{
			name: "invalid GITHUB_REPOSITORY format",
			envVars: map[string]string{
				"GITHUB_TOKEN":      "test_token",
				"GITHUB_REPOSITORY": "invalid_repo",
				"GITHUB_ISSUE":      "123",
			},
			expectError:         true,
			expectedRepoOwner:   "",
			expectedRepoName:    "",
			expectedIssueNumber: 0,
		},
		{
			name: "invalid GITHUB_ISSUE not a number",
			envVars: map[string]string{
				"GITHUB_TOKEN":      "test_token",
				"GITHUB_REPOSITORY": "test_owner/test_repo",
				"GITHUB_ISSUE":      "abc",
			},
			expectError:         true,
			expectedRepoOwner:   "",
			expectedRepoName:    "",
			expectedIssueNumber: 0,
		},
		{
			name: "invalid GITHUB_ISSUE zero",
			envVars: map[string]string{
				"GITHUB_TOKEN":      "test_token",
				"GITHUB_REPOSITORY": "test_owner/test_repo",
				"GITHUB_ISSUE":      "0",
			},
			expectError:         true,
			expectedRepoOwner:   "",
			expectedRepoName:    "",
			expectedIssueNumber: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			env, err := ProvideEnv()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
				if env.Token() != "test_token" ||
					env.RepoOwner() != tt.expectedRepoOwner ||
					env.RepoName() != tt.expectedRepoName ||
					env.IssueNumber() != tt.expectedIssueNumber {
					t.Errorf("Expected Env{Token: %s, RepoOwner: %s, RepoName: %s, IssueNumber: %d}, but got {Token: %s, RepoOwner: %s, RepoName: %s, IssueNumber: %d}",
						"test_token", tt.expectedRepoOwner, tt.expectedRepoName, tt.expectedIssueNumber,
						env.Token(), env.RepoOwner(), env.RepoName(), env.IssueNumber())
				}
			}
		})
	}
}
