#!/usr/bin/env node
import { chromium } from "playwright-core";
import { mkdir } from "node:fs/promises";
import { resolve } from "node:path";

const origin = process.env.SHOT_ORIGIN || "http://127.0.0.1";
const outDir = resolve(process.env.SHOT_DIR || "screenshots");
const chrome = process.env.CHROME || "/usr/bin/chromium-browser";

async function createGame() {
  const res = await fetch(`${origin}/api/v1/games`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ rover: "swift", seed: "MCC-TEST" }),
  });
  if (!res.ok) throw new Error(`create game ${res.status}`);
  return res.json();
}

async function waitPlayable(page) {
  await page.waitForFunction(() => {
    const canvas = document.querySelector("canvas");
    return Boolean(canvas && canvas.clientWidth > 80 && canvas.clientHeight > 80);
  }, { timeout: 20000 });
  await page.waitForTimeout(1200);
}

async function main() {
  await mkdir(outDir, { recursive: true });
  const game = await createGame();
  const id = game.id;
  if (!id) throw new Error("no game id");

  const browser = await chromium.launch({
    executablePath: chrome,
    args: ["--no-sandbox", "--disable-dev-shm-usage", "--hide-scrollbars"],
  });

  const desktop = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    deviceScaleFactor: 1,
    locale: "ru-RU",
  });
  const mobile = await browser.newContext({
    viewport: { width: 390, height: 844 },
    deviceScaleFactor: 2,
    isMobile: true,
    hasTouch: true,
    locale: "ru-RU",
    userAgent:
      "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
  });

  const tutPage = await desktop.newPage();
  await tutPage.addInitScript(() => {
    localStorage.removeItem("phaseline_tutorial_done");
    sessionStorage.removeItem("phaseline_tutorial_done");
  });
  await tutPage.goto(`${origin}/play/${id}`, { waitUntil: "networkidle" });
  await waitPlayable(tutPage);
  await tutPage.getByRole("button", { name: "ДАЛЬШЕ" }).waitFor({ timeout: 10000 });
  await tutPage.screenshot({ path: `${outDir}/tutorial.png`, type: "png" });

  const deskPage = await desktop.newPage();
  await deskPage.goto(`${origin}/play/${id}?shot=1`, { waitUntil: "networkidle" });
  await waitPlayable(deskPage);
  await deskPage.getByText("Колония", { exact: false }).waitFor();
  await deskPage.screenshot({ path: `${outDir}/desktop.png`, type: "png" });

  const mobPage = await mobile.newPage();
  await mobPage.goto(`${origin}/play/${id}?shot=1`, { waitUntil: "networkidle" });
  await waitPlayable(mobPage);
  await mobPage.getByText("Колония", { exact: false }).waitFor();
  await mobPage.screenshot({ path: `${outDir}/mobile.png`, type: "png" });

  await browser.close();
  console.log("wrote", outDir, "game", id, game.seed);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
