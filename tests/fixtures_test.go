package distribution

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

const fixtureSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const fixtureTag = "v0.3.0"

// The test executable doubles as offline git/curl/gh helpers. The production
// shell scripts run unchanged; no network access or GitHub writes are needed.
func TestMain(m *testing.M) {
	if os.Getenv("PREPARE_COMPOSITE_FIXTURE") == "1" {
		if err := prepareComposite(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if os.Getenv("DISTRIBUTION_HELPER") == "1" {
		os.Exit(runHelper(filepath.Base(os.Args[0]), os.Args[1:]))
	}
	os.Exit(m.Run())
}

func runHelper(name string, args []string) int {
	switch name {
	case "git":
		if len(args) > 0 && args[0] == "-C" {
			fmt.Println(os.Getenv("FIXTURE_LOCAL_SHA"))
		} else {
			fmt.Println(os.Getenv("FIXTURE_REFS"))
		}
		return 0
	case "go":
		fmt.Fprintln(os.Stderr, "unexpected consumer build")
		return 99
	case "curl":
		url := ""
		for _, arg := range args {
			if strings.HasPrefix(arg, "https://") {
				url = arg
			}
		}
		name := filepath.Base(url)
		if url != "https://github.com/kudoas/sync-issue-field/releases/download/"+os.Getenv("FIXTURE_TAG")+"/"+name {
			return 23
		}
		if err := copyFile(filepath.Join(os.Getenv("FIXTURE_ASSETS"), name), flagValue(args, "--output")); err != nil {
			return 22
		}
		return 0
	case "gh":
		if len(args) < 2 || args[0] != "release" {
			return 2
		}
		remote := os.Getenv("FIXTURE_REMOTE")
		switch args[1] {
		case "view":
			entries, err := os.ReadDir(remote)
			if err != nil {
				return 2
			}
			for _, entry := range entries {
				fmt.Println(entry.Name())
			}
		case "download":
			name := flagValue(args, "--pattern")
			if err := copyFile(filepath.Join(remote, name), filepath.Join(flagValue(args, "--dir"), name)); err != nil {
				return 2
			}
		case "upload":
			for _, arg := range args {
				if arg == "--clobber" {
					return 2
				}
			}
			source := args[len(args)-1]
			name := filepath.Base(source)
			if name == os.Getenv("FIXTURE_FAIL_UPLOAD") {
				return 2
			}
			if _, err := os.Stat(filepath.Join(remote, name)); !os.IsNotExist(err) {
				return 2
			}
			if err := copyFile(source, filepath.Join(remote, name)); err != nil {
				return 2
			}
			log, err := os.OpenFile(os.Getenv("FIXTURE_UPLOADS"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				return 2
			}
			_, writeErr := fmt.Fprintln(log, name)
			closeErr := log.Close()
			if writeErr != nil || closeErr != nil {
				return 2
			}
		case "edit":
			// Publishing before every asset exists is a failure, even if a later
			// upload would otherwise fill the missing files.
			for _, name := range []string{"SHA256SUMS", "source-commit.txt", "sync-issue-field_linux_amd64", "sync-issue-field_linux_arm64"} {
				if _, err := os.Stat(filepath.Join(remote, name)); err != nil {
					return 2
				}
			}
			if err := os.WriteFile(os.Getenv("FIXTURE_PUBLICATION"), []byte(strings.Join(args, "\n")), 0600); err != nil {
				return 2
			}
		default:
			return 2
		}
		return 0
	}
	return 2
}

func flagValue(args []string, flag string) string {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag {
			return args[i+1]
		}
	}
	return ""
}

func copyFile(source, destination string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, data, 0600)
}

func rootDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(file))
}

type fixture struct {
	dir, assets, bin, runner string
	env                      map[string]string
}

