import { Resvg } from "@resvg/resvg-js";
import { existsSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import satori from "satori";

const __dirname = dirname(fileURLToPath(import.meta.url));
const rootDir = join(__dirname, "..");

const colors = {
  primary: "#6ab377",
  bg: "#0f1410",
  bgAlt: "#151c17",
  text: "#e8ebe9",
  textMuted: "#a0a8a2",
};

async function loadFont(url) {
  const response = await fetch(url);
  if (!response.ok)
    throw new Error(
      `Failed to fetch font: ${response.status} ${response.statusText}`,
    );
  return response.arrayBuffer();
}

function element(type, props, ...children) {
  return { type, props: { ...props, children } };
}

async function generateOgImage() {
  const outputPath = join(rootDir, "public/og-image.png");
  if (existsSync(outputPath)) {
    console.log("Skipping og-image.png generation (already exists)");
    return;
  }

  const [fontRegular, fontBold] = await Promise.all([
    loadFont(
      "https://fonts.gstatic.com/s/inter/v18/UcCO3FwrK3iLTeHuS_nVMrMxCp50SjIw2boKoduKmMEVuLyfAZ9hjp-Ek-_EeA.woff",
    ),
    loadFont(
      "https://fonts.gstatic.com/s/inter/v18/UcCO3FwrK3iLTeHuS_nVMrMxCp50SjIw2boKoduKmMEVuGKYAZ9hjp-Ek-_EeA.woff",
    ),
  ]);

  const iconSrc = `data:image/svg+xml,${encodeURIComponent(
    readFileSync(join(rootDir, "public/icon.svg"), "utf-8"),
  )}`;

  const svg = await satori(
    element(
      "div",
      {
        style: {
          height: "100%",
          width: "100%",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          background: `linear-gradient(135deg, ${colors.bg} 0%, ${colors.bgAlt} 100%)`,
          fontFamily: "Inter",
        },
      },
      element(
        "div",
        {
          style: {
            display: "flex",
            alignItems: "center",
            gap: "24px",
            marginBottom: "24px",
          },
        },
        element("img", { src: iconSrc, width: 100, height: 100 }),
        element(
          "span",
          {
            style: {
              fontSize: 80,
              fontWeight: 700,
              color: colors.text,
              letterSpacing: "-0.02em",
            },
          },
          "Switchyard",
        ),
      ),
      element(
        "div",
        {
          style: {
            display: "flex",
            fontSize: 36,
            fontWeight: 500,
            color: colors.textMuted,
            marginBottom: "32px",
          },
        },
        "A rules-based browser launcher for Linux.",
      ),
      element(
        "div",
        {
          style: {
            display: "flex",
            padding: "12px 28px",
            background: colors.primary,
            color: "#ffffff",
            borderRadius: "12px",
            fontSize: 24,
            fontWeight: 600,
          },
        },
        "Download Now",
      ),
    ),
    {
      width: 1200,
      height: 630,
      fonts: [
        { name: "Inter", data: fontRegular, weight: 400, style: "normal" },
        { name: "Inter", data: fontBold, weight: 700, style: "normal" },
      ],
    },
  );

  const resvg = new Resvg(svg, { fitTo: { mode: "width", value: 1200 } });
  const pngBuffer = resvg.render().asPng();

  writeFileSync(outputPath, pngBuffer);
  console.log("Generated public/og-image.png");
}

generateOgImage().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
