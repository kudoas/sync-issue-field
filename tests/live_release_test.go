package distribution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

type liveIssue struct {
	ID        int64                    `json:"id"`
	Number    int                      `json:"number"`
	URL       string                   `json:"html_url"`
	Labels    []struct{ Name string }  `json:"labels"`
	Assignees []struct{ Login string } `json:"assignees"`
	Milestone *struct{ Number int }    `json:"milestone"`
}

type liveState struct {
	Label        string    `json:"label"`
	Milestone    int       `json:"milestone"`
	Parent       liveIssue `json:"parent"`
	Child        liveIssue `json:"child"`
	LabelCreated bool      `json:"label_created"`
}

func liveAPI(method, path string, input, output interface{}) error {
	var body io.Reader
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, "https://api.github.com"+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GITHUB_TOKEN"))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("%s %s: HTTP %d", method, path, response.StatusCode)
	}
	if output != nil {
		return json.NewDecoder(response.Body).Decode(output)
	}
	return nil
}

func liveFields(issue liveIssue) map[string]interface{} {
	labels, assignees := []string{}, []string{}
	for _, label := range issue.Labels {
		labels = append(labels, label.Name)
	}
	for _, assignee := range issue.Assignees {
		assignees = append(assignees, assignee.Login)
	}
	sort.Strings(labels)
	sort.Strings(assignees)
	milestone := 0
	if issue.Milestone != nil {
		milestone = issue.Milestone.Number
	}
	return map[string]interface{}{"labels": labels, "assignees": assignees, "milestone": milestone}
}

