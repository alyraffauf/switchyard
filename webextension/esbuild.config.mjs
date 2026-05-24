import * as esbuild from "esbuild";

const watch = process.argv.includes("--watch");
const ctx = await esbuild.context({
  entryPoints: ["src/App.tsx"],
  bundle: true,
  outdir: "build",
  format: "esm",
  target: "es2022",
  sourcemap: watch ? "inline" : false,
  minify: !watch,
});

if (watch) {
  await ctx.watch();
  console.log("Watching...");
} else {
  await ctx.rebuild();
  await ctx.dispose();
}
