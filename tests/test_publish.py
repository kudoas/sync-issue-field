import os
import pathlib
import subprocess
import tempfile
import unittest


ROOT = pathlib.Path(__file__).resolve().parents[1]


class PublishTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="sync-issue-publish-")
        self.addCleanup(self.temp.cleanup)
        self.directory = pathlib.Path(self.temp.name)
        self.assets = self.directory / "assets"
        self.remote = self.directory / "remote"
        self.bin = self.directory / "bin"
        for directory in (self.assets, self.remote, self.bin):
            directory.mkdir()
        for name in ("SHA256SUMS", "source-commit.txt", "sync-issue-field_linux_amd64", "sync-issue-field_linux_arm64"):
            (self.assets / name).write_text(name)
        self.log = self.directory / "uploads"
        shim = self.bin / "gh"
        shim.write_text("""#!/usr/bin/env python3
import os, pathlib, shutil, sys
args = sys.argv[1:]
remote = pathlib.Path(os.environ['FIXTURE_REMOTE'])
if args[:2] == ['release', 'view']:
    print('\\n'.join(sorted(path.name for path in remote.iterdir())))
elif args[:2] == ['release', 'download']:
    name = args[args.index('--pattern') + 1]
    shutil.copyfile(remote / name, pathlib.Path(args[args.index('--dir') + 1]) / name)
elif args[:2] == ['release', 'upload']:
    assert '--clobber' not in args
    source = pathlib.Path(args[-1])
    assert not (remote / source.name).exists()
    shutil.copyfile(source, remote / source.name)
    with open(os.environ['FIXTURE_UPLOADS'], 'a') as output:
        output.write(source.name + '\\n')
else:
    sys.exit(2)
""")
        shim.chmod(0o755)

    def publish(self):
        return subprocess.run(
            ["bash", str(ROOT / "scripts/publish-release.sh"), "v0.3.0", str(self.assets)],
            env={**os.environ, "PATH": f"{self.bin}{os.pathsep}{os.environ['PATH']}",
                 "GITHUB_REPOSITORY": "example/action", "FIXTURE_REMOTE": str(self.remote),
                 "FIXTURE_UPLOADS": str(self.log)},
            text=True, capture_output=True, timeout=10,
        )

    def test_uploads_all_assets(self):
        result = self.publish()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(sorted(p.name for p in self.assets.iterdir()), sorted(self.log.read_text().splitlines()))

    def test_recovers_partial_upload_without_overwriting(self):
        name = "source-commit.txt"
        (self.remote / name).write_bytes((self.assets / name).read_bytes())
        result = self.publish()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertNotIn(name, self.log.read_text().splitlines())

    def test_complete_retry_does_not_upload(self):
        self.assertEqual(self.publish().returncode, 0)
        self.log.unlink()
        result = self.publish()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertFalse(self.log.exists())

    def test_conflicting_asset_fails_before_any_upload(self):
        (self.remote / "source-commit.txt").write_text("different")
        result = self.publish()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("differs", result.stderr)
        self.assertFalse(self.log.exists())
