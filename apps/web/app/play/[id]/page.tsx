"use client";

import { useParams } from "next/navigation";
import { GameShell } from "@/components/game/GameShell";

export default function PlayPage() {
  const { id } = useParams<{ id: string }>();
  if (!id) return null;
  return <GameShell resumeId={id} />;
}
