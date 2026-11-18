import { expect, test } from "bun:test";
import { spawnSync } from "node:child_process";
import {
  cpSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  readdirSync,
  rmSync,
  symlinkSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

test("a Markdown guide in a subdirectory gets a linked docs page", () => {
  const website = join(import.meta.dir, "..");
  const scratch = mkdtempSync(join(tmpdir(), "switchyard-nested-doc-"));
  try {
    for (const name of [
      "astro.config.mjs",
      "package.json",
      "public",
      "src",
      "tsconfig.json",
    ]) {
      cpSync(join(website, name), join(scratch, name), { recursive: true });
    }
    const installed = join(website, "node_modules");
    const modules = join(scratch, "node_modules");
    mkdirSync(modules);
    for (const name of readdirSync(installed)) {
      if (name === ".astro" || name === ".vite") continue;
      symlinkSync(join(installed, name), join(modules, name));
    }
    const guides = join(scratch, "src/content/docs/guides");
    mkdirSync(guides);
    writeFileSync(
      join(guides, "weekend.md"),
      "---\ntitle: Weekend setup\n---\n\nA nested guide.\n",
    );

    const build = spawnSync(join(modules, ".bin/astro"), ["build"], {
      cwd: scratch,
      encoding: "utf8",
      env: { ...process.env, ASTRO_TELEMETRY_DISABLED: "1" },
    });
    expect(build.status, build.stdout + build.stderr).toBe(0);
    const page = readFileSync(
      join(scratch, "dist/docs/guides/weekend/index.html"),
      "utf8",
    );
    expect(page).toContain("A nested guide.");
    const index = readFileSync(join(scratch, "dist/docs/index.html"), "utf8");
    expect(index).toContain('href="/docs/guides/weekend/"');
  } finally {
    rmSync(scratch, { recursive: true, force: true });
  }
}, 30_000);
