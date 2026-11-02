package distribution

import (
	"crypto/sha256"
	"debug/elf"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestVersionTagExecutesWithoutBuildAndPreservesInputs(t *testing.T) {
	r := newFixture(t).launch(t, nil)
	requireCode(t, r, 0)
	if !strings.HasSuffix(r.stdout, "example/consumer:42:fixture-token\n") {
		t.Fatalf("inputs were not passed to executable: %s", r.stdout)
	}
}

func TestCommitPinResolvesAnnotatedReleaseTag(t *testing.T) {
	refs := strings.Repeat("b", 40) + "\trefs/tags/" + fixtureTag + "\n" + fixtureSHA + "\trefs/tags/" + fixtureTag + "^{}"
	r := newFixture(t).launch(t, map[string]string{"ACTION_REF": fixtureSHA, "FIXTURE_REFS": refs})
	requireCode(t, r, 0)
}

func TestPrereleaseReferences(t *testing.T) {
	tag := "v0.2.0-rc.1"
	for _, ref := range []string{tag, fixtureSHA} {
		t.Run(ref, func(t *testing.T) {
			requireCode(t, newFixture(t).launch(t, map[string]string{
				"ACTION_REF": ref, "FIXTURE_TAG": tag,
				"FIXTURE_REFS": fixtureSHA + "\trefs/tags/" + tag,
			}), 0)
		})
	}
}

func TestVersionTagResolvesAnnotatedTagToCommit(t *testing.T) {
	refs := strings.Repeat("b", 40) + "\trefs/tags/" + fixtureTag + "\n" + fixtureSHA + "\trefs/tags/" + fixtureTag + "^{}"
	requireCode(t, newFixture(t).launch(t, map[string]string{"FIXTURE_REFS": refs}), 0)
}

func TestLocalActionUsesItsCheckoutCommit(t *testing.T) {
	requireCode(t, newFixture(t).launch(t, map[string]string{"ACTION_REF": "", "GITHUB_SHA": strings.Repeat("d", 40)}), 0)
}

func TestARM64SelectsARM64Asset(t *testing.T) {
	f := newFixture(t)
	if err := os.Remove(filepath.Join(f.assets, "sync-issue-field_linux_amd64")); err != nil {
		t.Fatal(err)
	}
	requireCode(t, f.launch(t, map[string]string{"RUNNER_ARCH": "ARM64"}), 0)
}

func TestApplicationExitStatusIsPreserved(t *testing.T) {
	requireCode(t, newFixture(t).launch(t, map[string]string{"FIXTURE_EXIT": "7"}), 7)
}

func TestCorruptBinaryIsNotExecuted(t *testing.T) {
	f := newFixture(t)
	write(t, filepath.Join(f.assets, "sync-issue-field_linux_amd64"), "corrupt")
	r := f.launch(t, nil)
	if r.code == 0 || !strings.Contains(r.stderr, "Checksum mismatch") || strings.Contains(r.stdout, "fixture-token") {
		t.Fatalf("corrupt binary was not rejected: %+v", r)
	}
}

func TestMissingDuplicateAndInvalidChecksumAreRejected(t *testing.T) {
	for _, kind := range []string{"missing", "duplicate", "invalid"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			path := filepath.Join(f.assets, "SHA256SUMS")
			body := ""
			switch kind {
			case "duplicate":
				body = strings.Repeat(read(t, path), 2)
			case "invalid":
				body = "invalid  source-commit.txt\n"
			}
			write(t, path, body)
			r := f.launch(t, nil)
			if r.code == 0 || !strings.Contains(r.stderr, "checksum") {
				t.Fatalf("invalid checksum was not rejected: %+v", r)
			}
		})
	}
}

func TestWrongSourceCommitIsRejectedWithValidChecksums(t *testing.T) {
	f := newFixture(t)
	write(t, filepath.Join(f.assets, "source-commit.txt"), strings.Repeat("c", 40)+"\n")
	if err := f.checksums(); err != nil {
		t.Fatal(err)
	}
	r := f.launch(t, nil)
	if r.code == 0 || !strings.Contains(r.stderr, "does not match") {
		t.Fatalf("wrong source commit was not rejected: %+v", r)
	}
}

