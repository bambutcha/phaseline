import { hexColors, terrainRu } from "./copy";
import { hexToPixel, parseHexId } from "./hex";
import { hexDark, hexPhaseEta, liveTerm, roversOf } from "./coach";
import type { HexView, RoverView, RoverType, Snapshot } from "./types";

type Pt = { x: number; y: number };
type View = { ox: number; oy: number; size: number };
type Dust = { x: number; y: number; a: number; s: number };
type Flash = { x: number; y: number; r: number; a: number };

export class GameRenderer {
  vis: Record<string, { x: number; y: number; ang: number }> = {};
  dust: Dust[] = [];
  flashes: Flash[] = [];
  holdProg: Record<string, { p: number; hex: string; next: string }> = {};
  view: View = { ox: 0, oy: 0, size: 40 };
  clock = 0;
  hover: HexView | null = null;
  selected: RoverType = "swift";
  startedAt = 0;
  lastApply = 0;

  reset() {
    this.vis = {};
    this.dust = [];
    this.flashes = [];
    this.holdProg = {};
    this.startedAt = performance.now();
  }

  flash(x: number, y: number) {
    this.flashes.push({ x, y, r: 8, a: 1 });
  }

  roverPixel(game: Snapshot, r: RoverView, offset = 0): Pt & { ang: number } {
    return this.smoothPixel(game, r, offset);
  }

  computeView(game: Snapshot, w: number, h: number): View {
    let minX = Infinity,
      maxX = -Infinity,
      minY = Infinity,
      maxY = -Infinity;
    for (const hex of game.map.hexes) {
      const p = hexToPixel(hex.q, hex.r, 1);
      minX = Math.min(minX, p.x);
      maxX = Math.max(maxX, p.x);
      minY = Math.min(minY, p.y);
      maxY = Math.max(maxY, p.y);
    }
    const pad = 2.6;
    const size = Math.min((w * 0.9) / (maxX - minX + pad * 2 || 1), (h * 0.84) / (maxY - minY + pad * 2 || 1));
    const cx = (minX + maxX) / 2,
      cy = (minY + maxY) / 2;
    return { size: Math.max(size, 22), ox: w / 2 - cx * size, oy: h / 2 - cy * size + 8 };
  }

  world(hex: { q: number; r: number }): Pt {
    const p = hexToPixel(hex.q, hex.r, this.view.size);
    return { x: this.view.ox + p.x, y: this.view.oy + p.y };
  }

  liveProgress(r: RoverView): number {
    const hex = `${r.q},${r.r}`;
    const next = (r.path && r.path[0]) || "";
    let p = r.progress || 0;
    const prev = this.holdProg[r.type];
    if (prev && prev.hex === hex && prev.next === next && r.state === "moving" && prev.p > 0.15 && p + 0.2 < prev.p) {
      p = prev.p;
    } else {
      this.holdProg[r.type] = { p, hex, next };
    }
    if (r.state === "moving" && !r.reversing && r.path?.length) {
      const extra = Math.max(0, (performance.now() - this.lastApply) / 1000);
      p = Math.min(0.985, p + extra * (r.type === "hauler" ? 0.34 : 0.5));
    }
    return p;
  }

  targetPixel(game: Snapshot, r: RoverView, offset: number) {
    const cur = this.world({ q: r.q, r: r.r });
    const progress = this.liveProgress(r);
    let nxt: Pt | null = null;
    if (r.reversing && r.reverseHex) nxt = this.world(parseHexId(r.reverseHex));
    else if (r.path?.length) nxt = this.world(parseHexId(r.path[0]));
    let p = cur;
    if (nxt && progress > 0) p = { x: cur.x + (nxt.x - cur.x) * progress, y: cur.y + (nxt.y - cur.y) * progress };
    return {
      x: p.x + offset,
      y: p.y,
      ang: nxt ? Math.atan2(nxt.y - cur.y, nxt.x - cur.x) : -Math.PI / 2,
    };
  }

