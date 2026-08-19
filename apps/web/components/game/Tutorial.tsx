"use client";

import { useState } from "react";
import { resumeAudio } from "@/lib/audio";
import { tutorialSlides } from "@/lib/copy";
import { useGameStore } from "@/stores/gameStore";

const TUTORIAL_KEY = "phaseline_tutorial_done";

export function tutorialDone(): boolean {
  if (typeof window === "undefined") return true;
  return Boolean(localStorage.getItem(TUTORIAL_KEY) || sessionStorage.getItem(TUTORIAL_KEY));
}

export function markTutorialDone() {
  try {
    localStorage.setItem(TUTORIAL_KEY, "1");
    sessionStorage.setItem(TUTORIAL_KEY, "1");
  } catch {
    /* ignore */
  }
}

export function Tutorial() {
  const open = useGameStore((s) => s.tutorialOpen);
  const setOpen = useGameStore((s) => s.setTutorialOpen);
  const [slide, setSlide] = useState(0);
  if (!open) return null;
  const item = tutorialSlides[slide];
  const last = slide === tutorialSlides.length - 1;

  const close = () => {
    markTutorialDone();
    resumeAudio();
    setOpen(false);
  };

  return (
    <div className="fixed inset-0 z-30 flex items-end justify-center bg-black/55 p-3 sm:items-center sm:p-4">
      <div className="glass max-h-[min(88dvh,560px)] w-full max-w-[520px] overflow-y-auto rounded-2xl p-4 shadow-gold sm:p-5">
        <p className="mb-1 text-xs tracking-[0.18em] text-gold">
          {slide + 1} / {tutorialSlides.length}
        </p>
        <h2 className="font-display text-lg font-semibold tracking-wide text-gold">{item.title}</h2>
        <p className="mt-3 leading-relaxed text-[#d5dde4]">{item.body}</p>
        <div className="mt-5 flex gap-2">
          <button
            type="button"
            className="min-h-11 flex-1 rounded-xl bg-[#1c242d] font-extrabold tracking-wide text-white"
            onClick={close}
          >
            ПРОПУСТИТЬ
          </button>
          <button
            type="button"
            className="min-h-11 flex-1 rounded-xl bg-gold font-extrabold tracking-wide text-black"
            onClick={() => {
              if (last) close();
              else setSlide((n) => n + 1);
            }}
          >
            {last ? "ПОНЯЛ — ИГРАЕМ" : "ДАЛЬШЕ"}
          </button>
        </div>
      </div>
    </div>
  );
}
