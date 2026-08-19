# PHASELINE — agent notes

Workspace: `/home/bambutcha/Files/Knowledge/IT/projects/Teams/Exp/phaseline`

Это тестовое fullstack (Moon Courier Crisis) как игра **PHASELINE**. Канон: `docs/GDD.md` и `docs/DEVELOPMENT.md`.

## Делать

- Server-authoritative симуляция в `apps/api/internal/sim`
- Тень в axial-координатах, не в пикселях
- Preview батареи = тот же код, что tick
- Mobile-first, Canvas 2D
- pgx + sqlc + goose, Gin, Postgres, Caddy, Compose
- Next.js + Tailwind + Zustand + TanStack Query, Canvas 2D
- Команды игрока **без** задержки связи

## Не делать в v1

Redis, Kafka, MinIO, GORM, Phaser/WebGL, мультиплеер, мета-прокачка, F2P-лимиты, LLM в рантайме.

## Порядок работ

Phase 0 health/compose → 1 sim tests → 2 REST create game → 3 WS tick → 4 canvas → 5 cargo/crisis → 6 ghost/blackbox → 7 juice → 8 README.

Не начинать с UI, пока `go test ./internal/sim/...` не зелёный на тике и сиде.