  smoothPixel(game: Snapshot, r: RoverView, offset: number) {
    const t = this.targetPixel(game, r, offset);
    if (!this.vis[r.type]) this.vis[r.type] = { x: t.x, y: t.y, ang: t.ang };
    const v = this.vis[r.type];
    v.x += (t.x - v.x) * 0.55;
    v.y += (t.y - v.y) * 0.55;
    let da = t.ang - v.ang;
    while (da > Math.PI) da -= Math.PI * 2;
    while (da < -Math.PI) da += Math.PI * 2;
    v.ang += da * 0.35;
    return v;
  }

  pickHex(game: Snapshot, x: number, y: number): HexView | null {
    let best: HexView | null = null,
      bestD = 1e9;
    for (const hex of game.map.hexes) {
      const p = this.world(hex);
      const d = (p.x - x) ** 2 + (p.y - y) ** 2;
      if (d < bestD) {
        bestD = d;
        best = hex;
      }
    }
    return bestD < this.view.size * this.view.size * 1.5 ? best : null;
  }

  pickRover(game: Snapshot, x: number, y: number): RoverView | null {
    for (const r of roversOf(game)) {
      const p = this.vis[r.type] || this.targetPixel(game, r, r.type === "hauler" ? 8 : -8);
      if ((p.x - x) ** 2 + (p.y - y) ** 2 < this.view.size * this.view.size * 0.4) return r;
    }
    return null;
  }

  frame(ctx: CanvasRenderingContext2D, game: Snapshot | null, w: number, h: number, ts: number) {
    this.clock = ts / 1000;
    ctx.clearRect(0, 0, w, h);
    this.drawSpace(ctx, w, h);
    if (!game?.map) return;
    this.view = this.computeView(game, w, h);
    const term = liveTerm(game, this.lastApply);
    const act = roversOf(game).find((x) => x.type === this.selected) || roversOf(game)[0];
    const pathSet = new Set(act?.path || []);
    for (const hex of game.map.hexes) {
      const p = this.world(hex);
      const dark = hexDark(hex, term);
      let fill = hexColors[hex.type] || "#666";
      if (dark) fill = "#0a1016";
      const hover = this.hover && this.hover.id === hex.id;
      const pulse = hex.type === "base" && !dark ? `rgba(224,192,90,${0.35 + Math.sin(this.clock * 3) * 0.2})` : null;
      this.drawHex(ctx, p.x, p.y, fill, dark ? "#1a1010" : pathSet.has(hex.id) ? "#fff" : hover ? "#e0c05a" : "#2a333c", pulse, !dark);
      if (dark) {
        ctx.strokeStyle = "rgba(224,104,104,0.35)";
        ctx.beginPath();
        ctx.moveTo(p.x - this.view.size * 0.22, p.y - this.view.size * 0.22);
        ctx.lineTo(p.x + this.view.size * 0.22, p.y + this.view.size * 0.22);
        ctx.stroke();
      }
    }
    this.drawShadowBand(ctx, game, w, h);
    this.drawGhost(ctx, game);
    this.drawPaths(ctx, game);
    this.drawPins(ctx, game);
    this.drawSalvage(ctx, game);
    this.drawRovers(ctx, game);
    this.drawFx(ctx);
    if (this.hover) this.drawHover(ctx, game, this.hover);
    if (this.startedAt && performance.now() - this.startedAt < 18000 && !roversOf(game).some((rv) => rv.state === "moving")) {
      ctx.textAlign = "center";
      ctx.font = "800 11px sans-serif";
      for (const hex of game.map.hexes) {
        if (hex.type !== "base") continue;
        const p = this.world(hex);
        ctx.fillStyle = "rgba(224,192,90,0.95)";
        ctx.fillText("БАЗА · заряд", p.x, p.y + 6);
      }
    }
  }