func TestMissingReleaseAssetFailsWithoutBuildFallback(t *testing.T) {
	f := newFixture(t)
	if err := os.Remove(filepath.Join(f.assets, "sync-issue-field_linux_amd64")); err != nil {
		t.Fatal(err)
	}
	r := f.launch(t, nil)
	requireCode(t, r, 22)
	if strings.Contains(r.stderr, "unexpected consumer build") {
		t.Fatal("launcher tried to build from source")
	}
}

func TestUnsupportedRefsAndRunnersFail(t *testing.T) {
	cases := map[string]map[string]string{
		"branch": {"ACTION_REF": "main"}, "alias": {"ACTION_REF": "v0"},
		"os": {"RUNNER_OS": "Windows"}, "arch": {"RUNNER_ARCH": "ARM"},
		"repository": {"ACTION_REPOSITORY": "../untrusted"}, "no tags": {"FIXTURE_REFS": ""},
		"unreleased commit": {"ACTION_REF": strings.Repeat("d", 40)},
	}
	for name, changes := range cases {
		t.Run(name, func(t *testing.T) {
			if r := newFixture(t).launch(t, changes); r.code == 0 {
				t.Fatalf("unsupported input accepted: %+v", r)
			}
		})
	}
}

func TestPackagedBinaries(t *testing.T) {
	directory := os.Getenv("RELEASE_ASSET_DIR")
	if directory == "" {
		t.Skip("build release assets and set RELEASE_ASSET_DIR first")
	}
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = rootDir()
	output, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	commit := strings.TrimSpace(string(output))
	if strings.TrimSpace(read(t, filepath.Join(directory, "source-commit.txt"))) != commit {
		t.Fatal("packaged source commit does not match HEAD")
	}
	expected := map[string]bool{"source-commit.txt": false, "sync-issue-field_linux_amd64": false, "sync-issue-field_linux_arm64": false}
	for _, line := range strings.Split(strings.TrimSpace(read(t, filepath.Join(directory, "SHA256SUMS"))), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatalf("invalid checksum line: %q", line)
		}
		digest, name := fields[0], fields[1]
		seen, exists := expected[name]
		if !exists || seen {
			t.Fatalf("unexpected or duplicate checksum: %s", name)
		}
		expected[name] = true
		if actual := fmt.Sprintf("%x", sha256.Sum256([]byte(read(t, filepath.Join(directory, name))))); actual != digest {
			t.Fatalf("checksum mismatch for %s", name)
		}
	}
	for name, seen := range expected {
		if !seen {
			t.Fatalf("missing checksum: %s", name)
		}
	}
	for arch, machine := range map[string]elf.Machine{"amd64": elf.EM_X86_64, "arm64": elf.EM_AARCH64} {
		file, err := elf.Open(filepath.Join(directory, "sync-issue-field_linux_"+arch))
		if err != nil {
			t.Fatal(err)
		}
		if file.Class != elf.ELFCLASS64 || file.Data != elf.ELFDATA2LSB || file.Machine != machine {
			t.Errorf("unexpected ELF architecture for %s: %+v", arch, file.FileHeader)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	// Exercise the actual release executable on Linux, stopping before any API call.
	if runtime.GOOS == "linux" {
		f := newFixture(t)
		for _, name := range []string{"source-commit.txt", "sync-issue-field_linux_amd64", "sync-issue-field_linux_arm64", "SHA256SUMS"} {
			if err := copyFile(filepath.Join(directory, name), filepath.Join(f.assets, name)); err != nil {
				t.Fatal(err)
			}
		}
		arch := "X64"
		if runtime.GOARCH == "arm64" {
			arch = "ARM64"
		}
		r := f.launch(t, map[string]string{
			"ACTION_REF": commit, "FIXTURE_REFS": commit + "\trefs/tags/" + fixtureTag,
			"RUNNER_ARCH": arch, "INPUT_ISSUE": "not-an-issue-number",
		})
		requireCode(t, r, 1)
		if !strings.Contains(r.stderr, "Issue number can't parse to int") {
			t.Fatalf("expected input validation before API access: %s", r.stderr)
		}
	}
}
