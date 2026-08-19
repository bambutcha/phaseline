# Web (`apps/web`)

Next.js 15 App Router + React 19 + TypeScript + Tailwind + Zustand + TanStack Query. Canvas 2D, не Phaser.

Канон запуска — `docker compose up --build` из корня репозитория. Образ сам делает `npm ci` и `next build`.

Локально без Docker:

```bash
npm test          # hex fixtures
npm run dev       # :3000, REST проксируется на API :8080
npm run build
```

В Docker браузер ходит только через Caddy: `/api/*`, `/ws/*`, `/health`. `NEXT_PUBLIC_API_BASE` пустой.

Маршруты:

- `/` — меню, новая смена, сид
- `/play/[id]` — смена
- `/s/[seed]` — шаринг без спойлеров контрактов
