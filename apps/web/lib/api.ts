import type { BlackBox, RoverType, Snapshot } from "./types";

const base = process.env.NEXT_PUBLIC_API_BASE ?? "";

async function parse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    throw new Error(`http ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export async function waitHealth(tries = 50): Promise<boolean> {
  for (let i = 0; i < tries; i++) {
    try {
      if ((await fetch(`${base}/health`, { cache: "no-store" })).ok) return true;
    } catch {
      /* retry */
    }
    await new Promise((r) => setTimeout(r, 400));
  }
  return false;
}

export async function createGame(seed?: string, rover: RoverType = "swift"): Promise<Snapshot> {
  const res = await fetch(`${base}/api/v1/games`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ rover, seed }),
  });
  return parse<Snapshot>(res);
}

export async function getGame(id: string): Promise<Snapshot> {
  return parse<Snapshot>(await fetch(`${base}/api/v1/games/${id}`, { cache: "no-store" }));
}

export async function getBlackBox(id: string): Promise<BlackBox> {
  return parse<BlackBox>(await fetch(`${base}/api/v1/games/${id}/blackbox`, { cache: "no-store" }));
}

export type SeedPreview = {
  seed: string;
  layout?: string;
  map: { hexCount: number; terminator: { pos: number; speed: number; direction: string } };
};

export async function getSeedPreview(seed: string): Promise<SeedPreview> {
  return parse<SeedPreview>(await fetch(`${base}/api/v1/seeds/${encodeURIComponent(seed)}`, { cache: "no-store" }));
}

export function gameWsUrl(id: string): string {
  const proto = typeof location !== "undefined" && location.protocol === "https:" ? "wss" : "ws";
  const host = typeof location !== "undefined" ? location.host : "localhost";
  const prefix = base.startsWith("http") ? base.replace(/^http/, proto === "wss" ? "https" : "ws") : `${proto}://${host}`;
  return `${prefix}/ws/game/${id}`;
}
