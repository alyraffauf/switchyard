import {
  siFirefox,
  siGooglechrome,
  siBrave,
  siOpera,
  siOperagx,
  siVivaldi,
  siSafari,
  siLibrewolf,
  siTorbrowser,
  siArc,
  siZenbrowser,
  siMullvad,
} from "simple-icons";

const ICON_MAP: Record<string, typeof siFirefox> = {
  "opera-gx": siOperagx, // before "opera" so substring match hits this first
  firefox: siFirefox,
  chrome: siGooglechrome,
  chromium: siGooglechrome,
  brave: siBrave,
  opera: siOpera,
  vivaldi: siVivaldi,
  librewolf: siLibrewolf,
  tor: siTorbrowser,
  arc: siArc,
  zen: siZenbrowser,
  mullvad: siMullvad,
  safari: siSafari,
};

export function BrowserIcon({ browserId }: { browserId: string }) {
  const lower = browserId.toLowerCase();
  const match = Object.entries(ICON_MAP).find(([key]) => lower.includes(key));
  const [, ic] = match ?? [];
  return (
    <span style={{ flexShrink: 0, width: 16, height: 16, marginRight: 8 }}>
      {ic && (
        <svg
          viewBox="0 0 24 24"
          width="16"
          height="16"
          fill={`#${ic.hex}`}
          aria-hidden="true"
        >
          <path d={ic.path} />
        </svg>
      )}
    </span>
  );
}
