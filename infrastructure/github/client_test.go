package github

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/shurcooL/githubv4"
)

type fixtureTransport func(*http.Request) (*http.Response, error)

func (f fixtureTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type graphQLRequest struct {
	Query     string                     `json:"query"`
	Variables map[string]json.RawMessage `json:"variables"`
}

// Use the real GraphQL client and HTTP request/response handling without a
// network connection, credentials, or writes to a shared GitHub Issue.
func fixtureClient(t *testing.T, queryPart, response string, check func(graphQLRequest), status ...int) GitHubClient {
	t.Helper()
	client := &http.Client{Transport: fixtureTransport(func(r *http.Request) (*http.Response, error) {
		t.Helper()
		defer r.Body.Close()
		if r.Method != http.MethodPost {
			t.Errorf("HTTP method = %s, want POST", r.Method)
		}
		var request graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(request.Query, queryPart) {
			t.Errorf("query %q does not contain %q", request.Query, queryPart)
		}
		if check != nil {
			check(request)
		}
		recorder := httptest.NewRecorder()
		recorder.Header().Set("Content-Type", "application/json")
		if len(status) > 0 {
			recorder.WriteHeader(status[0])
		}
		if _, err := io.WriteString(recorder, response); err != nil {
			t.Fatal(err)
		}
		return recorder.Result(), nil
	})}
	return &githubClient{client: githubv4.NewEnterpriseClient("https://fixture.invalid/graphql", client), context: context.Background()}
}

func TestIssueQueriesOverHTTP(t *testing.T) {
	query := &QueryRequest{RepositoryOwner: "example", RepositoryName: "repository", IssueNumber: 42}
	check := func(request graphQLRequest) {
		t.Helper()
		want := map[string]string{"repositoryOwner": `"example"`, "repositoryName": `"repository"`, "issueNumber": "42"}
		for name, value := range want {
			if string(request.Variables[name]) != value {
				t.Errorf("variable %s = %s, want %s", name, request.Variables[name], value)
			}
		}
	}
	t.Run("issue ID", func(t *testing.T) {
		client := fixtureClient(t, "issue(number: $issueNumber)", `{"data":{"repository":{"issue":{"id":"child"}}}}`, check)
		id, err := client.GetIssueNodeID(query)
		if err != nil || id != "child" {
			t.Fatalf("ID = %v, error = %v", id, err)
		}
	})
}

func TestParentSelectionOverHTTP(t *testing.T) {
	cases := []struct {
		name, parent, tracked string
		want                  githubv4.ID
		wantError             bool
	}{
		{"Sub-issue without tracking", `{"id":"native"}`, `[]`, "native", false},
		{"Sub-issue takes priority over tracking", `{"id":"native"}`, `[{"id":"legacy1"},{"id":"legacy2"}]`, "native", false},
		{"unique checklist fallback", `null`, `[{"id":"legacy"}]`, "legacy", false},
		{"no parent", `null`, `[]`, nil, false},
		{"ambiguous checklist fallback", `null`, `[{"id":"legacy1"},{"id":"legacy2"}]`, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := `{"data":{"repository":{"issue":{"parent":` + tc.parent + `,"trackedInIssues":{"nodes":` + tc.tracked + `}}}}}`
			client := fixtureClient(t, "parent{id}", response, func(request graphQLRequest) {
				if !strings.Contains(request.Query, "trackedInIssues(first: 2)") {
					t.Errorf("missing bounded compatibility query: %s", request.Query)
				}
				for name, want := range map[string]string{"repositoryOwner": `"example"`, "repositoryName": `"repository"`, "issueNumber": "42"} {
					if string(request.Variables[name]) != want {
						t.Errorf("%s = %s, want %s", name, request.Variables[name], want)
					}
				}
			})
			id, err := client.GetParentIssueNodeID(&QueryRequest{RepositoryOwner: "example", RepositoryName: "repository", IssueNumber: 42})
			if tc.wantError {
				if err == nil || !strings.Contains(err.Error(), "multiple tracking issues") {
					t.Fatalf("expected ambiguous parent error, got %v", err)
				}
			} else if err != nil || id != tc.want {
				t.Fatalf("parent = %v, error = %v, want %v", id, err, tc.want)
			}
		})
	}
}