  private hexPath(ctx: CanvasRenderingContext2D, x: number, y: number, size = this.view.size) {
    ctx.beginPath();
    for (let i = 0; i < 6; i++) {
      const a = (Math.PI / 180) * (60 * i - 30);
      const px = x + size * Math.cos(a),
        py = y + size * Math.sin(a);
      i === 0 ? ctx.moveTo(px, py) : ctx.lineTo(px, py);
    }
    ctx.closePath();
  }

  private drawHex(
    ctx: CanvasRenderingContext2D,
    x: number,
    y: number,
    fill: string,
    stroke: string,
    glow: string | null,
    lit: boolean,
  ) {
    if (lit) {
      ctx.fillStyle = "rgba(0,0,0,0.35)";
      this.hexPath(ctx, x, y + 5);
      ctx.fill();
    }
    this.hexPath(ctx, x, y);
    if (glow) {
      ctx.shadowColor = glow;
      ctx.shadowBlur = 16;
    }
    ctx.fillStyle = fill;
    ctx.fill();
    ctx.shadowBlur = 0;
    if (lit) {
      const g = ctx.createLinearGradient(x, y - this.view.size, x, y + this.view.size);
      g.addColorStop(0, "rgba(255,255,255,0.16)");
      g.addColorStop(0.45, "rgba(255,255,255,0)");
      g.addColorStop(1, "rgba(0,0,0,0.18)");
      ctx.fillStyle = g;
      ctx.fill();
    }
    ctx.strokeStyle = stroke || "#1c242d";
    ctx.lineWidth = 2;
    ctx.stroke();
  }

  private drawShadowBand(ctx: CanvasRenderingContext2D, game: Snapshot, w: number, h: number) {
    const t = liveTerm(game, this.lastApply);
    const rs = game.map.hexes.map((hex) => hex.r);
    const r0 = Math.min(...rs) - 1.15,
      r1 = Math.max(...rs) + 1.15;
    const pts: Pt[] = [];
    for (let r = r0; r <= r1; r += 0.12) {
      const p = hexToPixel(t.pos, r, this.view.size);
      pts.push({ x: this.view.ox + p.x, y: this.view.oy + p.y });
    }
    if (pts.length < 2) return;
    const east = t.direction !== "west";
    ctx.beginPath();
    ctx.moveTo(east ? -20 : w + 20, -20);
    ctx.lineTo(pts[0].x, -20);
    for (const p of pts) ctx.lineTo(p.x, p.y);
    ctx.lineTo(pts[pts.length - 1].x, h + 20);
    ctx.lineTo(east ? -20 : w + 20, h + 20);
    ctx.closePath();
    const shade = ctx.createLinearGradient(east ? 0 : w, 0, pts[Math.floor(pts.length / 2)].x, 0);
    shade.addColorStop(0, "rgba(4,8,14,0.82)");
    shade.addColorStop(1, "rgba(46,134,171,0.18)");
    ctx.fillStyle = shade;
    ctx.fill();
    ctx.save();
    ctx.strokeStyle = `rgba(224,192,90,${0.35 + Math.sin(this.clock * 3) * 0.12})`;
    ctx.lineWidth = 10;
    ctx.lineJoin = "round";
    ctx.beginPath();
    ctx.moveTo(pts[0].x, pts[0].y);
    for (let i = 1; i < pts.length; i++) ctx.lineTo(pts[i].x, pts[i].y);
    ctx.stroke();
    ctx.strokeStyle = `rgba(224,192,90,${0.75 + Math.sin(this.clock * 3) * 0.2})`;
    ctx.lineWidth = 3;
    ctx.stroke();
    ctx.restore();
    const mid = pts[Math.floor(pts.length * 0.18)] || pts[0];
    ctx.fillStyle = "rgba(224,192,90,0.92)";
    ctx.font = "700 12px sans-serif";
    ctx.textAlign = east ? "right" : "left";
    ctx.fillText(east ? "ТЕНЬ →" : "← ТЕНЬ", mid.x + (east ? -8 : 8), 18);
    ctx.textAlign = "left";
  }

