import { create } from "zustand";
import type { HexView, HintKind, RoverType, Snapshot } from "@/lib/types";

type GameStore = {
  snapshot: Snapshot | null;
  gameId: string | null;
  selectedRover: RoverType;
  selectedId: string | null;
  hint: { text: string; kind: HintKind };
  hoverHex: HexView | null;
  tutorialOpen: boolean;
  lastApply: number;
  startedAt: number;
  shareUrl: string;
  hintUntil: number;
  setSnapshot: (snap: Snapshot | null) => void;
  setGameId: (id: string | null) => void;
  setSelectedRover: (t: RoverType) => void;
  setSelectedId: (id: string | null) => void;
  setHint: (text: string, kind?: HintKind, holdMs?: number) => void;
  setHoverHex: (hex: HexView | null) => void;
  setTutorialOpen: (open: boolean) => void;
  markApply: () => void;
  resetSession: () => void;
};

export const useGameStore = create<GameStore>((set) => ({
  snapshot: null,
  gameId: null,
  selectedRover: "swift",
  selectedId: null,
  hint: { text: "СЕЙЧАС: нажми пульсирующую карточку внизу. Цифра на карте — точка забора.", kind: "" },
  hoverHex: null,
  tutorialOpen: true,
  lastApply: 0,
  startedAt: 0,
  shareUrl: "",
  hintUntil: 0,
  setSnapshot: (snapshot) => set({ snapshot }),
  setGameId: (gameId) => set({ gameId }),
  setSelectedRover: (selectedRover) =>
    set((s) => ({
      selectedRover,
      snapshot: s.snapshot ? { ...s.snapshot, activeRover: selectedRover } : s.snapshot,
    })),
  setSelectedId: (selectedId) => set({ selectedId }),
  setHint: (text, kind = "", holdMs = 0) =>
    set({ hint: { text, kind }, hintUntil: holdMs ? performance.now() + holdMs : 0 }),
  setHoverHex: (hoverHex) => set({ hoverHex }),
  setTutorialOpen: (tutorialOpen) => set({ tutorialOpen }),
  markApply: () => set({ lastApply: performance.now() }),
  resetSession: () =>
    set({
      snapshot: null,
      selectedRover: "swift",
      selectedId: null,
      shareUrl: "",
      startedAt: performance.now(),
      lastApply: performance.now(),
      hintUntil: 0,
    }),
}));
