"use client";

import { endReasonRu, outcomeTitle } from "@/lib/copy";
import type { BlackBox as BlackBoxData, Snapshot } from "@/lib/types";

export function BlackBox({
  open,
  snapshot,
  data,
  shareUrl,
  onShare,
  onAgain,
}: {
  open: boolean;
  snapshot: Snapshot | null;
  data?: BlackBoxData;
  shareUrl: string;
  onShare: () => void;
  onAgain: () => void;
}) {
  if (!open) return null;
  const outcome = data?.outcome || snapshot?.outcome || "";
  const title = outcomeTitle[outcome] || "Black Box";
  const verdict = data?.verdict || "Смена окончена.";
  const colony = data?.colonyScore ?? snapshot?.colonyScore ?? 0;
  const earth = data?.earthScore ?? snapshot?.earthScore ?? 0;
  const seed = data?.seed || snapshot?.seed || "";
  const reason = endReasonRu[data?.endReason || snapshot?.endReason || ""] || "";
  const ev = (data?.events || snapshot?.events || [])
    .slice(-8)
    .map((e) => e.kind)
    .join(" → ");

  return (
    <div className="fixed inset-0 z-40 flex items-center justify-center bg-black/88 p-4">
      <div className="glass w-full max-w-[520px] rounded-2xl p-5 shadow-gold">
        <h2 className="font-display text-xl tracking-wide text-gold">{title}</h2>
        <p className="mt-3 leading-relaxed text-[#d5dde4]">{verdict}</p>
        <p className="mt-3 whitespace-pre-wrap text-[13px] text-muted">
          {`Колония ${colony} · Земля ${earth} · ${seed}`}
          {reason ? ` · ${reason}` : ""}
          {ev ? `\n${ev}` : ""}
          {shareUrl ? `\n${shareUrl}` : ""}
        </p>
        <div className="mt-5 flex gap-2">
          <button
            type="button"
            className="min-h-11 flex-1 rounded-xl bg-[#1c242d] font-extrabold tracking-wide text-white"
            onClick={onShare}
          >
            ПОДЕЛИТЬСЯ СИДОМ
          </button>
          <button
            type="button"
            className="min-h-11 flex-1 rounded-xl bg-gold font-extrabold tracking-wide text-black"
            onClick={onAgain}
          >
            ЕЩЁ РАЗ
          </button>
        </div>
      </div>
    </div>
  );
}
