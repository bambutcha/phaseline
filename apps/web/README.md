# Web (`apps/web`)

Сейчас: статическая заглушка (nginx :3000), чтобы `docker compose` поднимал Caddy.

**Дальше (фаза 0–4):** заменить на Next.js App Router + React + TypeScript + Tailwind.

```bash
# когда начнёте фронт, из этого каталога:
npx create-next-app@latest . --typescript --tailwind --app --src=false --eslint
```

Сохранить `app/play/[id]`, `app/s/[seed]` из [docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md).

Canvas 2D, Zustand, TanStack Query. Не Phaser.

`NEXT_PUBLIC_API_BASE` пустой в Docker — браузер ходит на `/api` и `/ws` через Caddy.