func makeFixture(dir string) (*fixture, error) {
	f := &fixture{dir: dir, assets: filepath.Join(dir, "assets"), bin: filepath.Join(dir, "bin"), runner: filepath.Join(dir, "runner")}
	for _, path := range []string{f.assets, f.bin, f.runner} {
		if err := os.MkdirAll(path, 0700); err != nil {
			return nil, err
		}
	}
	for _, arch := range []string{"amd64", "arm64"} {
		body := "#!/bin/sh\nprintf '%s\\n' \"$GITHUB_REPOSITORY:$GITHUB_ISSUE:$GITHUB_TOKEN\"\nexit \"${FIXTURE_EXIT:-0}\"\n"
		if err := os.WriteFile(filepath.Join(f.assets, "sync-issue-field_linux_"+arch), []byte(body), 0600); err != nil {
			return nil, err
		}
	}
	if err := os.WriteFile(filepath.Join(f.assets, "source-commit.txt"), []byte(fixtureSHA+"\n"), 0600); err != nil {
		return nil, err
	}
	if err := f.checksums(); err != nil {
		return nil, err
	}
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	for _, name := range []string{"git", "curl", "gh", "go"} {
		if err := os.Symlink(executable, filepath.Join(f.bin, name)); err != nil {
			return nil, err
		}
	}
	f.env = map[string]string{
		"PATH":                f.bin + string(os.PathListSeparator) + os.Getenv("PATH"),
		"DISTRIBUTION_HELPER": "1", "RUNNER_OS": "Linux", "RUNNER_ARCH": "X64", "RUNNER_TEMP": f.runner,
		"ACTION_REPOSITORY": "kudoas/sync-issue-field", "ACTION_REF": fixtureTag, "ACTION_PATH": rootDir(),
		"INPUT_TOKEN": "fixture-token", "INPUT_REPOSITORY": "example/consumer", "INPUT_ISSUE": "42",
		"GITHUB_TOKEN": "caller-token", "GITHUB_REPOSITORY": "original/caller", "GITHUB_ISSUE": "999",
		"FIXTURE_REFS": fixtureSHA + "\trefs/tags/" + fixtureTag, "FIXTURE_LOCAL_SHA": fixtureSHA,
		"FIXTURE_ASSETS": f.assets, "FIXTURE_TAG": fixtureTag,
	}
	return f, nil
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f, err := makeFixture(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *fixture) checksums() error {
	entries, err := os.ReadDir(f.assets)
	if err != nil {
		return err
	}
	var sums strings.Builder
	for _, entry := range entries {
		if entry.Name() == "SHA256SUMS" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(f.assets, entry.Name()))
		if err != nil {
			return err
		}
		fmt.Fprintf(&sums, "%x  %s\n", sha256.Sum256(data), entry.Name())
	}
	return os.WriteFile(filepath.Join(f.assets, "SHA256SUMS"), []byte(sums.String()), 0600)
}

func mergedEnv(overrides ...map[string]string) []string {
	values := map[string]string{}
	for _, entry := range os.Environ() {
		key, value, _ := strings.Cut(entry, "=")
		values[key] = value
	}
	for _, override := range overrides {
		for key, value := range override {
			values[key] = value
		}
	}
	var env []string
	for key, value := range values {
		env = append(env, key+"="+value)
	}
	sort.Strings(env)
	return env
}

type result struct {
	code           int
	stdout, stderr string
}

func (f *fixture) run(t *testing.T, script string, changes map[string]string, args ...string) result {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", append([]string{filepath.Join(rootDir(), "scripts", script)}, args...)...)
	cmd.Env = mergedEnv(f.env, changes)
	cmd.Dir = rootDir()
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok && ctx.Err() == nil {
			code = exit.ExitCode()
		} else {
			t.Fatal(err)
		}
	}
	return result{code, stdout.String(), stderr.String()}
}

func (f *fixture) launch(t *testing.T, changes map[string]string) result {
	t.Helper()
	r := f.run(t, "run-action.sh", changes)
	entries, err := os.ReadDir(f.runner)
	if err != nil || len(entries) != 0 {
		t.Fatalf("temporary assets not removed: %v %v", entries, err)
	}
	return r
}

func requireCode(t *testing.T, r result, code int) {
	t.Helper()
	if r.code != code {
		t.Fatalf("exit = %d, want %d\nstdout: %s\nstderr: %s", r.code, code, r.stdout, r.stderr)
	}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// Called only by CI's compiled test executable, before `uses: ./`.
func prepareComposite() error {
	f, err := makeFixture(filepath.Join(os.Getenv("RUNNER_TEMP"), "composite-fixture"))
	if err != nil {
		return err
	}
	commit := os.Getenv("GITHUB_SHA")
	if err := os.WriteFile(filepath.Join(f.assets, "source-commit.txt"), []byte(commit+"\n"), 0600); err != nil {
		return err
	}
	for _, arch := range []string{"amd64", "arm64"} {
		body := `#!/bin/sh
set -eu
test "$GITHUB_REPOSITORY" = example/consumer || { echo 'repository input mismatch' >&2; exit 1; }
test "$GITHUB_ISSUE" = 133 || { echo 'issue input mismatch' >&2; exit 1; }
test "$GITHUB_TOKEN" = fixture-token || { echo 'token input mismatch' >&2; exit 1; }
echo executed > "$FIXTURE_RESULT"
`
		if err := os.WriteFile(filepath.Join(f.assets, "sync-issue-field_linux_"+arch), []byte(body), 0600); err != nil {
			return err
		}
	}
	if err := f.checksums(); err != nil {
		return err
	}
	values := map[string]string{
		"DISTRIBUTION_HELPER": "1", "FIXTURE_REFS": commit + "\trefs/tags/" + fixtureTag,
		"FIXTURE_LOCAL_SHA": commit, "FIXTURE_TAG": fixtureTag, "FIXTURE_ASSETS": f.assets,
		"FIXTURE_RESULT": filepath.Join(f.dir, "result"),
	}
	var env strings.Builder
	for name, value := range values {
		fmt.Fprintf(&env, "%s=%s\n", name, value)
	}
	for path, content := range map[string]string{os.Getenv("GITHUB_ENV"): env.String(), os.Getenv("GITHUB_PATH"): f.bin + "\n"} {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		_, writeErr := file.WriteString(content)
		closeErr := file.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
