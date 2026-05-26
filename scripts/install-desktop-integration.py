#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
# Switchyard desktop integration installer — https://github.com/alyraffauf/Switchyard

import argparse
import json
import shutil
import subprocess
import sys
from dataclasses import asdict, dataclass
from pathlib import Path

FLATPAK_CHROMIUM: dict[str, str] = {
    "com.brave.Browser": "BraveSoftware/Brave-Browser",
    "com.google.Chrome": "google-chrome",
    "com.microsoft.Edge": "microsoft-edge",
    "com.vivaldi.Vivaldi": "vivaldi",
    "io.github.ungoogled_software.ungoogled_chromium": "chromium",
    "net.imput.helium": "net.imput.helium",
    "org.chromium.Chromium": "chromium",
}

NATIVE_CHROMIUM = [
    "net.imput.helium",
    "google-chrome",
    "chromium",
    "BraveSoftware/Brave-Browser",
    "microsoft-edge",
    "vivaldi",
]


FLATPAK_FIREFOX: dict[str, str] = {
    "org.mozilla.firefox": ".mozilla",
    "io.gitlab.librewolf-community": ".librewolf",
}

NATIVE_FIREFOX = [".mozilla", ".librewolf"]


@dataclass
class NativeMessagingManifest:
    path: str
    name: str = "io.github.alyraffauf.switchyard"
    description: str = "Switchyard browser selector"
    type: str = "stdio"
    allowed_origins: list[str] | None = None
    allowed_extensions: list[str] | None = None


def run(cmd: list[str]) -> bool:
    return subprocess.run(cmd, capture_output=True).returncode == 0


def flatpak_installed(flatpak: str | None, id: str) -> bool:
    if flatpak:
        return run([flatpak, "info", id])
    return False


def find_switchyard(flatpak: str | None) -> str | None:
    switchyard = shutil.which("switchyard")

    if flatpak and flatpak_installed(flatpak, "io.github.alyraffauf.Switchyard"):
        return f"{flatpak} run io.github.alyraffauf.Switchyard"

    return switchyard


def write_wrapper(switchyard: str, path: Path):
    wrapper = f"""#!/bin/sh
if [ "${{container-}}" = flatpak ]; then
    exec /usr/bin/flatpak-spawn --host {switchyard} --native-host "$@"
else
    exec {switchyard} --native-host "$@"
fi
"""

    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(wrapper)
    path.chmod(0o755)


def install_manifests(switchyard: str, configs: list[Path], hosts_dir: str, **kwargs):
    for config_dir in configs:
        manifest_file = config_dir / hosts_dir / "io.github.alyraffauf.switchyard.json"
        manifest_file.parent.mkdir(parents=True, exist_ok=True)
        manifest = NativeMessagingManifest(
            path=str(manifest_file.parent / "switchyard-native-host-wrapper.sh"),
            **kwargs,
        )
        manifest_file.write_text(
            json.dumps(
                {key: val for key, val in asdict(manifest).items() if val is not None}
            )
        )
        print(f"  Installed {manifest_file}")
        write_wrapper(
            switchyard, manifest_file.parent / "switchyard-native-host-wrapper.sh"
        )


def install(yes: bool = False):
    print("Installing Switchyard native messaging host...")

    flatpak = shutil.which("flatpak")
    switchyard = find_switchyard(flatpak)

    if switchyard is None:
        print("Switchyard is not installed.")
        sys.exit(1)

    installed_chromium = [
        app_id for app_id in FLATPAK_CHROMIUM if flatpak_installed(flatpak, app_id)
    ]
    installed_firefox = [
        app_id for app_id in FLATPAK_FIREFOX if flatpak_installed(flatpak, app_id)
    ]
    to_override = installed_chromium + installed_firefox

    chromium = [
        Path.home() / f".var/app/{app_id}/config/{FLATPAK_CHROMIUM[app_id]}"
        for app_id in installed_chromium
    ] + [Path.home() / f".config/{config_dir}" for config_dir in NATIVE_CHROMIUM]

    firefox = [
        Path.home() / f".var/app/{app_id}/{FLATPAK_FIREFOX[app_id]}"
        for app_id in installed_firefox
    ] + [Path.home() / profile_dir for profile_dir in NATIVE_FIREFOX]

    install_manifests(
        switchyard,
        chromium,
        "NativeMessagingHosts",
        allowed_origins=[
            "chrome-extension://ncehhpikkabfdcceimdhjjjodogflokc/",
            "chrome-extension://gmdmmjfmpfbmddgphjbkbbmdolkifloi/",
        ],
    )

    install_manifests(
        switchyard,
        firefox,
        "native-messaging-hosts",
        allowed_extensions=["switchyard@alyraffauf.github.io"],
    )

    if to_override:
        print(
            "\nThe following browsers are installed via Flatpak and run in a "
            "sandbox. To let them talk to Switchyard, we need to grant each one "
            "permission to reach outside the sandbox. This does weaken their "
            "isolation slightly, so it's worth knowing before we proceed.\n"
        )
        for app_id in to_override:
            print(f"  {app_id}")
        print()

        if yes:
            answer = "y"
        else:
            answer = input("Grant permission to these browsers? [y/N] ").strip().lower()

        if answer != "y":
            print("Skipping permission grants.")
            to_override = []

        assert flatpak
        for app_id in to_override:
            granted = run(
                [
                    flatpak,
                    "override",
                    "--user",
                    "--talk-name=org.freedesktop.Flatpak",
                    app_id,
                ]
            )
            if granted:
                print(f"  Granted: {app_id}")
            else:
                print(f"  Failed to grant permission for {app_id}")

    print("\nManifests installed. Please restart your browser.")


def uninstall():
    chromium = [
        Path.home() / f".var/app/{app_id}/config/{FLATPAK_CHROMIUM[app_id]}"
        for app_id in FLATPAK_CHROMIUM
    ] + [Path.home() / f".config/{config_dir}" for config_dir in NATIVE_CHROMIUM]

    firefox = [
        Path.home() / f".var/app/{app_id}/{FLATPAK_FIREFOX[app_id]}"
        for app_id in FLATPAK_FIREFOX
    ] + [Path.home() / profile_dir for profile_dir in NATIVE_FIREFOX]

    count = 0
    for config_dir in chromium:
        hosts = config_dir / "NativeMessagingHosts"
        for file in (
            hosts / "io.github.alyraffauf.switchyard.json",
            hosts / "switchyard-native-host-wrapper.sh",
        ):
            if file.exists():
                file.unlink()
                print(f"  Removed {file}")
                count += 1
    for profile_dir in firefox:
        hosts = profile_dir / "native-messaging-hosts"
        for file in (
            hosts / "io.github.alyraffauf.switchyard.json",
            hosts / "switchyard-native-host-wrapper.sh",
        ):
            if file.exists():
                file.unlink()
                print(f"  Removed {file}")
                count += 1
    print(f"\nRemoved {count} file(s).")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        prog="install-desktop-integration.py",
        description="Set up desktop integration between the Switchyard app and browser extension.",
    )

    group = parser.add_mutually_exclusive_group()
    group.add_argument(
        "--install", action="store_true", help="install the native messaging manifest."
    )

    parser.add_argument(
        "--yes", "-y", action="store_true", help="skip confirmation prompts."
    )

    group.add_argument(
        "--uninstall",
        action="store_true",
        help="remove the native messaging host manifest.",
    )

    args = parser.parse_args()

    if not args.install and not args.uninstall:
        parser.print_help()
        sys.exit(0)

    if args.install:
        install(yes=args.yes)

    if args.uninstall:
        uninstall()
