"use client";

import { useEffect, useRef, type MutableRefObject } from "react";
import { GameRenderer } from "@/lib/draw";
import { useGameStore } from "@/stores/gameStore";

export function GameCanvas({
  rendererRef,
  onClickHex,
}: {
  rendererRef: MutableRefObject<GameRenderer | null>;
  onClickHex: (x: number, y: number) => void;
}) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const setHoverHex = useGameStore((s) => s.setHoverHex);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const renderer = rendererRef.current || new GameRenderer();
    rendererRef.current = renderer;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;
    let raf = 0;
    let lastW = 0;
    let lastH = 0;
    const resize = () => {
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      const w = canvas.clientWidth;
      const h = canvas.clientHeight;
      canvas.width = w * dpr;
      canvas.height = h * dpr;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      lastW = w;
      lastH = h;
    };
    resize();
    const loop = (ts: number) => {
      if (canvas.clientWidth !== lastW || canvas.clientHeight !== lastH) resize();
      const state = useGameStore.getState();
      renderer.selected = state.selectedRover;
      renderer.hover = state.hoverHex;
      renderer.lastApply = state.lastApply;
      renderer.startedAt = state.startedAt;
      renderer.frame(ctx, state.snapshot, canvas.clientWidth, canvas.clientHeight, ts);
      raf = requestAnimationFrame(loop);
    };
    raf = requestAnimationFrame(loop);
    window.addEventListener("resize", resize);
    return () => {
      cancelAnimationFrame(raf);
      window.removeEventListener("resize", resize);
    };
  }, [rendererRef]);

  return (
    <canvas
      ref={canvasRef}
      className="block h-full w-full touch-manipulation"
      onClick={(ev) => {
        const rect = ev.currentTarget.getBoundingClientRect();
        onClickHex(ev.clientX - rect.left, ev.clientY - rect.top);
      }}
      onPointerMove={(ev) => {
        const game = useGameStore.getState().snapshot;
        const r = rendererRef.current;
        if (!game || !r) return;
        const rect = ev.currentTarget.getBoundingClientRect();
        const hex = r.pickHex(game, ev.clientX - rect.left, ev.clientY - rect.top);
        setHoverHex(hex);
        r.hover = hex;
      }}
    />
  );
}
