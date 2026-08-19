"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { createGame, getSeedPreview, waitHealth } from "@/lib/api";
import { layoutRu } from "@/lib/copy";
import { resumeAudio } from "@/lib/audio";

export default function SeedPage() {
  const { seed } = useParams<{ seed: string }>();
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState("");
  const preview = useQuery({
    queryKey: ["seed", seed],
    queryFn: () => getSeedPreview(seed),
    enabled: Boolean(seed),
  });

  const start = async () => {
    setBusy(true);
    setErr("");
    resumeAudio();
    const ok = await waitHealth();
    if (!ok) {
      setErr("Сервер не поднялся. docker compose up --build");
      setBusy(false);
      return;
    }
    try {
      const data = await createGame(seed, "swift");
      if (!data?.id) throw new Error("no id");
      router.push(`/play/${data.id}`);
    } catch {
      setErr("Не удалось открыть эту смену.");
      setBusy(false);
    }
  };

  const data = preview.data;
  const dir = data?.map.terminator.direction === "west" ? "на запад" : "на восток";
  const lay = data?.layout ? layoutRu[data.layout] || data.layout : "без спойлеров контрактов";

  return (
    <main className="relative flex min-h-dvh items-center justify-center overflow-hidden bg-ink px-5">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top,rgba(224,192,90,0.14),transparent_55%)]" />
      <div className="scanlines vignette pointer-events-none absolute inset-0" />
      <div className="relative z-10 w-full max-w-lg">
        <p className="text-xs tracking-[0.32em] text-gold">SHARED SHIFT</p>
        <h1 className="mt-2 font-display text-4xl tracking-[0.14em] text-gold">{seed}</h1>
        <p className="mt-4 text-[#d5dde4]">
          Та же карта и направление тени. Контракты и кризис скрыты — играй вслепую, как автор сида.
        </p>
        <ul className="mt-5 space-y-1 text-sm text-muted">
          <li>ландшафт: {preview.isLoading ? "…" : lay}</li>
          <li>гексов: {data?.map.hexCount ?? "…"}</li>
          <li>терминатор: {data ? dir : "…"}</li>
        </ul>
        {err ? <p className="mt-3 text-sm text-danger">{err}</p> : null}
        <div className="mt-6 flex gap-2">
          <button
            type="button"
            className="min-h-11 flex-1 rounded-xl bg-[#1c242d] font-extrabold"
            onClick={() => router.push("/")}
          >
            МЕНЮ
          </button>
          <button
            type="button"
            disabled={busy}
            className="min-h-11 flex-1 rounded-xl bg-gold font-extrabold text-black shadow-gold disabled:opacity-50"
            onClick={() => void start()}
          >
            {busy ? "СВЯЗЬ…" : "НАЧАТЬ ЭТУ СМЕНУ"}
          </button>
        </div>
      </div>
    </main>
  );
}
