import { chromium } from "playwright";
import { mkdir } from "node:fs/promises";

const baseURL = process.env.UI_BASE_URL ?? "http://127.0.0.1:4173";
const outputDir = process.env.SCREEN_OUTPUT ?? "/home/ubuntu/work/ollama-dz23-work/docs/images/screens";

await mkdir(outputDir, { recursive: true });
const browser = await chromium.launch({ headless: true });
const page = await browser.newPage({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 1 });

for (const route of ["/projects", "/library", "/scheduled", "/skills", "/plugins", "/tasks", "/agentic", "/settings"]) {
  const name = route.slice(1) || "home";
  await page.goto(`${baseURL}${route}`, { waitUntil: "networkidle" });
  await page.screenshot({ path: `${outputDir}/class-a-plus-${name}.png`, fullPage: true });
  console.log(`${route} title=${await page.title()} h1=${await page.locator("h1").first().textContent().catch(() => "")}`);
}

await browser.close();
