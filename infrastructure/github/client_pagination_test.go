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

type paginationTransport func(*http.Request) (*http.Response, error)

func (f paginationTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Exercise the public client, real GraphQL HTTP decoding and use case together:
// a partial read must never become a child mutation.
func TestCompleteIssueFieldsBeforeMutation(t *testing.T) {
	cases := []struct {
		name              string
		assignees, labels int
		failPage          int
		failure           string
	}{
		{"empty sets clear child", 0, 0, -1, ""},
		{"old limits", 5, 20, -1, ""},
		{"past old limits", 6, 21, -1, ""},
		{"full first page and maximum assignees", 10, 100, -1, ""},
		{"second page", 10, 101, -1, ""},
		{"third page", 10, 201, -1, ""},
		{"second page GraphQL error", 10, 201, 1, "graphql"},
		{"last page HTTP error", 10, 201, 2, "http"},
		{"missing cursor", 10, 201, 0, "missing cursor"},
		{"repeated cursor", 10, 201, 1, "repeated cursor"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ids := func(prefix string, n int) []string {
				result := make([]string, n)
				for i := range result {
					result[i] = fmt.Sprintf("%s-%d", prefix, i)
				}
				return result
			}
			wantAssignees, wantLabels := ids("user", tc.assignees), ids("label", tc.labels)
			childAssignees, childLabels := []string{"child-only-user"}, []string{"child-only-label"}
			pages, mutations := 0, 0
			client := &http.Client{Transport: paginationTransport(func(r *http.Request) (*http.Response, error) {
				defer r.Body.Close()
				var request struct {
					Query     string
					Variables map[string]json.RawMessage
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Fatal(err)
				}
				recorder := httptest.NewRecorder()
				recorder.Header().Set("Content-Type", "application/json")
				respond := func(body string) (*http.Response, error) {
					if _, err := io.WriteString(recorder, body); err != nil {
						t.Fatal(err)
					}
					return recorder.Result(), nil
				}
				switch {
				case strings.Contains(request.Query, "updateIssue(input: $input)"):
					mutations++
					var input struct {
						ID                    string
						AssigneeIDs, LabelIDs []string
					}
					if err := json.Unmarshal(request.Variables["input"], &input); err != nil {
						t.Fatal(err)
					}
					if input.ID != "child" {
						t.Fatalf("mutation target = %q", input.ID)
					}
					childAssignees, childLabels = input.AssigneeIDs, input.LabelIDs
					return respond(`{"data":{"updateIssue":{"issue":{"id":"child"}}}}`)
				case strings.Contains(request.Query, "parent{id}"):
					return respond(`{"data":{"repository":{"issue":{"parent":{"id":"parent"},"trackedInIssues":{"nodes":[]}}}}}`)
				case strings.Contains(request.Query, "labels(first: 100"):
					page := pages
					pages++
					if pages > 4 {
						t.Fatal("pagination did not terminate")
					}
					if string(request.Variables["nodeID"]) != `"parent"` {
						t.Fatal("labels requested for wrong issue")
					}
					if page == 0 {
						if !strings.Contains(request.Query, "assignees(first: 10)") {
							t.Fatal("not fetching the supported assignee maximum")
						}
					} else {
						wantCursor := fmt.Sprintf(`"cursor-%d"`, page-1)
						if string(request.Variables["cursor"]) != wantCursor {
							t.Fatalf("cursor = %s, want %s", request.Variables["cursor"], wantCursor)
						}
					}
					if page == tc.failPage {
						if tc.failure == "http" {
							recorder.WriteHeader(http.StatusUnauthorized)
							return respond(`{"message":"Bad credentials"}`)
						}
						if tc.failure == "graphql" {
							return respond(`{"data":{"node":{"labels":{"nodes":[{"id":"partial-label"}],"pageInfo":{"hasNextPage":false,"endCursor":"partial"}}}},"errors":[{"message":"page unavailable"}]}`)
						}
					}
					start := page * 100
					end := min(start+100, tc.labels)
					nodes := make([]map[string]string, 0, end-start)
					for _, id := range wantLabels[start:end] {
						nodes = append(nodes, map[string]string{"id": id})
					}
					cursor := fmt.Sprintf("cursor-%d", page)
					if page == tc.failPage && tc.failure == "missing cursor" {
						cursor = ""
					}
					if page == tc.failPage && tc.failure == "repeated cursor" {
						cursor = "cursor-0"
					}
					node := map[string]interface{}{
						"labels": map[string]interface{}{"nodes": nodes, "pageInfo": map[string]interface{}{"hasNextPage": end < tc.labels, "endCursor": cursor}},
					}
					if page == 0 {
						users := make([]map[string]string, 0, tc.assignees)
						for _, id := range wantAssignees {
							users = append(users, map[string]string{"id": id})
						}
						node["assignees"] = map[string]interface{}{"nodes": users}
					}
					body, err := json.Marshal(map[string]interface{}{"data": map[string]interface{}{"node": node}})
					if err != nil {
						t.Fatal(err)
					}
					return respond(string(body))
				case strings.Contains(request.Query, "issue(number: $issueNumber){id}"):
					return respond(`{"data":{"repository":{"issue":{"id":"child"}}}}`)
				default:
					t.Fatalf("unexpected query (including unexpected mutation): %s", request.Query)
					return nil, nil
				}
			})}
			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client)
			uc := usecase.NewSyncIssueFieldsUseCase(github.NewGithubClient(ctx, "fixture-token"), github.SyncOptions{})
			err := uc.Execute(ctx, "example", "repo", 42)
			if tc.failure != "" {
				if err == nil {
					t.Fatal("expected read error")
				}
				if mutations != 0 || !reflect.DeepEqual(childAssignees, []string{"child-only-user"}) || !reflect.DeepEqual(childLabels, []string{"child-only-label"}) {
					t.Fatalf("child changed after read failure: mutations=%d", mutations)
				}
				if pages != tc.failPage+1 {
					t.Fatalf("pages=%d, want %d", pages, tc.failPage+1)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if mutations != 1 {
				t.Fatalf("mutations=%d, want 1", mutations)
			}
			if !reflect.DeepEqual(childAssignees, wantAssignees) || !reflect.DeepEqual(childLabels, wantLabels) {
				t.Fatalf("mutation fields differ: assignees=%v labels=%v", childAssignees, childLabels)
			}
			wantPages := max(1, (tc.labels+99)/100)
			if pages != wantPages {
				t.Fatalf("pages=%d, want %d", pages, wantPages)
			}
		})
	}
}
