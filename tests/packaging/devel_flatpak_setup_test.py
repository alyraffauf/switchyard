"""Check that browser setup recognizes the Flatpak shipped by the build workflow."""

import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
HOST = "io.github.alyraffauf.switchyard"
APP = "io.github.alyraffauf.Switchyard"


class DevelFlatpakSetupTest(unittest.TestCase):
    def test_uses_devel_flatpak_if_it_is_the_only_installed_build(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            flatpak = home / "flatpak"
            flatpak.write_text(
                f"""#!/bin/sh
case "$1:$2" in
    info:{APP}.Devel) exit 0 ;;
    run:{APP}.Devel) printf '%s\\n' "$@" ;;
    *) exit 1 ;;
esac
"""
            )
            flatpak.chmod(0o755)
            env = {"HOME": str(home), "PATH": str(home)}
            installer = subprocess.run(
                [sys.executable, ROOT / "browser-setup/main.py", "--install", "--yes"],
                env=env,
                capture_output=True,
                text=True,
                timeout=10,
            )
            self.assertEqual(
                installer.returncode, 0, installer.stdout + installer.stderr
            )
            manifest = home / f".mozilla/native-messaging-hosts/{HOST}.json"
            wrapper = json.loads(manifest.read_text())["path"]
            result = subprocess.check_output(
                [wrapper, "ordinary arg"], env=env, text=True, timeout=10
            )
            self.assertEqual(
                result.splitlines(),
                ["run", f"{APP}.Devel", "--native-host", "ordinary arg"],
            )


if __name__ == "__main__":
    unittest.main()
