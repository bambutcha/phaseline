# API (`apps/api`)

Go-сервис. Канон: [docs/ARCHITECTURE.md](../../docs/ARCHITECTURE.md), [docs/API.md](../../docs/API.md), [docs/SIM.md](../../docs/SIM.md).

```
cmd/server          точка входа (сейчас только /health /ready)
internal/sim        чистая симуляция — начинать отсюда
internal/game       (фаза 3) runtime партий
internal/http       (фаза 2) REST
internal/ws         (фаза 3) WebSocket
internal/seed       (фаза 1) генерация карты
internal/db         sqlc generate
db/migrations       goose
db/queries          sqlc SQL
db/schema.sql       схема для sqlc
sqlc.yaml
```

```bash
go test ./internal/sim/...
sqlc generate
goose -dir db/migrations postgres "$DATABASE_URL" up
go run ./cmd/server
```
