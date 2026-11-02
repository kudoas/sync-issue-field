"""Stage an offline release for the local composite Action smoke test in CI."""

import os
import pathlib
import shutil

from test_distribution import DistributionTest, TAG


fixture = DistributionTest()
fixture.setUp()
try:
    destination = pathlib.Path(os.environ["RUNNER_TEMP"]) / "composite-fixture"
    shutil.copytree(fixture.directory, destination)
    fixture.assets = destination / "assets"
    commit = os.environ["GITHUB_SHA"]
    (fixture.assets / "source-commit.txt").write_text(commit + "\n")
    for arch in ("amd64", "arm64"):
        (fixture.assets / f"sync-issue-field_linux_{arch}").write_text(
            '#!/bin/sh\nset -eu\n'
            'test "$GITHUB_REPOSITORY" = "example/consumer"\n'
            'test "$GITHUB_ISSUE" = "133"\n'
            'test "$GITHUB_TOKEN" = "fixture-token"\n'
            'echo executed > "$FIXTURE_RESULT"\n'
        )
    fixture.checksums()
    values = {
        "FIXTURE_REFS": f"{commit}\trefs/tags/{TAG}",
        "FIXTURE_LOCAL_SHA": commit,
        "FIXTURE_TAG": TAG,
        "FIXTURE_ASSETS": str(fixture.assets),
        "FIXTURE_RESULT": str(destination / "result"),
    }
    with open(os.environ["GITHUB_ENV"], "a") as output:
        for name, value in values.items():
            output.write(f"{name}={value}\n")
    with open(os.environ["GITHUB_PATH"], "a") as output:
        output.write(f"{destination / 'bin'}\n")
finally:
    fixture.doCleanups()
