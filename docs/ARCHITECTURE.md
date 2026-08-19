# Architecture

## Обзор

```
браузер (Next.js)
  REST  /api/v1/*
  WS    /ws/game/{id}
       │
     Caddy :80
       ├─ /api/* /ws/* /health /ready → api:8080
       └─ *                            → web:3000
              │
         Go API (один бинарь)
           http/  Gin
           ws/    GameHub
           game/  tick-loop 10 Hz на активную партию
           sim/   чистая логика, без SQL/HTTP
              │
         PostgreSQL
           games, contracts, game_events, ghost_runs
```

**Server-authoritative.** Клиент рисует snapshot и шлёт команды. Читы на клиенте не меняют исход.

## Пакеты Go (`apps/api`)

```
cmd/server          — main, конфиг, миграции, listen
internal/http       — REST handlers, DTO
internal/ws         — upgrade, fan-out tick
internal/game       — жизненный цикл партии, очередь intent, тик
internal/sim        — карта, A*, батарея, грузы, терминатор
internal/seed       — детерминированная генерация из MCC-XXXX
internal/db         — sqlc output (после generate)
db/migrations       — goose
db/queries          — sqlc SQL
```

Правило: **HTTP не вызывает формулы напрямую.** Handler → `game` → `sim`.

## Тик

- 10 Hz (`dt = 0.1s`) на **активную** игру.
- Нет глобального тика пустых лобби.
- Snapshot на каждый тик + сразу после intent (deploy, reroute).
- События пишутся в память партии и flush в `game_events` пачками (не каждый тик одной INSERT, если можно batch).

## Сид

Формат: `MCC-` + 4 символа `[0-9A-Z]` без `O/I` (читаемость).  
Пустой seed на `POST /games` → сервер генерирует.

Один seed обязан давать:

- одинаковую карту (типы гексов, pickup/dropoff баз);
- одинаковые контракты и кассеты;
- одинаковый кризис и время срабатывания;
- одинаковое направление терминатора.

Тест: `NewGame(seed)` дважды → одинаковые map / contracts / salvage.

## Состояние партии (в памяти + DB)

Пока игра `active`, канон — RAM (`game.Runtime`).  
Postgres: создание, финал, события, ghost.

После `finished` runtime можно выгрузить.  
`GET /games/{id}` после финала читает из DB + blackbox.

## WebSocket

- URL: `/ws/game/{id}`
- Один клиент v1 (не мультиплеер). Несколько вкладок: последняя подписанная получает тики; не падать.
- Heartbeat ping/pong.
- Reconnect: клиент делает `GET /games/{id}` и снова WS — сервер шлёт полный snapshot, не дельту.

## Фронт

```
app/page.tsx              меню, New Game, seed
app/play/[id]/page.tsx    игра
app/s/[seed]/page.tsx     превью сида без спойлеров контрактов
stores/gameStore.ts       Zustand snapshot
lib/hex.ts                math, зеркало Go sim/hex.go
lib/api.ts                REST
hooks/useGameSession.ts   WebSocket + intents
components/game/GameCanvas.tsx
```

Canvas рисует: гексы, тень, ровер, ghost, preview path.  
DOM: HUD, карточки, модалки.

`lib/hex.ts` и `internal/sim/hex.go` должны сходиться на фикстурах из `docs/HEX.md`.

## Конфиг

| Переменная | Кто | Смысл |
|---|---|---|
| `DATABASE_URL` | api | pgx DSN |
| `PORT` | api | `:8080` |
| `LOG_LEVEL` | api | slog |
| `NEXT_PUBLIC_API_BASE` | web | пусто = same origin |

Браузер **никогда** не ходит на `localhost:8080` в Docker. Только `/api` и `/ws` через Caddy.

## Что сознательно нет

Отдельные сервисы, Redis pub/sub, Kafka, object storage, SSR карты, session-auth (v1 без логина).
