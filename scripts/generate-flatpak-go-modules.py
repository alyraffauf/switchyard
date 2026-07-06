#!/usr/bin/env python3
import hashlib
import json
import shlex
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "flatpak" / "go-modules.json"


def go_escape(value: str) -> str:
    escaped = []
    for char in value:
        if "A" <= char <= "Z":
            escaped.append("!" + char.lower())
        else:
            escaped.append(char)
    return "".join(escaped)


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as file:
        for chunk in iter(lambda: file.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def go_mod_download() -> list[dict[str, str]]:
    proc = subprocess.run(
        ["go", "mod", "download", "-json", "all"],
        cwd=ROOT,
        check=True,
        text=True,
        stdout=subprocess.PIPE,
    )

    modules = []
    decoder = json.JSONDecoder()
    output = proc.stdout.lstrip()

    while output:
        module, index = decoder.raw_decode(output)
        modules.append(module)
        output = output[index:].lstrip()

    return modules


def source_entry(module: dict[str, str], suffix: str) -> dict[str, str]:
    path = module["Path"]
    version = module["Version"]
    escaped_path = go_escape(path)
    escaped_version = go_escape(version)
    cache_dir = f"go/pkg/mod/cache/download/{escaped_path}/@v"
    filename = f"{escaped_version}.{suffix}"

    return {
        "type": "file",
        "url": f"https://proxy.golang.org/{escaped_path}/@v/{filename}",
        "sha256": sha256(Path(module[{"info": "Info", "mod": "GoMod", "zip": "Zip"}[suffix]])),
        "dest": cache_dir,
        "dest-filename": filename,
    }


def cache_setup_command(module: dict[str, str]) -> str:
    # proxy.golang.org .info files carry a timestamp and are not byte-stable, so
    # pinning their sha256 eventually drifts and breaks the build. go does not
    # checksum .info, so synthesize a minimal one locally instead of downloading
    # it. .ziphash is written here too since go.sum already vouches for it.
    path = module["Path"]
    version = module["Version"]
    escaped_path = go_escape(path)
    escaped_version = go_escape(version)
    cache_dir = f"go/pkg/mod/cache/download/{escaped_path}/@v"
    ziphash = f"{cache_dir}/{escaped_version}.ziphash"
    info = f"{cache_dir}/{escaped_version}.info"
    info_json = json.dumps({"Version": version}, separators=(",", ":"))

    return (
        f"mkdir -p {shlex.quote(cache_dir)} && "
        f"printf '%s\\n' {shlex.quote(module['Sum'])} > {shlex.quote(ziphash)} && "
        f"printf '%s' {shlex.quote(info_json)} > {shlex.quote(info)}"
    )


def main() -> int:
    sources = []
    modules = sorted(go_mod_download(), key=lambda m: (m["Path"], m["Version"]))
    for module in modules:
        if "Error" in module:
            print(module["Error"], file=sys.stderr)
            return 1
        sources.extend(source_entry(module, suffix) for suffix in ("mod", "zip"))

    sources.append(
        {
            "type": "shell",
            "commands": [cache_setup_command(module) for module in modules],
        }
    )

    OUT.write_text(json.dumps(sources, indent=2) + "\n")
    print(f"Wrote {OUT.relative_to(ROOT)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
