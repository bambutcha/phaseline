"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { resumeAudio, setSoundEnabled, soundEnabled } from "@/lib/audio";
import { createGame, waitHealth } from "@/lib/api";
import { tutorialSlides } from "@/lib/copy";

export default function HomePage() {
  const router = useRouter();
  const [seed, setSeed] = useState("");
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState("");
  const [sound, setSound] = useState(true);
  const [help, setHelp] = useState(false);

  useEffect(() => {
    setSound(soundEnabled());
    const shared = new URLSearchParams(window.location.search).get("seed");
    if (shared) router.replace(`/s/${shared}`);
  }, [router]);

  const start = async (nextSeed?: string) => {
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
      const trimmed = (nextSeed ?? seed).trim();
      const data = await createGame(trimmed || undefined, "swift");
      if (!data?.id) throw new Error("no id");
      router.push(`/play/${data.id}`);
    } catch {
      setErr("Не удалось создать смену.");
      setBusy(false);
    }
  };

  return (
    <main className="relative flex min-h-dvh flex-col items-center justify-center overflow-hidden bg-ink px-5 py-10">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top,rgba(224,192,90,0.16),transparent_55%),radial-gradient(ellipse_at_bottom,rgba(46,134,171,0.18),transparent_50%)]" />
      <div className="scanlines vignette pointer-events-none absolute inset-0" />
      <div className="relative z-10 w-full max-w-lg">
        <p className="text-xs tracking-[0.32em] text-gold">MOON COURIER CRISIS</p>
        <h1 className="mt-2 font-display text-[clamp(2rem,11vw,3.75rem)] font-semibold tracking-[0.12em] text-gold sm:tracking-[0.18em]">PHASELINE</h1>
        <p className="mt-4 max-w-md text-[17px] leading-snug text-[#d5dde4]">
          Два ровера. Тень как стена. Три минуты. Цель — 100 колонии. Всех не спасти.
        </p>
        <label className="mt-8 block text-xs uppercase tracking-[0.16em] text-muted">Сид (необязательно)</label>
        <input
          value={seed}
          onChange={(e) => setSeed(e.target.value.toUpperCase())}
          placeholder="MCC-XXXX"
          className="mt-2 min-h-11 w-full rounded-xl border border-white/10 bg-black/35 px-3 font-display tracking-[0.12em] outline-none ring-gold/40 focus:ring-2"
        />
        {err ? <p className="mt-3 text-sm text-danger">{err}</p> : null}
        <div className="mt-5 flex flex-col gap-2 sm:flex-row">
          <button
            type="button"
            disabled={busy}
            onClick={() => void start()}
            className="min-h-11 flex-1 rounded-xl bg-gold font-extrabold tracking-[0.12em] text-black shadow-gold disabled:opacity-50"
          >
            {busy ? "СВЯЗЬ…" : "НОВАЯ СМЕНА"}
          </button>
          <button
            type="button"
            onClick={() => setHelp(true)}
            className="min-h-11 flex-1 rounded-xl bg-[#1c242d] font-extrabold tracking-[0.12em]"
          >
            КАК ИГРАТЬ
          </button>
        </div>
        <button
          type="button"
          className="mt-4 text-sm text-muted underline-offset-4 hover:text-white hover:underline"
          onClick={() => {
            const next = !sound;
            setSound(next);
            setSoundEnabled(next);
          }}
        >
          Звук: {sound ? "вкл" : "выкл"}
        </button>
      </div>
      {help ? (
        <div className="fixed inset-0 z-20 flex items-end justify-center bg-black/55 p-4 sm:items-center">
          <div className="glass w-full max-w-[520px] rounded-2xl p-5">
            <h2 className="font-display text-lg text-gold">Как играть</h2>
            <ol className="mt-3 list-decimal space-y-2 pl-5 text-[#d5dde4]">
              {tutorialSlides.map((s) => (
                <li key={s.title}>
                  <b>{s.title}</b> {s.body}
                </li>
              ))}
            </ol>
            <button
              type="button"
              className="mt-5 min-h-11 w-full rounded-xl bg-gold font-extrabold text-black"
              onClick={() => setHelp(false)}
            >
              ПОНЯЛ
            </button>
          </div>
        </div>
      ) : null}
    </main>
  );
}
