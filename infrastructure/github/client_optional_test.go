package github_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/kudoas/sync-issue-field/infrastructure/github"
	"github.com/kudoas/sync-issue-field/usecase"
	"golang.org/x/oauth2"
)

func TestOptionalFieldsOverHTTP(t *testing.T) {
	cases := []struct {
		name                          string
		options                       github.SyncOptions
		clearMilestone, emptyProjects bool
		failure                       string
	}{
		{name: "defaults leave optional child values alone"},
		{name: "only milestone", options: github.SyncOptions{Milestone: true}},
		{name: "only projects", options: github.SyncOptions{Projects: true}},
		{name: "both", options: github.SyncOptions{Milestone: true, Projects: true}},
		{name: "clear milestone", options: github.SyncOptions{Milestone: true}, clearMilestone: true},
		{name: "no parent project", options: github.SyncOptions{Projects: true}, emptyProjects: true},
		{name: "milestone read denied", options: github.SyncOptions{Milestone: true}, failure: "milestone"},
		{name: "project read denied before issue update", options: github.SyncOptions{Milestone: true, Projects: true}, failure: "projectItems"},
		{name: "project mutation fails after issue update", options: github.SyncOptions{Projects: true}, failure: "addProjectV2ItemById"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reads := map[string]int{}
			issueWrites, projectWrites := 0, 0
			milestone := "child-milestone"
			projects := []string{"child-project"}
			client := &http.Client{Transport: paginationTransport(func(r *http.Request) (*http.Response, error) {
				defer r.Body.Close()
				var request struct {
					Query     string
					Variables map[string]json.RawMessage
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Fatal(err)
				}
				rec := httptest.NewRecorder()
				rec.Header().Set("Content-Type", "application/json")
				respond := func(body string) (*http.Response, error) {
					if _, err := io.WriteString(rec, body); err != nil {
						t.Fatal(err)
					}
					return rec.Result(), nil
				}
				for _, forbidden := range []string{"fieldValues", "optionId"} {
					if strings.Contains(request.Query, forbidden) {
						t.Fatalf("unused field queried: %s", request.Query)
					}
				}
				kind := ""
				switch {
				case strings.Contains(request.Query, "addProjectV2ItemById"):
					kind = "addProjectV2ItemById"
					projectWrites++
				case strings.Contains(request.Query, "updateIssue(input:"):
					issueWrites++
					var input map[string]json.RawMessage
					if err := json.Unmarshal(request.Variables["input"], &input); err != nil {
						t.Fatal(err)
					}
					if string(input["assigneeIds"]) != `["user"]` || string(input["labelIds"]) != `["label"]` {
						t.Fatalf("basic fields missing: %s", request.Variables["input"])
					}
					raw, present := input["milestoneId"]
					if present != tc.options.Milestone {
						t.Fatalf("milestone mutation present=%v", present)
					}
					if present {
						if string(raw) == "null" {
							milestone = ""
						} else if err := json.Unmarshal(raw, &milestone); err != nil {
							t.Fatal(err)
						}
					}
					return respond(`{"data":{"updateIssue":{"issue":{"id":"child"}}}}`)
				case strings.Contains(request.Query, "parent{id}"):
					return respond(`{"data":{"repository":{"issue":{"parent":{"id":"parent"},"trackedInIssues":{"nodes":[]}}}}}`)
				case strings.Contains(request.Query, "labels(first:"):
					if strings.Contains(request.Query, "milestone") || strings.Contains(request.Query, "projectItems") {
						t.Fatalf("basic query includes optional fields: %s", request.Query)
					}
					return respond(`{"data":{"node":{"assignees":{"nodes":[{"id":"user"}]},"labels":{"nodes":[{"id":"label"}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}}`)
				case strings.Contains(request.Query, "projectItems"):
					kind = "projectItems"
					reads[kind]++
					if !tc.options.Projects {
						t.Fatal("projects requested while disabled")
					}
				case strings.Contains(request.Query, "milestone"):
					kind = "milestone"
					reads[kind]++
					if !tc.options.Milestone {
						t.Fatal("milestone requested while disabled")
					}
				case strings.Contains(request.Query, "issue(number: $issueNumber){id}"):
					return respond(`{"data":{"repository":{"issue":{"id":"child"}}}}`)
				default:
					t.Fatalf("unexpected query: %s", request.Query)
				}
				if tc.failure == kind {
					return respond(`{"errors":[{"message":"fixture permission denied"}]}`)
				}
				switch kind {
				case "milestone":
					if tc.clearMilestone {
						return respond(`{"data":{"node":{"milestone":null}}}`)
					}
					return respond(`{"data":{"node":{"milestone":{"id":"parent-milestone"}}}}`)
				case "projectItems":
					if tc.emptyProjects {
						return respond(`{"data":{"node":{"projectItems":{"nodes":[]}}}}`)
					}
					return respond(`{"data":{"node":{"projectItems":{"nodes":[{"project":{"id":"parent-project"}},{"project":{"id":"second-project"}}]}}}}`)
				case "addProjectV2ItemById":
					var input struct{ ProjectID, ContentID string }
					if err := json.Unmarshal(request.Variables["input"], &input); err != nil {
						t.Fatal(err)
					}
					if input.ProjectID != "parent-project" || input.ContentID != "child" {
						t.Fatalf("wrong project mutation: %+v", input)
					}
					projects = append(projects, input.ProjectID)
					return respond(`{"data":{"addProjectV2ItemById":{"item":{"id":"item"}}}}`)
				}
				return nil, fmt.Errorf("unexpected query kind %s", kind)
			})}
			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client)
			err := usecase.NewSyncIssueFieldsUseCase(github.NewGithubClient(ctx, "fixture"), tc.options).Execute(ctx, "owner", "repo", 42)
			if tc.failure != "" {
				if err == nil || !strings.Contains(err.Error(), "fixture permission denied") {
					t.Fatalf("missing API error: %v", err)
				}
				if tc.failure == "addProjectV2ItemById" {
					if issueWrites != 1 || projectWrites != 1 || !strings.Contains(err.Error(), "issue fields were updated") {
						t.Fatalf("partial failure not reported: %v", err)
					}
				} else if issueWrites != 0 || projectWrites != 0 {
					t.Fatalf("mutated after read failure: %d/%d", issueWrites, projectWrites)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if issueWrites != 1 {
				t.Fatalf("issue writes=%d", issueWrites)
			}
			wantMilestone := "child-milestone"
			if tc.options.Milestone {
				wantMilestone = "parent-milestone"
				if tc.clearMilestone {
					wantMilestone = ""
				}
			}
			if milestone != wantMilestone {
				t.Fatalf("milestone=%q, want %q", milestone, wantMilestone)
			}
			wantProjects := []string{"child-project"}
			wantProjectWrites := 0
			if tc.options.Projects && !tc.emptyProjects {
				wantProjects = append(wantProjects, "parent-project")
				wantProjectWrites = 1
			}
			if !reflect.DeepEqual(projects, wantProjects) || projectWrites != wantProjectWrites {
				t.Fatalf("projects=%v writes=%d", projects, projectWrites)
			}
			for field, enabled := range map[string]bool{"milestone": tc.options.Milestone, "projectItems": tc.options.Projects} {
				want := 0
				if enabled {
					want = 1
				}
				if reads[field] != want {
					t.Fatalf("%s reads=%d, want %d", field, reads[field], want)
				}
			}
		})
	}
}
