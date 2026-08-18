# PHASELINE

**Deliver before the line.**

Браузерная стратегия про лунную доставку: линия фазы ползёт по гекс-карте, спасти все контракты нельзя. Стек и правила зафиксированы — это репозиторий тестового задания Fullstack (Moon Courier Crisis) с авторской задумкой.

Игра ещё не реализована. Здесь лежат GDD, контракты API, схема БД и каркас, с которого начинать код.

## Документы (читать в этом порядке)

1. [docs/GDD.md](docs/GDD.md) — дизайн игры, скоуп v1
2. [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) — фазы 0–8, с чего писать код
3. [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — сервер, WS, пакеты
4. [docs/SIM.md](docs/SIM.md) — формулы тика
5. [docs/HEX.md](docs/HEX.md) — axial / cube / A*
6. [docs/API.md](docs/API.md) — REST + WebSocket
7. [docs/COPY.md](docs/COPY.md) — туториал и вердикты Black Box

## Стек

| Слой | |
|---|---|
| Backend | Go, Gin, pgx, sqlc, goose, WebSocket |
| Frontend | Next.js (пока заглушка), React, TypeScript, Tailwind, Zustand, Canvas 2D |
| DB | PostgreSQL 16 |
| DevOps | Docker Compose, Caddy |

Нет Redis, Kafka, MinIO, GORM.

## Запуск каркаса

```bash
cp .env.example .env
docker compose up --build
```

- UI-заглушка: http://localhost/
- Health: http://localhost/health

Симуляция без UI:

```bash
cd apps/api && go test ./internal/sim/...
```

## Структура

```
apps/api     Go API + internal/sim
apps/web     заглушка → позже Next.js
deploy/      Caddyfile
docs/        GDD и контракты
screenshots/ для сдачи
```

## Сдача (когда игра будет готова)

- Ссылка на git
- Скриншоты: mobile, desktop, тень, Black Box
- README: запуск, что сделано, логика, где данные, **что сделано с AI**

Шаблон секций для финального README — в конце GDD / DEVELOPMENT.

## Принцип разработки

Сначала `sim/` и тесты, потом REST, потом WebSocket, потом Canvas. Не наоборот.
