package distribution

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func publishFixture(t *testing.T) *fixture {
	t.Helper()
	f := newFixture(t)
	remote := filepath.Join(f.dir, "remote")
	if err := os.Mkdir(remote, 0700); err != nil {
		t.Fatal(err)
	}
	f.env["FIXTURE_REMOTE"] = remote
	f.env["FIXTURE_UPLOADS"] = filepath.Join(f.dir, "uploads")
	f.env["GITHUB_REPOSITORY"] = "example/action"
	return f
}

func (f *fixture) publish(t *testing.T) result {
	t.Helper()
	return f.run(t, "publish-release.sh", nil, fixtureTag, f.assets)
}

func TestUploadsAllAssets(t *testing.T) {
	f := publishFixture(t)
	requireCode(t, f.publish(t), 0)
	entries, err := os.ReadDir(f.assets)
	if err != nil {
		t.Fatal(err)
	}
	var expected []string
	for _, entry := range entries {
		expected = append(expected, entry.Name())
	}
	actual := strings.Fields(read(t, f.env["FIXTURE_UPLOADS"]))
	sort.Strings(actual)
	if strings.Join(expected, "\n") != strings.Join(actual, "\n") {
		t.Fatalf("uploaded %v, want %v", actual, expected)
	}
}

func TestRecoversPartialUploadWithoutOverwriting(t *testing.T) {
	f := publishFixture(t)
	name := "source-commit.txt"
	if err := copyFile(filepath.Join(f.assets, name), filepath.Join(f.env["FIXTURE_REMOTE"], name)); err != nil {
		t.Fatal(err)
	}
	requireCode(t, f.publish(t), 0)
	for _, uploaded := range strings.Fields(read(t, f.env["FIXTURE_UPLOADS"])) {
		if uploaded == name {
			t.Fatal("existing asset was uploaded again")
		}
	}
}

func TestCompleteRetryDoesNotUpload(t *testing.T) {
	f := publishFixture(t)
	requireCode(t, f.publish(t), 0)
	if err := os.Remove(f.env["FIXTURE_UPLOADS"]); err != nil {
		t.Fatal(err)
	}
	requireCode(t, f.publish(t), 0)
	if _, err := os.Stat(f.env["FIXTURE_UPLOADS"]); !os.IsNotExist(err) {
		t.Fatalf("unexpected upload on retry: %v", err)
	}
}

func TestConflictingAssetFailsBeforeAnyUpload(t *testing.T) {
	f := publishFixture(t)
	write(t, filepath.Join(f.env["FIXTURE_REMOTE"], "source-commit.txt"), "different")
	r := f.publish(t)
	if r.code == 0 || !strings.Contains(r.stderr, "differs") {
		t.Fatalf("conflicting asset not rejected: %+v", r)
	}
	if _, err := os.Stat(f.env["FIXTURE_UPLOADS"]); !os.IsNotExist(err) {
		t.Fatalf("uploaded before checking conflicts: %v", err)
	}
}
