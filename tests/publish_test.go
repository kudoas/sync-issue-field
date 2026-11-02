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
	f.env["FIXTURE_PUBLICATION"] = filepath.Join(f.dir, "publication")
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
	if _, err := os.Stat(f.env["FIXTURE_PUBLICATION"]); !os.IsNotExist(err) {
		t.Fatalf("published despite conflicting asset: %v", err)
	}
}

func TestPublicationFlags(t *testing.T) {
	for _, tc := range []struct {
		tag, prerelease, latest string
	}{
		{"v0.2.0-rc.1", "true", "false"},
		{"v0.2.0", "false", "true"},
	} {
		t.Run(tc.tag, func(t *testing.T) {
			f := publishFixture(t)
			r := f.run(t, "publish-release.sh", nil, tc.tag, f.assets)
			requireCode(t, r, 0)
			publication := read(t, f.env["FIXTURE_PUBLICATION"])
			for _, want := range []string{tc.tag, "--draft=false", "--prerelease=" + tc.prerelease, "--latest=" + tc.latest} {
				if !strings.Contains(publication, want) {
					t.Errorf("publication %q does not contain %q", publication, want)
				}
			}
		})
	}
}

func TestFailedUploadKeepsReleaseUnpublishedAndCanRetry(t *testing.T) {
	f := publishFixture(t)
	f.env["FIXTURE_FAIL_UPLOAD"] = "sync-issue-field_linux_arm64"
	if r := f.publish(t); r.code == 0 {
		t.Fatal("expected upload failure")
	}
	if _, err := os.Stat(f.env["FIXTURE_PUBLICATION"]); !os.IsNotExist(err) {
		t.Fatalf("published despite upload failure: %v", err)
	}
	delete(f.env, "FIXTURE_FAIL_UPLOAD")
	requireCode(t, f.publish(t), 0)
	if _, err := os.Stat(f.env["FIXTURE_PUBLICATION"]); err != nil {
		t.Fatal("retry did not publish the completed release:", err)
	}
}

func TestMissingAssetPreventsPublication(t *testing.T) {
	f := publishFixture(t)
	if err := os.Remove(filepath.Join(f.assets, "sync-issue-field_linux_arm64")); err != nil {
		t.Fatal(err)
	}
	if r := f.publish(t); r.code == 0 {
		t.Fatal("incomplete asset set was accepted")
	}
	for _, path := range []string{f.env["FIXTURE_UPLOADS"], f.env["FIXTURE_PUBLICATION"]} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("incomplete assets caused a remote mutation: %s: %v", path, err)
		}
	}
}