// Opt-in live test. All mutable resources are created for this run and cleaned
// up separately even when the Action or the readback assertion fails.
func TestLiveIssueSync(t *testing.T) {
	if os.Getenv("LIVE_RELEASE_TEST") != "1" {
		t.Skip("live verification is opt-in")
	}
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo != "kudoas/sync-issue-field" || os.Getenv("RUNNER_TEMP") == "" {
		t.Fatal("unexpected verification repository or missing state path")
	}
	path := filepath.Join(os.Getenv("RUNNER_TEMP"), "live-fixture.json")
	prefix := "/repos/" + repo
	var state liveState
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &state); err != nil {
			t.Fatal(err)
		}
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	save := func() {
		t.Helper()
		data, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	call := func(method, suffix string, input, output interface{}) {
		t.Helper()
		if err := liveAPI(method, prefix+suffix, input, output); err != nil {
			t.Fatal(err)
		}
	}
	issuePath := func(number int) string { return fmt.Sprintf("/issues/%d", number) }
	switch os.Getenv("LIVE_PHASE") {
	case "setup":
		state.Label = fmt.Sprintf("rc-%s-%s-%s-%s", os.Getenv("GITHUB_RUN_ID"), os.Getenv("GITHUB_RUN_ATTEMPT"), os.Getenv("LIVE_RELATION"), os.Getenv("LIVE_REFERENCE"))
		save()
		call("POST", "/labels", map[string]interface{}{"name": state.Label, "color": "cccccc"}, nil)
		state.LabelCreated = true
		save()
		var milestone struct{ Number int }
		call("POST", "/milestones", map[string]interface{}{"title": state.Label}, &milestone)
		state.Milestone = milestone.Number
		save()
		body := "Temporary v0.2.0-rc.1 verification fixture. Closed automatically after the run."
		call("POST", "/issues", map[string]interface{}{"title": "[RC verification] Child " + state.Label, "body": body}, &state.Child)
		save()
		call("POST", "/issues", map[string]interface{}{"title": "[RC verification] Parent " + state.Label, "body": body, "labels": []string{state.Label}, "assignees": []string{"kudoas"}, "milestone": state.Milestone}, &state.Parent)
		save()
		if os.Getenv("LIVE_RELATION") == "sub-issue" {
			call("POST", issuePath(state.Parent.Number)+"/sub_issues", map[string]interface{}{"sub_issue_id": state.Child.ID}, nil)
			var parent liveIssue
			call("GET", issuePath(state.Child.Number)+"/parent", nil, &parent)
			if parent.Number != state.Parent.Number {
				t.Fatal("native parent relation was not created")
			}
		} else {
			call("PATCH", issuePath(state.Parent.Number), map[string]interface{}{"body": body + "\n\n- [ ] " + state.Child.URL}, nil)
		}
		// Clear GitHub's automatic milestone inheritance to isolate the Action.
		call("PATCH", issuePath(state.Child.Number), map[string]interface{}{"milestone": nil, "labels": []string{}, "assignees": []string{}}, nil)
		call("GET", issuePath(state.Parent.Number), nil, &state.Parent)
		call("GET", issuePath(state.Child.Number), nil, &state.Child)
		save()
		if reflect.DeepEqual(liveFields(state.Parent), liveFields(state.Child)) {
			t.Fatal("child already matches parent before running the Action")
		}
		// Record both tracking mechanisms; only checklist tracking is asynchronous.
		var tracking struct {
			Errors []struct{ Message string }
			Data   struct {
				Repository struct {
					Issue struct {
						TrackedInIssues struct{ Nodes []struct{ ID string } }
						Parent          *struct{ ID string }
					}
				}
			}
		}
		query := fmt.Sprintf(`{repository(owner:"kudoas",name:"sync-issue-field"){issue(number:%d){parent{id} trackedInIssues(first:5){nodes{id}}}}}`, state.Child.Number)
		for attempt := 0; attempt < 10; attempt++ {
			if err := liveAPI("POST", "/graphql", map[string]interface{}{"query": query}, &tracking); err != nil {
				t.Fatal(err)
			}
			if len(tracking.Errors) > 0 {
				t.Fatalf("tracking query failed: %v", tracking.Errors)
			}
			if os.Getenv("LIVE_RELATION") == "sub-issue" || len(tracking.Data.Repository.Issue.TrackedInIssues.Nodes) > 0 {
				break
			}
			time.Sleep(2 * time.Second)
		}
		t.Logf("parent=%s child=%s native_parent=%v tracked_parents=%d before=%v expected=%v", state.Parent.URL, state.Child.URL, tracking.Data.Repository.Issue.Parent != nil, len(tracking.Data.Repository.Issue.TrackedInIssues.Nodes), liveFields(state.Child), liveFields(state.Parent))
		if os.Getenv("LIVE_RELATION") == "checklist" && len(tracking.Data.Repository.Issue.TrackedInIssues.Nodes) == 0 {
			t.Fatal("checklist tracking was not established")
		}
		output, err := os.OpenFile(os.Getenv("GITHUB_OUTPUT"), os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		_, err = fmt.Fprintf(output, "child=%d\nparent=%d\n", state.Child.Number, state.Parent.Number)
		if closeErr := output.Close(); err != nil || closeErr != nil {
			t.Fatalf("write output: %v %v", err, closeErr)
		}
	case "verify":
		if state.Child.Number == 0 || state.Parent.Number == 0 {
			t.Fatal("missing test fixture")
		}
		var actual liveIssue
		call("GET", issuePath(state.Child.Number), nil, &actual)
		expected, observed := liveFields(state.Parent), liveFields(actual)
		data, err := json.MarshalIndent(map[string]interface{}{"relation": os.Getenv("LIVE_RELATION"), "reference": os.Getenv("LIVE_REFERENCE"), "before": liveFields(state.Child), "expected": expected, "actual": observed, "parent": state.Parent.URL, "child": state.Child.URL}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(filepath.Dir(path), "live-result.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
		t.Logf("expected=%v actual=%v", expected, observed)
		if !reflect.DeepEqual(expected, observed) {
			t.Fatal("Action did not synchronize the expected fields")
		}
	case "cleanup":
		var failures []string
		for _, number := range []int{state.Child.Number, state.Parent.Number} {
			if number == 0 {
				continue
			}
			if err := liveAPI("PATCH", prefix+issuePath(number), map[string]interface{}{"state": "closed"}, nil); err != nil {
				failures = append(failures, err.Error())
			}
		}
		if state.Milestone != 0 {
			if err := liveAPI("DELETE", fmt.Sprintf("%s/milestones/%d", prefix, state.Milestone), nil, nil); err != nil {
				failures = append(failures, err.Error())
			}
		}
		if state.LabelCreated {
			if err := liveAPI("DELETE", prefix+"/labels/"+state.Label, nil, nil); err != nil {
				failures = append(failures, err.Error())
			}
		}
		if len(failures) > 0 {
			t.Fatal(strings.Join(failures, "\n"))
		}
		t.Log("test Issues closed and temporary label/milestone deleted")
	default:
		t.Fatal("unknown live verification phase")
	}
}