  private drawPolyline(ctx: CanvasRenderingContext2D, points: Pt[], color: string, width: number) {
    if (points.length < 2) return;
    ctx.beginPath();
    ctx.moveTo(points[0].x, points[0].y);
    for (let i = 1; i < points.length; i++) ctx.lineTo(points[i].x, points[i].y);
    ctx.strokeStyle = color;
    ctx.lineWidth = width;
    ctx.lineJoin = "round";
    ctx.lineCap = "round";
    ctx.stroke();
  }

  private drawSpace(ctx: CanvasRenderingContext2D, w: number, h: number) {
    const sky = ctx.createLinearGradient(0, 0, 0, h);
    sky.addColorStop(0, "#071018");
    sky.addColorStop(0.45, "#0b1520");
    sky.addColorStop(1, "#05070a");
    ctx.fillStyle = sky;
    ctx.fillRect(0, 0, w, h);
    const glow = ctx.createRadialGradient(w * 0.78, h * 0.12, 10, w * 0.78, h * 0.12, w * 0.55);
    glow.addColorStop(0, "rgba(70, 130, 180, 0.28)");
    glow.addColorStop(0.35, "rgba(46, 86, 130, 0.1)");
    glow.addColorStop(1, "rgba(0,0,0,0)");
    ctx.fillStyle = glow;
    ctx.fillRect(0, 0, w, h);
    ctx.beginPath();
    ctx.arc(w * 0.86, h * 0.1, Math.max(28, w * 0.045), 0, Math.PI * 2);
    ctx.fillStyle = "#8eb4d4";
    ctx.fill();
    ctx.beginPath();
    ctx.arc(w * 0.845, h * 0.09, Math.max(28, w * 0.045) * 0.92, 0, Math.PI * 2);
    ctx.fillStyle = "#071018";
    ctx.fill();
    for (let i = 0; i < 90; i++) {
      const sx = (i * 127.1) % w;
      const sy = (i * 71.3) % h;
      const tw = 0.25 + 0.75 * (0.5 + 0.5 * Math.sin(this.clock * (1.2 + (i % 5) * 0.4) + i));
      ctx.fillStyle = `rgba(255,255,255,${0.15 + tw * 0.7})`;
      ctx.fillRect(sx, sy, i % 9 === 0 ? 2.2 : 1.2, i % 9 === 0 ? 2.2 : 1.2);
    }
  }

  private drawGhost(ctx: CanvasRenderingContext2D, game: Snapshot) {
    if (!game.ghost?.points?.length) return;
    ctx.save();
    ctx.globalAlpha = 0.35;
    ctx.strokeStyle = "rgba(154,173,221,0.9)";
    ctx.lineWidth = 2;
    ctx.setLineDash([4, 6]);
    ctx.beginPath();
    game.ghost.points.forEach((pt, i) => {
      const p = this.world(pt);
      if (i === 0) ctx.moveTo(p.x, p.y);
      else ctx.lineTo(p.x, p.y);
    });
    ctx.stroke();
    ctx.setLineDash([]);
    let gp = null;
    for (const pt of game.ghost.points) if (pt.t <= (game.t || 0)) gp = pt;
    if (gp) {
      const p = this.world(gp);
      ctx.globalAlpha = 0.7;
      ctx.beginPath();
      ctx.arc(p.x, p.y, this.view.size * 0.2, 0, Math.PI * 2);
      ctx.fillStyle = "#9aaddd";
      ctx.fill();
      ctx.fillStyle = "rgba(243,246,248,0.8)";
      ctx.font = "700 10px sans-serif";
      ctx.textAlign = "center";
      ctx.fillText("призрак", p.x, p.y - this.view.size * 0.32);
    }
    ctx.restore();
  }

