"use client";

import { batPct, hexPhaseEta, liveTerm, roverStateRu, roversOf } from "@/lib/coach";
import { layoutRu } from "@/lib/copy";
import type { RoverType, Snapshot } from "@/lib/types";
import { useGameStore } from "@/stores/gameStore";

function Meter({ value, label }: { value: number; label: string }) {
  return (
    <div className="min-w-0 flex-1">
      <div className="h-1.5 overflow-hidden rounded-full bg-black/40 sm:h-2">
        <i
          className="block h-full origin-left rounded-full bg-gradient-to-r from-danger via-gold to-ok"
          style={{ transform: `scaleX(${value / 100})` }}
        />
      </div>
      <div className="mt-0.5 truncate text-[10px] text-muted sm:text-[11px]">{label}</div>
    </div>
  );
}

export function Hud({
  snapshot,
  selectedRover,
  onSelect,
  crisis,
}: {
  snapshot: Snapshot | null;
  selectedRover: RoverType;
  onSelect: (t: RoverType) => void;
  crisis: string;
}) {
  const hint = useGameStore((s) => s.hint);
  const lastApply = useGameStore((s) => s.lastApply);
  const list = roversOf(snapshot);
  const sw = list.find((x) => x.type === "swift");
  const ha = list.find((x) => x.type === "hauler");
  const need = snapshot?.goal?.colonyNeed || 100;
  const dur = snapshot?.goal?.duration || 180;
  const left = Math.max(0, dur - (snapshot?.t || 0));
  const fill = Math.min(100, ((snapshot?.colonyScore || 0) / need) * 100);
  const act = list.find((x) => x.type === selectedRover) || list[0];
  const here = snapshot?.map.hexes.find((x) => x.id === act?.hex);
  const eta = here && snapshot ? hexPhaseEta(here, liveTerm(snapshot, lastApply)) : 0;
  const lay = snapshot?.layout ? layoutRu[snapshot.layout] || snapshot.layout : "";

  return (
    <div className="relative z-10 shrink-0 px-3 pt-2 sm:px-4 sm:pt-3">
      <header className="flex items-center gap-2">
        <div className="min-w-0 flex-1">
          <div className="font-display text-[13px] font-semibold tracking-[0.16em] text-gold sm:text-[15px] sm:tracking-[0.18em]">
            PHASELINE
          </div>
          <div className="truncate font-sans text-[11px] tracking-normal text-muted">
            {snapshot?.seed || "…"}
            {lay ? ` · ${lay}` : ""}
          </div>
        </div>
        <div className="flex min-w-0 shrink-0 gap-1.5">
          <button
            type="button"
            onClick={() => onSelect("swift")}
            className={`min-h-10 w-[7.25rem] rounded-xl border px-2 py-1 text-left text-[12px] font-extrabold sm:min-h-11 sm:w-[8.25rem] sm:px-3 sm:text-[13px] ${
              selectedRover === "swift" ? "border-gold bg-gold/10 shadow-gold" : "border-white/10 bg-white/5"
            }`}
          >
            Swift
            <small className="block truncate text-[10px] font-medium text-muted sm:text-[11px]">
              {sw ? `${batPct(sw).toFixed(0)}% · ${roverStateRu(sw)}` : "быстрый"}
            </small>
          </button>
          <button
            type="button"
            onClick={() => onSelect("hauler")}
            className={`min-h-10 w-[7.25rem] rounded-xl border px-2 py-1 text-left text-[12px] font-extrabold sm:min-h-11 sm:w-[8.25rem] sm:px-3 sm:text-[13px] ${
              selectedRover === "hauler" ? "border-gold bg-gold/10 shadow-gold" : "border-white/10 bg-white/5"
            }`}
          >
            Hauler
            <small className="block truncate text-[10px] font-medium text-muted sm:text-[11px]">
              {ha ? `${batPct(ha).toFixed(0)}% · ${roverStateRu(ha)}` : "тягач"}
            </small>
          </button>
        </div>
      </header>

      <div className="mt-2 sm:mt-3">
        <div className="mb-1 flex justify-between gap-2 text-[12px] text-muted sm:text-[13px]">
          <span className="min-w-0 truncate">
            Колония <b className="text-white">{snapshot?.colonyScore ?? 0}</b> / <b className="text-white">{need}</b>
            {" · "}Земля <b className="text-white">{snapshot?.earthScore ?? 0}</b>
          </span>
          <span className="shrink-0">
            <b className="text-white">{left.toFixed(0)}</b> с
          </span>
        </div>
        <div className="h-2 overflow-hidden rounded-full bg-white/10 sm:h-2.5">
          <div
            className="h-full rounded-full bg-gradient-to-r from-cold to-ok transition-[width] duration-300"
            style={{ width: `${fill}%` }}
          />
        </div>
      </div>

      <div
        className={`mt-2 line-clamp-3 rounded-xl border px-3 py-2 text-[13px] font-semibold leading-snug sm:mt-3 sm:line-clamp-none sm:px-3.5 sm:py-3 sm:text-[15px] ${
          hint.kind === "warn"
            ? "border-danger bg-danger/15 text-[#ffd4d4]"
            : hint.kind === "ok"
              ? "border-ok bg-ok/10 text-[#c8f0d4]"
              : "border-white/10 bg-gradient-to-b from-white/10 to-black/20"
        }`}
      >
        {hint.text}
      </div>
      {crisis ? (
        <div className="mt-2 line-clamp-2 rounded-[10px] border border-danger bg-danger/20 px-3 py-1.5 text-[13px] font-bold sm:py-2">
          {crisis}
        </div>
      ) : null}

      <div className="mt-2 flex gap-3 text-[10px] text-muted sm:text-xs">
        <Meter value={batPct(sw)} label={`Swift ${batPct(sw).toFixed(0)}% · ${sw?.slotsUsed ?? 0}/${sw?.slotsMax ?? 2}`} />
        <Meter value={batPct(ha)} label={`Hauler ${batPct(ha).toFixed(0)}% · ${ha?.slotsUsed ?? 0}/${ha?.slotsMax ?? 2}`} />
        <div className="hidden shrink-0 self-end pb-0.5 sm:block">Тень: {here ? eta.toFixed(0) : "—"} с</div>
      </div>
    </div>
  );
}
