"use client";

import { roversOf } from "@/lib/coach";
import { riskRu, statusRu, urgRu, weightRu } from "@/lib/copy";
import type { Contract, RoverType, Snapshot } from "@/lib/types";
import { useGameStore } from "@/stores/gameStore";

function blockedReason(c: Contract, selected: RoverType, game: Snapshot) {
  const r = roversOf(game).find((x) => x.type === selected) || roversOf(game)[0];
  if (!r) return "";
  if (c.weight === "heavy" && !r.canHeavy) return "swift_no_heavy";
  if (c.status === "queued" && (r.slotsUsed || 0) >= (r.slotsMax || 2)) return "slots_full";
  return "";
}

export function Orders({
  snapshot,
  selectedRover,
  onDispatch,
}: {
  snapshot: Snapshot | null;
  selectedRover: RoverType;
  onDispatch: (id: string) => void;
}) {
  const selectedId = useGameStore((s) => s.selectedId);
  if (!snapshot?.contracts) return null;
  const playable = snapshot.contracts.findIndex((x) => x.status === "queued" && !blockedReason(x, selectedRover, snapshot));
  const idle = !roversOf(snapshot).some((r) => r.state === "moving");

  return (
    <div className="flex snap-x snap-mandatory gap-2 overflow-x-auto px-3 py-1.5 [-ms-overflow-style:none] [scrollbar-width:none] sm:px-4 sm:py-2 [&::-webkit-scrollbar]:hidden">
      {snapshot.contracts.map((c, i) => {
        const done = ["delivered", "failed", "expired", "lost_to_shadow"].includes(c.status);
        const block = blockedReason(c, selectedRover, snapshot);
        const active = c.status === "accepted" || c.status === "in_transit" || selectedId === c.id;
        const due =
          c.deadline > 0 && !done ? `${Math.ceil(c.deadline)}с` : urgRu[c.urgency as keyof typeof urgRu] || c.urgency;
        return (
          <button
            key={c.id}
            type="button"
            onClick={() => onDispatch(c.id)}
            className={`w-[min(78vw,220px)] shrink-0 snap-start rounded-xl border bg-panel/90 p-2.5 text-left backdrop-blur-sm transition hover:-translate-y-0.5 sm:w-[200px] ${
              active ? "border-gold shadow-gold" : "border-white/10"
            } ${done ? "opacity-40" : ""} ${block && c.status === "queued" ? "border-danger" : ""} ${
              i === playable && idle ? "animate-nudge" : ""
            }`}
          >
            <h3 className="mb-1 truncate text-[13px] font-bold">
              <span className="mr-1.5 inline-flex h-[22px] w-[22px] items-center justify-center rounded-full bg-gold text-xs font-extrabold text-black">
                {i + 1}
              </span>
              {c.title}
            </h3>
            <div className="text-[12px] text-muted sm:hidden">
              {weightRu[c.weight]} · +{c.colonyValue} · {due}
              {block ? " · смени ровер" : ""}
            </div>
            <div className="hidden grid-cols-2 gap-x-2 gap-y-0.5 text-[12px] text-muted sm:grid">
              <span>
                вес <b className="text-white">{weightRu[c.weight]}</b>
              </span>
              <span>
                награда <b className="text-white">+{c.colonyValue}</b>
              </span>
              <span>
                срок <b className="text-white">{due}</b>
              </span>
              <span>
                риск <b className="text-white">{riskRu[c.risk as keyof typeof riskRu] || c.risk}</b>
              </span>
              <span>{statusRu[c.status] || c.status}</span>
              <span>{block ? "смени ровер" : "нажми — поедет"}</span>
            </div>
          </button>
        );
      })}
    </div>
  );
}
