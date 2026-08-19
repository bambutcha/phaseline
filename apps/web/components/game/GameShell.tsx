"use client";

import { useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { resumeAudio } from "@/lib/audio";
import type { GameRenderer } from "@/lib/draw";
import { createGame, waitHealth } from "@/lib/api";
import { useGameSession } from "@/hooks/useGameSession";
import { useGameStore } from "@/stores/gameStore";
import { BlackBox } from "./BlackBox";
import { GameCanvas } from "./GameCanvas";
import { Hud } from "./Hud";
import { Legend } from "./Legend";
import { Orders } from "./Orders";
import { Tutorial, tutorialDone } from "./Tutorial";

export function GameShell({ resumeId }: { resumeId: string }) {
  const router = useRouter();
  const rendererRef = useRef<GameRenderer | null>(null);
  const setTutorialOpen = useGameStore((s) => s.setTutorialOpen);
  const setHint = useGameStore((s) => s.setHint);
  const session = useGameSession(rendererRef, resumeId);

  useEffect(() => {
    const shot = new URLSearchParams(window.location.search).get("shot");
    setTutorialOpen(!shot && !tutorialDone());
  }, [setTutorialOpen]);

  const share = async () => {
    const url = `${location.origin}${session.shareUrl || `/s/${session.snapshot?.seed || ""}`}`;
    try {
      await navigator.clipboard.writeText(url);
      setHint("Ссылка скопирована. Друг получит ту же карту.");
    } catch {
      prompt("Сид:", url);
    }
  };

  const again = async () => {
    setHint("Готовим новую смену…");
    const ok = await waitHealth();
    if (!ok) {
      setHint("Сервер не поднялся. docker compose up --build", "warn");
      return;
    }
    try {
      const data = await createGame(undefined, "swift");
      if (!data?.id) throw new Error("no id");
      router.replace(`/play/${data.id}`);
    } catch {
      setHint("Не удалось создать смену.", "warn");
    }
  };

  return (
    <div className="flex h-dvh min-h-0 flex-col overflow-hidden bg-ink">
      <Hud
        snapshot={session.snapshot}
        selectedRover={session.selectedRover}
        onSelect={session.selectRover}
        crisis={session.crisis}
      />
      <div className="relative min-h-0 flex-1">
        <div className="scanlines vignette absolute inset-0">
          <GameCanvas rendererRef={rendererRef} onClickHex={session.onCanvasClick} />
        </div>
      </div>
      <Legend />
      <Orders snapshot={session.snapshot} selectedRover={session.selectedRover} onDispatch={session.dispatch} />
      <div className="flex shrink-0 gap-2 px-3 pb-[calc(10px+env(safe-area-inset-bottom))] pt-1 sm:px-4 sm:pb-[calc(12px+env(safe-area-inset-bottom))]">
        <button
          type="button"
          className="min-h-11 flex-1 rounded-xl bg-[#1c242d] font-extrabold tracking-wide"
          onClick={() => {
            resumeAudio();
            setTutorialOpen(true);
          }}
        >
          КАК ИГРАТЬ
        </button>
        <button
          type="button"
          className="min-h-11 flex-1 rounded-xl bg-[#1c242d] font-extrabold tracking-wide"
          onClick={() => void again()}
        >
          НОВАЯ СМЕНА
        </button>
      </div>
      <Tutorial />
      <BlackBox
        open={Boolean(session.finished)}
        snapshot={session.snapshot}
        data={session.blackbox}
        shareUrl={session.shareUrl}
        onShare={() => void share()}
        onAgain={() => void again()}
      />
    </div>
  );
}
