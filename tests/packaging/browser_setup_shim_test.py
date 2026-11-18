"""Check that the browser setup shell entrypoint keeps confirmation input available."""

import errno
import os
from pathlib import Path
import pty
import select
import shutil
import subprocess
import sys
import tempfile
import time
import unittest

ROOT = Path(__file__).resolve().parents[2]
SHIM = ROOT / "scripts/install-desktop-integration.sh"
PROMPT = b"Grant permission to these browsers? [y/N] "


class BrowserSetupShimTest(unittest.TestCase):
    def test_confirm_with_redirected_input(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            env = self.stubs(Path(temp))
            result = subprocess.run(
                ["bash", SHIM, "--install"],
                env=env, input="y\n", text=True, capture_output=True, timeout=10,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual((Path(temp) / "overrides").read_text(), "override\n")

    def test_decline_when_the_shell_reads_its_script_from_stdin(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            env = self.stubs(Path(temp))
            bash = shutil.which("bash")
            assert bash is not None
            pid, terminal = pty.fork()
            if pid == 0:
                os.execve(bash, [bash, "-c", '"$0" -s -- --install < "$1"', bash, str(SHIM)], env)
            transcript = b""
            answered = False
            deadline = time.monotonic() + 10
            try:
                while not answered:
                    readable, _, _ = select.select([terminal], [], [], max(0, deadline - time.monotonic()))
                    if not readable:
                        self.fail(f"confirmation prompt timed out: {transcript!r}")
                    try:
                        output = os.read(terminal, 4096)
                    except OSError as error:
                        # On Linux a closed pty slave reports EIO to the reader.
                        if error.errno != errno.EIO:
                            raise
                        break
                    if not output:
                        break
                    transcript += output
                    if PROMPT in transcript:
                        os.write(terminal, b"n\n")
                        answered = True
                _, status = os.waitpid(pid, 0)
                pid = 0
                self.assertEqual(os.waitstatus_to_exitcode(status), 0, transcript.decode(errors="replace"))
                self.assertTrue(answered)
                self.assertFalse((Path(temp) / "overrides").exists())
            finally:
                os.close(terminal)
                if pid:
                    os.waitpid(pid, 0)

    def stubs(self, home: Path) -> dict[str, str]:
        scripts = {
            "curl": """exec python3 -c 'import os; print(open(os.environ["SETUP_SOURCE"]).read())'""",
            "flatpak": """case "$1:$2" in
                info:org.mozilla.firefox) exit 0 ;;
                override:--user) printf 'override\\n' >> "$HOME/overrides" ;;
                *) exit 1 ;;
            esac""",
            "switchyard": "exit 0",
        }
        for name, body in scripts.items():
            path = home / name
            path.write_text("#!/bin/sh\n" + body + "\n")
            path.chmod(0o755)
        (home / "python3").symlink_to(sys.executable)
        return os.environ | {
            "HOME": str(home),
            "PATH": f"{home}{os.pathsep}{os.environ['PATH']}",
            "SETUP_SOURCE": str(ROOT / "browser-setup/main.py"),
        }


if __name__ == "__main__":
    unittest.main()
