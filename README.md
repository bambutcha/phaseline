# PHASELINE

**Deliver before the line.**

Браузерная игра для тестового задания **Moon Courier Crisis**. На лунной гекс-карте ползёт терминатор фазы: спасти все заказы нельзя. Нужно набрать **80 очков колонии** за смену (~**90 секунд**). Если цель набрана раньше, смена **не обрывается** — считаешь в конце, по Black Box.

Два ровера сразу:

- **Swift** — быстрый, не берёт тяжёлое;
- **Hauler** — медленнее, везёт стержни и Helium-3.

## Соответствие ТЗ

| Требование вакансии | Где |
|---|---|
| Карта Луны | axial-гексы, `apps/api/internal/sim/seed.go` |
| Пины заказов | Canvas + карточки с номерами |
| Заказ: вес / награда / срочность / риск | `Contract` в snapshot |
| Ровер: батарея / слоты / статус | Swift + Hauler, HUD |
| Выбрать заказ и ровер, запустить | клик по карточке / пину |
| Postgres | `games`, `contracts`, `game_events`, `ghost_runs` |
| Тяжёлый груз влияет на доставку | `WeightMod`, Swift reject |
| Отказ, если батареи или слотов не хватает | preview + `reject` |
| Зоны карты отличаются | crater / ridge / cold_sink / solar / base |
| Хотя бы одна невыполнимая доставка | контракт `impossible` + тяжёлое для Swift |
| Счёт обновляется | colony / earth после сдачи |
| Цель победы | 80 колонии; пиррова если земля < 30 |

Авторская оболочка (PHASELINE): терминатор как антагонист, triage, ghost прошлого забега на том же сиде, share `/s/MCC-XXXX`, Black Box.

## Запуск

```bash
docker compose up --build
```

- Игра: http://localhost/
- Health: http://localhost/health
- API напрямую: `:8080`

Postgres **не** торчит на хост `:5432`.

```bash
cd apps/api && go test ./...
```

## Как играть

1. Туториал на первом заходе вкладки.
2. Нажми заказ внизу. Цифра на карте — забор. Ровер сам сдаёт.
3. Тяжёлое — только **Hauler** (кнопка сверху).
4. На ходу клик по другой клетке: ровер **остаётся на месте**, доезжает текущую клетку, затем сворачивает. Не откатывается на базу.
5. Красный путь = батареи не хватит. На золотой базе зарядка.
6. Повтор сида показывает полупрозрачного **призрака** прошлого тебя.

Поделиться: `http://localhost/s/MCC-XXXX`.

## Стек

- Go, Gin, pgx, sqlc, goose, WebSocket
- PostgreSQL 16
- Canvas 2D в nginx (Next.js в v1 нет: один HTML, mobile-first)
- Caddy, Docker Compose

Без Redis, Kafka, MinIO, GORM, Phaser.

## Структура

```text
apps/api                    API + server-authoritative sim
apps/api/internal/sim       тик, A*, тень в axial, грузы
apps/web/public/index.html  Canvas-клиент
deploy/Caddyfile            /api /ws /health /s/:seed
docs/                       GDD и спецификации
screenshots/                mobile, desktop, тень, Black Box
```

## Скриншоты

- [desktop](screenshots/desktop.png) — карта, пины, заказы
- [mobile](screenshots/mobile.png) — узкий layout
- [tutorial](screenshots/tutorial.png) — обучение на старте

## Использование AI

- **Сгенерировано:** Compose/Caddy boilerplate, большая часть Canvas UI, sqlc, черновики copy.
- **Спроектировано вручную:** формулы тика (`MoveCost`, батарея, тень в axial), детерминизм seed, два ровера, reroute с фиксацией текущего ребра, порог конца смены, тесты.
- **Проверено:** `go test ./...`, playtest `docker compose up --build`.
