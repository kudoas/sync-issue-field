"""Exercise the production launcher with fixture release assets, without GitHub writes."""

import hashlib
import os
import pathlib
import platform
import shutil
import struct
import subprocess
import tempfile
import unittest


ROOT = pathlib.Path(__file__).resolve().parents[1]
SHA = "a" * 40
TAG = "v0.3.0"


class DistributionTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="sync-issue-field-test-")
        self.addCleanup(self.temp.cleanup)
        self.directory = pathlib.Path(self.temp.name)
        self.assets = self.directory / "assets"
        self.assets.mkdir()
        self.shims = self.directory / "bin"
        self.shims.mkdir()
        self.runner_temp = self.directory / "runner"
        self.runner_temp.mkdir()
        for arch in ("amd64", "arm64"):
            (self.assets / f"sync-issue-field_linux_{arch}").write_text(
                '#!/bin/sh\nprintf "%s\\n" "$GITHUB_REPOSITORY:$GITHUB_ISSUE:$GITHUB_TOKEN"\n'
                'exit "${FIXTURE_EXIT:-0}"\n'
            )
        (self.assets / "source-commit.txt").write_text(SHA + "\n")
        self.checksums()
        self.shim("git", '#!/bin/sh\nif [ "$1" = -C ]; then\n'
                  '  printf "%s\\n" "$FIXTURE_LOCAL_SHA"\n'
                  'else\n  printf "%s\\n" "$FIXTURE_REFS"\nfi\n')
        self.shim("go", "#!/bin/sh\necho 'unexpected consumer build' >&2\nexit 99\n")
        self.shim("curl", """#!/usr/bin/env python3
import os, pathlib, shutil, sys
args = sys.argv[1:]
url = next(arg for arg in args if arg.startswith('https://'))
name = url.rsplit('/', 1)[1]
if url != f"https://github.com/kudoas/sync-issue-field/releases/download/{os.environ['FIXTURE_TAG']}/{name}":
    sys.exit(23)
source = pathlib.Path(os.environ['FIXTURE_ASSETS']) / name
if not source.exists():
    sys.exit(22)
shutil.copyfile(source, args[args.index('--output') + 1])
""")
        if not shutil.which("sha256sum"):
            self.shim("sha256sum", "#!/bin/sh\nexec shasum -a 256 \"$@\"\n")
        self.env = {
            **os.environ,
            "PATH": f"{self.shims}{os.pathsep}{os.environ['PATH']}",
            "RUNNER_OS": "Linux",
            "RUNNER_ARCH": "X64",
            "RUNNER_TEMP": str(self.runner_temp),
            "ACTION_REPOSITORY": "kudoas/sync-issue-field",
            "ACTION_REF": TAG,
            "ACTION_PATH": str(ROOT),
            "GITHUB_TOKEN": "fixture-token",
            "GITHUB_REPOSITORY": "example/consumer",
            "GITHUB_ISSUE": "42",
            "FIXTURE_REFS": f"{SHA}\trefs/tags/{TAG}",
            "FIXTURE_ASSETS": str(self.assets),
            "FIXTURE_TAG": TAG,
            "FIXTURE_LOCAL_SHA": SHA,
        }

    def shim(self, name, content):
        path = self.shims / name
        path.write_text(content)
        path.chmod(0o755)

    def checksums(self):
        lines = []
        for path in sorted(self.assets.iterdir()):
            if path.name != "SHA256SUMS":
                lines.append(f"{hashlib.sha256(path.read_bytes()).hexdigest()}  {path.name}\n")
        (self.assets / "SHA256SUMS").write_text("".join(lines))

    def launch(self, **changes):
        result = subprocess.run(
            ["bash", str(ROOT / "scripts/run-action.sh")],
            env={**self.env, **changes}, text=True, capture_output=True, timeout=10,
        )
        self.assertEqual(list(self.runner_temp.iterdir()), [], "temporary assets must be removed")
        return result

    def test_version_tag_executes_without_build_and_preserves_inputs(self):
        result = self.launch()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout, "example/consumer:42:fixture-token\n")

    def test_commit_pin_resolves_annotated_release_tag(self):
        refs = f"{'b' * 40}\trefs/tags/{TAG}\n{SHA}\trefs/tags/{TAG}^{{}}"
        result = self.launch(ACTION_REF=SHA, FIXTURE_REFS=refs)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_local_action_uses_its_checkout_commit(self):
        result = self.launch(ACTION_REF="", GITHUB_SHA="d" * 40)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_version_tag_resolves_annotated_tag_to_commit(self):
        refs = f"{'b' * 40}\trefs/tags/{TAG}\n{SHA}\trefs/tags/{TAG}^{{}}"
        result = self.launch(FIXTURE_REFS=refs)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_arm64_selects_arm64_asset(self):
        (self.assets / "sync-issue-field_linux_amd64").unlink()
        result = self.launch(RUNNER_ARCH="ARM64")
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_application_exit_status_is_preserved(self):
        self.assertEqual(self.launch(FIXTURE_EXIT="7").returncode, 7)

    def test_corrupt_binary_is_not_executed(self):
        (self.assets / "sync-issue-field_linux_amd64").write_text("corrupt")
        result = self.launch()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Checksum mismatch", result.stderr)
        self.assertEqual(result.stdout, "")

    def test_missing_duplicate_and_invalid_checksum_are_rejected(self):
        path = self.assets / "SHA256SUMS"
        original = path.read_text()
        for content in ("", original + original, "invalid  source-commit.txt\n"):
            with self.subTest(content=content):
                path.write_text(content)
                result = self.launch()
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("checksum", result.stderr)

    def test_wrong_source_commit_is_rejected_even_with_valid_checksums(self):
        (self.assets / "source-commit.txt").write_text("c" * 40 + "\n")
        self.checksums()
        result = self.launch()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("does not match", result.stderr)

    def test_missing_release_asset_fails_without_build_fallback(self):
        (self.assets / "sync-issue-field_linux_amd64").unlink()
        result = self.launch()
        self.assertEqual(result.returncode, 22)
        self.assertNotIn("unexpected consumer build", result.stderr)

    def test_unsupported_refs_and_runners_fail(self):
        for changes in (
            {"ACTION_REF": "main"}, {"ACTION_REF": "v0"},
            {"RUNNER_OS": "Windows"}, {"RUNNER_ARCH": "ARM"},
            {"ACTION_REPOSITORY": "../untrusted"}, {"FIXTURE_REFS": ""},
            {"ACTION_REF": "d" * 40},
        ):
            with self.subTest(changes=changes):
                self.assertNotEqual(self.launch(**changes).returncode, 0)

    @unittest.skipUnless(os.environ.get("RELEASE_ASSET_DIR"), "build release assets first")
    def test_packaged_binaries(self):
        directory = pathlib.Path(os.environ["RELEASE_ASSET_DIR"])
        commit = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT, text=True).strip()
        self.assertEqual((directory / "source-commit.txt").read_text().strip(), commit)
        expected = {"source-commit.txt", "sync-issue-field_linux_amd64", "sync-issue-field_linux_arm64"}
        lines = (directory / "SHA256SUMS").read_text().splitlines()
        self.assertEqual({line.split()[1] for line in lines}, expected)
        for line in lines:
            digest, name = line.split()
            self.assertEqual(hashlib.sha256((directory / name).read_bytes()).hexdigest(), digest)
        for arch, machine in (("amd64", 62), ("arm64", 183)):
            header = (directory / f"sync-issue-field_linux_{arch}").read_bytes()[:20]
            self.assertEqual(header[:6], b"\x7fELF\x02\x01")
            self.assertEqual(struct.unpack_from("<H", header, 18)[0], machine)
        # On Linux, run the actual packaged binary through the production launcher.
        # Invalid issue input stops before the GitHub API; no Issue is modified.
        if platform.system() == "Linux":
            for name in expected | {"SHA256SUMS"}:
                shutil.copyfile(directory / name, self.assets / name)
            arch = "ARM64" if platform.machine() in ("aarch64", "arm64") else "X64"
            result = self.launch(
                ACTION_REF=commit, FIXTURE_REFS=f"{commit}\trefs/tags/{TAG}",
                RUNNER_ARCH=arch, GITHUB_ISSUE="not-an-issue-number",
            )
            self.assertEqual(result.returncode, 1, result.stderr)
            self.assertIn("Issue number can't parse to int", result.stderr)


if __name__ == "__main__":
    unittest.main()
