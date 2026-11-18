"""Exercise the just install and uninstall recipes without installing to the system."""

import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
APP_ID = "io.github.alyraffauf.Switchyard"


class StagedInstallTest(unittest.TestCase):
    def test_staging_paths_with_spaces(self) -> None:
        with tempfile.TemporaryDirectory(prefix="switchyard-staging-") as temp:
            tree = Path(temp)
            shutil.copy2(ROOT / "justfile", tree)
            shutil.copytree(ROOT / "data", tree / "data")
            go = tree / "go"
            go.write_text("#!/bin/sh\nprintf 'stub binary\\n' > switchyard\n")
            go.chmod(0o755)
            stage = tree / "package root"
            prefix = "/opt/Browser Tools"
            target = stage / prefix.lstrip("/")
            env = os.environ | {
                "PATH": f"{tree}:{os.environ['PATH']}",
                "DESTDIR": str(stage),
                "PREFIX": prefix,
            }
            expected = {
                target / "bin/switchyard",
                target / f"share/applications/{APP_ID}.desktop",
                target / f"share/metainfo/{APP_ID}.metainfo.xml",
                target / f"share/icons/hicolor/scalable/apps/{APP_ID}.svg",
            }

            subprocess.run(["just", "install"], cwd=tree, env=env, check=True)
            self.assertEqual({file for file in stage.rglob("*") if file.is_file()}, expected)

            unrelated = target / "bin/another-app"
            unrelated.touch()
            subprocess.run(["just", "uninstall"], cwd=tree, env=env, check=True)
            self.assertEqual({file for file in stage.rglob("*") if file.is_file()}, {unrelated})


if __name__ == "__main__":
    unittest.main()
