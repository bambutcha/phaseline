"use client";

import { useCallback, useEffect, useRef, type MutableRefObject } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { beep } from "@/lib/audio";
import { gameWsUrl, getBlackBox, getGame } from "@/lib/api";
import { coach, crisisText, hexDark, liveTerm, roversOf } from "@/lib/coach";
import { crisisRu } from "@/lib/copy";
import type { GameRenderer } from "@/lib/draw";
import type { RoverType, Snapshot } from "@/lib/types";
import { useGameStore } from "@/stores/gameStore";

export function useGameSession(renderer: MutableRefObject<GameRenderer | null>, resumeId: string) {
  const wsRef = useRef<WebSocket | null>(null);
  const closeIntent = useRef(false);
  const reconnect = useRef<ReturnType<typeof setTimeout> | null>(null);

  const snapshot = useGameStore((s) => s.snapshot);
  const gameId = useGameStore((s) => s.gameId);
  const selectedRover = useGameStore((s) => s.selectedRover);
  const hoverHex = useGameStore((s) => s.hoverHex);
  const lastApply = useGameStore((s) => s.lastApply);
  const setSnapshot = useGameStore((s) => s.setSnapshot);
  const setGameId = useGameStore((s) => s.setGameId);
  const setHint = useGameStore((s) => s.setHint);
  const setSelectedRover = useGameStore((s) => s.setSelectedRover);
  const setSelectedId = useGameStore((s) => s.setSelectedId);
  const markApply = useGameStore((s) => s.markApply);
  const resetSession = useGameStore((s) => s.resetSession);

  const apply = useCallback(
    (msg: Snapshot) => {
      if (!msg || msg.error || !msg.map) return;
      const prev = useGameStore.getState().snapshot;
      const prevScore = prev?.colonyScore || 0;
      const prevFail = (prev?.contracts || []).filter((c) => ["expired", "lost_to_shadow", "failed"].includes(c.status)).length;
      const selected = useGameStore.getState().selectedRover;
      const next = { ...msg, activeRover: selected };
      setSnapshot(next);
      markApply();
      if (renderer.current) renderer.current.lastApply = performance.now();
      if ((msg.colonyScore || 0) > prevScore) {
        beep(880, 0.1);
        const r = roversOf(next).find((x) => x.type === selected) || roversOf(next)[0];
        const p = r && renderer.current ? renderer.current.roverPixel(next, r, 0) : { x: 80, y: 80 };
        renderer.current?.flash(p.x, p.y);
      }
      for (const e of msg.deltaEvents || []) {
        if (e.kind === "rejected" || e.kind === "stranded") beep(140, 0.12);
        if (e.kind === "crisis") setHint(crisisRu[(e.payload?.kind as string) || ""] || "Кризис! Правила на карте изменились.", "warn", 2800);
        if (e.kind === "pickup") setHint("Груз в кузове. Сдавай на золотую базу на свету — в тень нельзя.", "ok", 2200);
        if (e.kind === "dropoff_moved") setHint("Точка сдачи ушла в тень. Вези на живую золотую базу.", "warn", 2800);
        if (e.kind === "deliver") setHint("Груз сдан. Бери следующий заказ — тень не ждёт.", "ok", 2200);
        if (e.kind === "salvage") setHint("Кассета подобрана. Ещё немного колонии — бери следующую, пока светло.", "ok", 2200);
        if (e.kind === "contract_failed") beep(120, 0.1);
      }
      const failNow = (msg.contracts || []).filter((c) => ["expired", "lost_to_shadow", "failed"].includes(c.status)).length;
      if (failNow > prevFail) setHint("Заказ сгорел. Не жалей — бери следующий, пока тень ближе.", "warn", 2200);
    },
    [markApply, renderer, setHint, setSnapshot],
  );

  const connect = useCallback(
    (id: string) => {
      closeIntent.current = true;
      wsRef.current?.close();
      if (reconnect.current) clearTimeout(reconnect.current);
      closeIntent.current = false;
      setGameId(id);
      const ws = new WebSocket(gameWsUrl(id));
      wsRef.current = ws;
      ws.onmessage = (ev) => apply(JSON.parse(ev.data) as Snapshot);
      ws.onclose = () => {
        if (closeIntent.current || useGameStore.getState().snapshot?.status === "finished") return;
        setHint("Связь с ровером пропала, переподключаюсь…");
        reconnect.current = setTimeout(() => connect(id), 1200);
      };
    },
    [apply, setGameId, setHint],
  );

  const resumeMut = useMutation({
    mutationFn: (id: string) => getGame(id),
    onSuccess: (data, id) => {
      apply(data);
      connect(id);
    },
    onError: () => setHint("Смена не найдена. Вернись в меню и начни новую.", "warn"),
  });

  const blackboxQ = useQuery({
    queryKey: ["blackbox", gameId],
    queryFn: () => getBlackBox(gameId!),
    enabled: Boolean(gameId && snapshot?.status === "finished"),
    retry: 1,
  });

  useEffect(() => {
    resetSession();
    renderer.current?.reset();
    if (resumeId) resumeMut.mutate(resumeId);
    return () => {
      closeIntent.current = true;
      wsRef.current?.close();
      if (reconnect.current) clearTimeout(reconnect.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [resumeId]);

  useEffect(() => {
    if (!snapshot) return;
    if (useGameStore.getState().hintUntil > performance.now()) return;
    const warn = snapshot.reject?.reason || (snapshot.routePreview && !snapshot.routePreview.feasible);
    const kind = warn ? "warn" : snapshot.status === "finished" ? "ok" : "";
    setHint(coach(snapshot, selectedRover, hoverHex, lastApply), kind);
  }, [hoverHex, lastApply, selectedRover, setHint, snapshot]);

  const send = useCallback((payload: Record<string, unknown>) => {
    const ws = wsRef.current;
    if (ws?.readyState === 1) ws.send(JSON.stringify(payload));
  }, []);

  const selectRover = useCallback(
    (type: RoverType) => {
      setSelectedRover(type);
      send({ type: "select_rover", rover: type });
    },
    [send, setSelectedRover],
  );

  const dispatch = useCallback(
    (contractId: string) => {
      setSelectedId(contractId);
      const game = useGameStore.getState().snapshot;
      if (!game || game.status === "finished") return;
      send({ type: "dispatch", contractId, rover: useGameStore.getState().selectedRover });
    },
    [send, setSelectedId],
  );

  const goto = useCallback(
    (hexId: string) => {
      send({ type: "goto", hexId, rover: useGameStore.getState().selectedRover });
    },
    [send],
  );

  const onCanvasClick = useCallback(
    (x: number, y: number) => {
      const game = useGameStore.getState().snapshot;
      const r = renderer.current;
      if (!game || !r || game.status === "finished" || wsRef.current?.readyState !== 1) return;
      const rv = r.pickRover(game, x, y);
      if (rv) {
        selectRover(rv.type);
        beep(440, 0.04);
        return;
      }
      const hex = r.pickHex(game, x, y);
      if (!hex) return;
      if (hexDark(hex, liveTerm(game, useGameStore.getState().lastApply))) {
        setHint("В тень нельзя. Это стена.", "warn", 1800);
        beep(140, 0.08);
        return;
      }
      const c = (game.contracts || []).find((ct) => {
        if (["delivered", "failed", "expired", "lost_to_shadow"].includes(ct.status)) return false;
        const key = ct.status === "in_transit" ? ct.dropoff : ct.pickup;
        return key === hex.id;
      });
      if (c) {
        dispatch(c.id);
        beep(620, 0.05);
      } else {
        goto(hex.id);
        beep(380, 0.04);
      }
    },
    [dispatch, goto, renderer, selectRover, setHint],
  );

  return {
    snapshot,
    gameId,
    selectedRover,
    selectRover,
    dispatch,
    onCanvasClick,
    blackbox: blackboxQ.data,
    shareUrl: blackboxQ.data?.shareUrl || (snapshot?.seed ? `/s/${snapshot.seed}` : ""),
    crisis: crisisText(snapshot),
    finished: snapshot?.status === "finished",
  };
}
