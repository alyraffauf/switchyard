import * as esbuild from "esbuild";

const watch = process.argv.includes("--watch");
const ctx = await esbuild.context({
  entryPoints: ["src/App.tsx", "src/background.ts"],
  bundle: true,
  outdir: "build",
  format: "esm",
  target: "es2022",
  sourcemap: watch ? "inline" : false,
  minify: !watch,
  define: watch ? {} : { "process.env.NODE_ENV": '"production"' },
});

if (watch) {
  await ctx.watch();
  console.log("Watching...");
} else {
  await ctx.rebuild();
  await ctx.dispose();
}