func TestIssueFieldsOverHTTP(t *testing.T) {
	response := `{"data":{"node":{"assignees":{"nodes":[{"id":"user"}]},"labels":{"nodes":[{"id":"label"}]},"milestone":{"id":"milestone"},"projectItems":{"nodes":[{"project":{"id":"project"},"fieldValues":{"nodes":[]}}]}}}}`
	client := fixtureClient(t, "node(id: $nodeID)", response, func(request graphQLRequest) {
		if string(request.Variables["nodeID"]) != `"parent"` {
			t.Errorf("nodeID = %s", request.Variables["nodeID"])
		}
	})
	fields, err := client.GetIssueFields(githubv4.ID("parent"))
	if err != nil {
		t.Fatal(err)
	}
	// Compare the serialized values, since the implementation returns a pointer
	// wrapped in the ID interface for the milestone.
	actual, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	want, err := json.Marshal(&IssueFields{
		AssigneeIDs: []githubv4.ID{"user"}, LabelIDs: []githubv4.ID{"label"},
		MilestoneID: githubv4.ID("milestone"), ProjectIDs: []githubv4.ID{"project"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != string(want) {
		t.Fatalf("fields = %s, want %s", actual, want)
	}
}

func TestMutationsOverHTTP(t *testing.T) {
	assignees, labels := []githubv4.ID{"user"}, []githubv4.ID{"label"}
	milestone := githubv4.ID("milestone")
	cases := []struct {
		name, query, response, input string
		mutate                       func(GitHubClient) error
	}{
		{
			name: "issue", query: "updateIssue(input: $input)",
			response: `{"data":{"updateIssue":{"issue":{"id":"child"}}}}`,
			input:    `{"id":"child","assigneeIds":["user"],"labelIds":["label"],"milestoneId":"milestone"}`,
			mutate: func(client GitHubClient) error {
				return client.MutateIssue(githubv4.UpdateIssueInput{ID: "child", AssigneeIDs: &assignees, LabelIDs: &labels, MilestoneID: &milestone})
			},
		},
		{
			name: "project", query: "addProjectV2ItemById(input: $input)",
			response: `{"data":{"addProjectV2ItemById":{"item":{"id":"item"}}}}`,
			input:    `{"projectId":"project","contentId":"child"}`,
			mutate: func(client GitHubClient) error {
				return client.MutateProject(githubv4.AddProjectV2ItemByIdInput{ProjectID: "project", ContentID: "child"})
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := fixtureClient(t, tc.query, tc.response, func(request graphQLRequest) {
				var actual, want map[string]interface{}
				if err := json.Unmarshal(request.Variables["input"], &actual); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal([]byte(tc.input), &want); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(actual, want) {
					t.Errorf("input = %v, want %v", actual, want)
				}
			})
			if err := tc.mutate(client); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAPIErrorsAreReturned(t *testing.T) {
	calls := map[string]func(GitHubClient) error{
		"issue ID":     func(client GitHubClient) error { _, err := client.GetIssueNodeID(&QueryRequest{}); return err },
		"parent ID":    func(client GitHubClient) error { _, err := client.GetParentIssueNodeID(&QueryRequest{}); return err },
		"fields":       func(client GitHubClient) error { _, err := client.GetIssueFields("parent"); return err },
		"update issue": func(client GitHubClient) error { return client.MutateIssue(githubv4.UpdateIssueInput{ID: "child"}) },
		"add project": func(client GitHubClient) error {
			return client.MutateProject(githubv4.AddProjectV2ItemByIdInput{ProjectID: "project", ContentID: "child"})
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			for _, tc := range []struct {
				name, body, want string
				status           int
			}{
				{"GraphQL", `{"errors":[{"message":"fixture denied"}]}`, "fixture denied", http.StatusOK},
				{"HTTP unauthorized", `{"message":"Bad credentials"}`, "401", http.StatusUnauthorized},
			} {
				t.Run(tc.name, func(t *testing.T) {
					client := fixtureClient(t, "", tc.body, nil, tc.status)
					if err := call(client); err == nil || !strings.Contains(err.Error(), tc.want) {
						t.Fatalf("expected %s error, got %v", tc.name, err)
					}
				})
			}
		})
	}
}