  private drawPaths(ctx: CanvasRenderingContext2D, game: Snapshot) {
    for (const r of roversOf(game)) {
      const path = r.path || [];
      if (!path.length) continue;
      const off = r.type === "hauler" ? 8 : -8;
      const rp = this.smoothPixel(game, r, off);
      const ahead = path.map((id) => this.world(parseHexId(id)));
      const ok = r.type !== this.selected || !game.routePreview || game.routePreview.feasible;
      const col = r.type === "hauler" ? "rgba(90,166,212,0.95)" : ok ? "rgba(255,255,255,0.92)" : "rgba(224,104,104,0.95)";
      ctx.save();
      ctx.setLineDash([7, 6]);
      this.drawPolyline(ctx, [rp, ...ahead], col, 3);
      ctx.setLineDash([]);
      ctx.restore();
      if (ahead.length) {
        ctx.fillStyle = col;
        ctx.beginPath();
        ctx.arc(ahead[ahead.length - 1].x, ahead[ahead.length - 1].y, 4, 0, Math.PI * 2);
        ctx.fill();
      }
    }
  }

  private drawPins(ctx: CanvasRenderingContext2D, game: Snapshot) {
    const counts: Record<string, number> = {};
    (game.contracts || []).forEach((c, i) => {
      if (["delivered", "failed", "expired", "lost_to_shadow"].includes(c.status)) return;
      const key = c.status === "in_transit" ? c.dropoff : c.pickup;
      counts[key] = (counts[key] || 0) + 1;
      const p = this.world(parseHexId(key));
      const bump = (counts[key] - 1) * 12;
      const bounce = Math.sin(this.clock * 4 + i) * 2;
      const rad = Math.max(12, this.view.size * 0.28);
      ctx.beginPath();
      ctx.arc(p.x + bump, p.y - bounce, rad, 0, Math.PI * 2);
      ctx.fillStyle = c.weight === "heavy" ? "#e06868" : "#e0c05a";
      ctx.shadowColor = ctx.fillStyle;
      ctx.shadowBlur = 12;
      ctx.fill();
      ctx.shadowBlur = 0;
      ctx.fillStyle = "#0b0f14";
      ctx.font = `800 ${Math.max(11, this.view.size * 0.26)}px sans-serif`;
      ctx.textAlign = "center";
      ctx.textBaseline = "middle";
      ctx.fillText(String(i + 1), p.x + bump, p.y - bounce + 1);
      ctx.fillStyle = "rgba(243,246,248,0.9)";
      ctx.font = "700 10px sans-serif";
      ctx.fillText(c.status === "in_transit" ? "сдать" : "забрать", p.x + bump, p.y - bounce + rad + 10);
      const hex = game.map.hexes.find((h) => h.id === key);
      if (hex && c.status !== "in_transit") {
        const eta = hexPhaseEta(hex, liveTerm(game, this.lastApply));
        ctx.fillStyle = eta < 18 ? "#e06868" : "rgba(243,246,248,0.7)";
        ctx.fillText("тень " + Math.ceil(eta) + "с", p.x + bump, p.y - bounce + rad + 22);
      }
      ctx.textBaseline = "alphabetic";
    });
  }

  private drawSalvage(ctx: CanvasRenderingContext2D, game: Snapshot) {
    for (const sv of game.salvage || []) {
      if (sv.status !== "available") continue;
      const hex = game.map.hexes.find((h) => h.id === sv.hex);
      if (!hex || hexDark(hex, liveTerm(game, this.lastApply))) continue;
      const p = this.world(hex);
      const rad = Math.max(8, this.view.size * 0.18);
      ctx.save();
      ctx.translate(p.x, p.y + Math.sin(this.clock * 3) * 1.5);
      ctx.rotate(Math.PI / 4);
      ctx.fillStyle = "#5ec8c8";
      ctx.shadowColor = "#5ec8c8";
      ctx.shadowBlur = 10;
      ctx.fillRect(-rad, -rad, rad * 2, rad * 2);
      ctx.restore();
      ctx.fillStyle = "#071018";
      ctx.font = `800 ${Math.max(9, this.view.size * 0.2)}px sans-serif`;
      ctx.textAlign = "center";
      ctx.textBaseline = "middle";
      ctx.fillText(`+${sv.value}`, p.x, p.y + Math.sin(this.clock * 3) * 1.5);
      ctx.textBaseline = "alphabetic";
    }
  }

