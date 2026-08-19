import { crisisRu, rejectRu } from "./copy";
import type { HexView, RoverType, RoverView, Snapshot } from "./types";

export function roversOf(game: Snapshot | null): RoverView[] {
  if (!game) return [];
  if (game.rovers?.length) return game.rovers;
  return game.rover ? [game.rover] : [];
}

export function batPct(r?: RoverView | null): number {
  if (!r) return 0;
  const max = r.maxBattery || 100;
  return Math.max(0, Math.min(100, (100 * (r.battery || 0)) / max));
}

export function roverStateRu(r?: RoverView | null): string {
  if (!r) return "ждёт";
  if (r.state === "moving") return "едет";
  if (r.state === "stranded") return "сел";
  return "ждёт";
}

export function hexDark(hex: HexView, t: { pos: number; direction: string }): boolean {
  if (t.direction === "west") return hex.q - 0.5 > t.pos;
  return hex.q + 0.5 < t.pos;
}

export function hexPhaseEta(hex: HexView, t: { pos: number; speed: number; direction: string }): number {
  if (!t.speed || hexDark(hex, t)) return 0;
  const dist = t.direction === "west" ? t.pos - (hex.q - 0.5) : hex.q + 0.5 - t.pos;
  return Math.max(0, dist / t.speed);
}

export function liveTerm(game: Snapshot, lastApply: number, now = performance.now()) {
  const t = game.map.terminator || { pos: 0, speed: 0, direction: "east" };
  let pos = t.pos;
  if (game.status !== "finished") {
    const extra = Math.max(0, (now - lastApply) / 1000);
    pos += (t.direction === "west" ? -1 : 1) * (t.speed || 0) * extra;
  }
  return { pos, speed: t.speed || 0, direction: t.direction };
}

export function coach(game: Snapshot | null, selected: RoverType, hover: HexView | null, lastApply: number): string {
  if (!game?.map) return "Подключение к смене…";
  if (game.reject?.reason) return rejectRu[game.reject.reason] || "Нельзя.";
  if (game.status === "finished") return "Смена окончена. Читай вердикт Black Box.";
  const list = roversOf(game);
  const r = list.find((x) => x.type === selected) || list[0];
  const queued = (game.contracts || []).filter((c) => c.status === "queued");
  const active = (game.contracts || []).filter((c) => c.status === "accepted" || c.status === "in_transit");
  const burning = queued.filter((c) => c.deadline > 0 && c.deadline < 20);
  if (r?.state === "stranded") return "Ровер сел — тень или батарея. Нажми второго сверху.";
  if (hover && hexDark(hover, liveTerm(game, lastApply))) return "Эта клетка уже тень. Ходить туда нельзя.";
  if (r?.state === "moving") {
    if (active.some((c) => c.status === "in_transit")) {
      return "Везёт на светлую базу. В тень нельзя — если сдача погасла, точка сама переедет.";
    }
    return "Едет. Новый клик не откатывает: доезжает текущую клетку, потом новый путь.";
  }
  const live = (game.salvage || []).filter((x) => x.status === "available");
  const need = game.goal?.colonyNeed || 100;
  if (!queued.length && !active.length && live.length) {
    return `Заказы кончились. Подбери бирюзовые кассеты — ещё +${live[0].value} колонии каждая, пока тень не забрала.`;
  }
  if ((game.colonyScore || 0) >= need) {
    if (live.length) return "Цель набрана. Кассеты ещё можно подобрать — тень не останавливается.";
    return "Цель набрана. Смена ещё идёт — тень не останавливается.";
  }
  if (burning.length && !active.length) return "СЕЙЧАС: бери горящий заказ (срок на карточке). Остальное тень сожрёт.";
  if (!active.length) return "СЕЙЧАС: нажми пульсирующую карточку. Кассеты — запасные очки. Цель — 100 колонии.";
  return "Ровер на месте. Нажми П-цифру или другую карточку. На золотой базе батарея растёт.";
}

export function crisisText(game: Snapshot | null): string {
  if (!game?.crisis?.fired || game.status === "finished") return "";
  return crisisRu[game.crisis.kind] || "Кризис.";
}
