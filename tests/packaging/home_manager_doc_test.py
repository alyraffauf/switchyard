"""Evaluate the documented Home Manager setup without activating it."""

import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
HOME_MANAGER = os.environ.get("SWITCHYARD_TEST_HOME_MANAGER")
EVALUATION = """
let
  sw = builtins.getFlake (builtins.getEnv "SWITCHYARD_TEST_FLAKE");
  hm = builtins.getEnv "SWITCHYARD_TEST_HOME_MANAGER";
  result = import (hm + "/modules") {
    pkgs = sw.inputs.nixpkgs.legacyPackages.x86_64-linux;
    configuration = {
      imports = [ sw.homeManagerModules.switchyard ./home.nix ];
      home = {
        username = "switchyard-test";
        homeDirectory = "/tmp/switchyard-test";
        stateVersion = "26.05";
      };
    };
  };
in {
  hasMimeAppsFile = result.config.xdg.configFile ? "mimeapps.list";
  defaults = result.config.xdg.mimeApps.defaultApplications;
}
"""


class HomeManagerDocTest(unittest.TestCase):
    @unittest.skipUnless(HOME_MANAGER, "set SWITCHYARD_TEST_HOME_MANAGER to a local Home Manager source")
    def test_documented_setup_generates_mimeapps_file(self) -> None:
        docs = (ROOT / "website/src/content/docs/home-manager.md").read_text()
        snippet = re.search(r"```nix\n(# home\.nix\n.*?)\n```", docs, re.DOTALL)
        self.assertIsNotNone(snippet)
        assert snippet is not None
        with tempfile.TemporaryDirectory() as temp:
            tree = Path(temp)
            (tree / "home.nix").write_text(snippet.group(1))
            (tree / "eval.nix").write_text(EVALUATION)
            actual = json.loads(
                subprocess.check_output(
                    [
                        "nix", "--offline", "--option", "accept-flake-config", "false",
                        "eval", "--no-write-lock-file", "--impure", "--json", "--file", tree / "eval.nix",
                    ],
                    env=os.environ | {"SWITCHYARD_TEST_FLAKE": f"path:{ROOT}"},
                    text=True,
                    timeout=45,
                )
            )
        self.assertTrue(actual["hasMimeAppsFile"])
        self.assertEqual(
            actual["defaults"],
            {
                mime: ["io.github.alyraffauf.Switchyard.desktop"]
                for mime in ("x-scheme-handler/http", "x-scheme-handler/https", "text/html")
            },
        )


if __name__ == "__main__":
    unittest.main()
