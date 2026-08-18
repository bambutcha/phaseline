# Development — с чего начинать

Документы прочитаны в таком порядке: [GDD.md](./GDD.md) → этот файл → [ARCHITECTURE.md](./ARCHITECTURE.md) → [SIM.md](./SIM.md) / [HEX.md](./HEX.md) → [API.md](./API.md).

## Жёсткое правило

Не начинать с Canvas, A* «для красоты» или мета-прогрессии.  
Порядок фаз ниже обязателен.

## Фазы

| Фаза | Цель | Готово, когда |
|---|---|---|
| **0** | Compose + Caddy + `/health` + заглушка web | `docker compose up` открывает страницу и health |
| **1** | `internal/sim` + hex + seed gen + тесты | `go test ./internal/sim/...` зелёный, детерминизм сида |
| **2** | goose + sqlc + `POST /games` | Postman/curl создаёт игру, карта в JSON |
| **3** | Tick + WS + deploy | ровер едет, батарея падает, тень ползёт (хоть без сока) |
| **4** | Canvas + mobile layout | playable на телефоне и PC |
| **5** | Контракты, грузы, кризис | правила из SIM.md |
| **6** | Ghost + `/s/[seed]` + Black Box | реиграбельность |
| **7** | Juice, звук, туториал | wow |
| **8** | README, скриншоты, AI-секция | сдача |

## Команды (когда код появится)

```bash
docker compose up --build    # :80
cd apps/api && go test ./...
cd apps/api && sqlc generate
cd apps/api && goose -dir db/migrations postgres "$DATABASE_URL" up
```

Локально без Docker: Postgres 16, `DATABASE_URL` из `.env.example`, API `:8080`, Next `:3000`. Для сдачи канон — Compose + Caddy.

## Конвенции

**Go**

- Бизнес-логика только в `sim` / `game`.
- Не AutoMigrate. Только goose.
- sqlc, не GORM.
- Тесты сида: одинаковый seed → байт-в-байт одинаковый map JSON (или `cmp.Equal`).

**TS**

- `lib/hex.ts` зеркалит Go. Сначала фикстуры из HEX.md.
- Zustand хранит server snapshot, не пересчитывает батарею «примерно».
- Canvas 2D, не Phaser.

**Git**

- Коммиты по фазам, не «wip всё».
- Не коммитить `.env`.

## AI usage (заполнить в README при сдаче)

Писать честно:

- что сгенерировано (boilerplate, UI);
- что руками (формулы, баланс, тесты);
- как проверяли (go test, playtest на телефоне).

## Стек — не трогать без причины

Go + Gin + pgx + sqlc + Postgres + WS + Next + Tailwind + Zustand + TanStack Query + Caddy + Compose.

Redis / Kafka / MinIO — нет.

## Чеклист перед первой строкой UI

- [ ] `sim.Tick` двигает терминатор
- [ ] тень в axial, не pixelX
- [ ] MoveCost совпадает в preview и tick
- [ ] cubeRound покрыт тестом
- [ ] `GET /health` 200