  private drawRovers(ctx: CanvasRenderingContext2D, game: Snapshot) {
    for (const r of roversOf(game)) {
      const off = r.type === "hauler" ? 8 : -8;
      const p = this.smoothPixel(game, r, off);
      if (r.state === "moving") this.dust.push({ x: p.x, y: p.y, a: 0.5, s: 3 });
      ctx.save();
      ctx.translate(p.x, p.y);
      ctx.rotate(p.ang);
      ctx.fillStyle = r.state === "stranded" ? "#e06868" : r.type === "hauler" ? "#5aa6d4" : "#fff";
      ctx.shadowColor = ctx.fillStyle;
      ctx.shadowBlur = r.state === "moving" ? 16 : 6;
      ctx.beginPath();
      ctx.moveTo(this.view.size * 0.28, 0);
      ctx.lineTo(-this.view.size * 0.18, this.view.size * 0.16);
      ctx.lineTo(-this.view.size * 0.18, -this.view.size * 0.16);
      ctx.closePath();
      ctx.fill();
      ctx.restore();
      if (r.type === this.selected) {
        ctx.strokeStyle = "rgba(224,192,90,0.8)";
        ctx.lineWidth = 2;
        ctx.beginPath();
        ctx.arc(p.x, p.y, this.view.size * 0.38 + Math.sin(this.clock * 4) * 2, 0, Math.PI * 2);
        ctx.stroke();
      }
    }
  }

  private drawFx(ctx: CanvasRenderingContext2D) {
    this.dust = this.dust.filter((d) => {
      d.a -= 0.03;
      d.y += 0.4;
      d.s *= 0.96;
      return d.a > 0;
    });
    for (const d of this.dust) {
      ctx.fillStyle = `rgba(201,162,39,${d.a})`;
      ctx.beginPath();
      ctx.arc(d.x, d.y, d.s, 0, Math.PI * 2);
      ctx.fill();
    }
    this.flashes = this.flashes.filter((f) => {
      f.a -= 0.04;
      f.r += 2.5;
      return f.a > 0;
    });
    for (const f of this.flashes) {
      ctx.strokeStyle = `rgba(110,207,138,${f.a})`;
      ctx.lineWidth = 3;
      ctx.beginPath();
      ctx.arc(f.x, f.y, f.r, 0, Math.PI * 2);
      ctx.stroke();
    }
  }

  private drawHover(ctx: CanvasRenderingContext2D, game: Snapshot, hex: HexView) {
    const p = this.world(hex);
    const term = liveTerm(game, this.lastApply);
    const lab = terrainRu[hex.type] || hex.type;
    const eta = hexDark(hex, term) ? "уже тень — стена" : `тень через ${Math.ceil(hexPhaseEta(hex, term))}с`;
    ctx.fillStyle = "rgba(5,8,12,0.9)";
    ctx.fillRect(p.x - 118, p.y - this.view.size - 22, 236, 36);
    ctx.fillStyle = "#f3f6f8";
    ctx.font = "12px sans-serif";
    ctx.textAlign = "center";
    ctx.fillText(lab, p.x, p.y - this.view.size - 6);
    ctx.fillStyle = "#8b96a1";
    ctx.font = "11px sans-serif";
    ctx.fillText(eta, p.x, p.y - this.view.size + 10);
    ctx.textAlign = "left";
  }
}
