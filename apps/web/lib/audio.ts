let ctx: AudioContext | null = null;

export function resumeAudio() {
  try {
    ctx = ctx || new AudioContext();
    void ctx.resume();
  } catch {
    /* ignore */
  }
}

export function soundEnabled(): boolean {
  if (typeof localStorage === "undefined") return true;
  return localStorage.getItem("phaseline_sound") !== "0";
}

export function setSoundEnabled(on: boolean) {
  if (typeof localStorage === "undefined") return;
  localStorage.setItem("phaseline_sound", on ? "1" : "0");
}

export function beep(freq: number, dur = 0.08) {
  if (!soundEnabled()) return;
  try {
    ctx = ctx || new AudioContext();
    const o = ctx.createOscillator();
    const g = ctx.createGain();
    o.frequency.value = freq;
    o.type = "triangle";
    g.gain.setValueAtTime(0.05, ctx.currentTime);
    g.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + dur);
    o.connect(g);
    g.connect(ctx.destination);
    o.start();
    o.stop(ctx.currentTime + dur);
  } catch {
    /* ignore */
  }
}
