import type { FooterLink } from "../types";
import { DONATE_URL, REPO_URL } from "./urls";

export const footerLinks: FooterLink[] = [
  { href: REPO_URL, label: "GitHub" },
  { href: `${REPO_URL}/issues`, label: "Issues" },
  { href: DONATE_URL, label: "Donate" },
];
