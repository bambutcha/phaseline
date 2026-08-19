# API (`apps/api`)

Go-сервис. Канон: [docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md), [docs/API.md](../../docs/API.md), [docs/SIM.md](../../docs/SIM.md).

```
cmd/server          точка входа, миграции, listen :8080
internal/sim        тик, A*, тень, грузы, кассеты
internal/game       runtime партий, 10 Гц
internal/http       REST + WebSocket
internal/db         sqlc generate
db/migrations       goose
db/queries          sqlc SQL
sqlc.yaml
```

```bash
go test ./...
sqlc generate
goose -dir db/migrations postgres "$DATABASE_URL" up
go run ./cmd/server
```

Канон запуска — `docker compose up --build` из корня.
