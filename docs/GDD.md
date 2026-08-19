# PHASELINE — Game Design Document

**Подзаголовок:** Deliver before the line.  
**Репозиторий:** `phaseline`  
**Жанр:** браузерная стратегия / puzzle-logistics / sci-fi micro-sim  
**Платформы:** mobile-first web + desktop web  
**Контекст:** тестовое Fullstack Middle (Moon Courier Crisis) + авторская интерпретация  
**Скоуп кодинга:** только v1 (MVP). Product Vision — в конце, не реализовывать до сдачи.

Связанные документы:

- [ARCHITECTURE.md](./ARCHITECTURE.md) — сервер, WS, пакеты
- [API.md](./API.md) — REST + WebSocket контракт
- [HEX.md](./HEX.md) — axial/cube, A*, pixel
- [SIM.md](./SIM.md) — формулы тика, батарея, грузы
- [COPY.md](./COPY.md) — туториал и вердикты Black Box
- [DEVELOPMENT.md](./DEVELOPMENT.md) — фазы разработки и порядок работы

---

## 1. Elevator pitch

На Луне ползёт **линия фазы** — граница дня и вечной ночи. Ты последний курьер колонии. Заказов больше, чем батареи. Спасти всех нельзя.

**PHASELINE** — 3–5 минут triage: выбираешь, кого спасаешь, прокладываешь маршрут по гекс-карте, смотришь, как тень съедает мир, и пытаешься обойти **призрак своего прошлого забега** на том же сиде.

> Не гонка на реакцию — гонка на решение. Тень не ждёт.

**Команды без задержки.** Клик → маршрут → ровер едет сразу. Напряжение от тени, веса и triage, не от input lag.

---

## 2. Название

| Критерий | PHASELINE |
|---|---|
| Форма | одно слово, 9 букв |
| Смысл | phase line / терминатор Луны |
| UI | `Phase Line ETA: 42s` |
| Сиды | `MCC-XXXX` (Moon Courier Crisis + seed) |

Альтернативы (не использовать, пока не решим сменить бренд): `REGOLITH`, `TERMINUS`, `ASHLINE`.

---

## 3. USP

1. **Тень — антагонист**, не таймер в углу: гасит клетки, меняет риск, убивает заказы.
2. **Triage, не оптимизация:** 5–7 контрактов, унести 2–3. Каждый забег — выбор.
3. **Ghost Run:** полупрозрачный «ты из прошлого» на том же сиде. Без мультиплеера.
4. **Share Seed:** ` /s/MCC-7F2A ` — тот же мир, те же заказы, тот же кризис.
5. **Black Box:** реплей + лог + вердикт колонии, не `Score: 340`.
6. **Мгновенный отклик.**
7. **Mobile + PC parity** — один UX, разные раскладки.

---

## 4. Core loop (3–5 минут)

```
сид / новый забег
  → брифинг: карта, контракты, скорость тени, кризис
  → выбор ровера (Swift | Hauler)
  → принять 1–2 контракта
  → маршрут (тап/клик по гексам)
  → DEPLOY → ровер едет в реальном времени
  → delivered / failed / stranded / lost_to_shadow
  → следующий контракт, пока база жива
  → буря или поглощение картой → Black Box → share link
```

---

## 5. Карта

- **Сетка:** pointy-top hex, **11–13 гексов** (компактно для телефона).
- **Координаты:** axial `q, r`. ID: `"q,r"` например `"3,-1"`.
- **Генерация:** детерминированный seed `MCC-XXXX`. Один seed = одна карта, один набор контрактов, один кризис, одно направление тени.

| Тип | Эффект | Визуал |
|---|---|---|
| Regolith | база | серая пыль |
| Solar Plateau | зарядка на солнце | золото |
| Crater | +риск, +время | провал |
| Ridge | +расход батареи | рельеф |
| Base Alpha | старт / сдача | огни |
| Cold Sink | быстрый разряд в тени | синий |
| Dust Field | drift при кризисе пыли | частицы |

Математика: [HEX.md](./HEX.md).  
Тень считается **в axial-проекции**, не в пикселях канваса.

---

## 6. Terminator (линия фазы)

- Движется непрерывно. Направление из сида.
- Клетка в тени:
  - **непроходима** — войти нельзя;
  - зарядка солнечных панелей отключена;
  - если тень накрыла ровер на клетке — stranded;
  - дедлайны и незабранные пины сгорают;
  - визуал: тускнеет, крест на гексе.
- **Phase ETA** на гекс — обязательный HUD: «до тени: 38s».
- Формулы: [SIM.md](./SIM.md).

---

## 7. Ровер (v1: 2 loadout)

| | **Swift** | **Hauler** |
|---|---|---|
| Груз | 1 лёгкий + 1 средний | до 2 тяжёлых |
| Скорость | выше | ниже |
| Батарея | меньше | больше |

Оба ровера на смене сразу. Выбор активного — кнопка Swift/Hauler, без задержки связи.

---

## 8. Контракты

5–7 на забег. Активных одновременно: **максимум 2**.

```
id, title, cargoType, weight (light|medium|heavy),
pickup HexId, dropoff HexId,
colonyValue, earthValue,
deadline? (sec), specialRule
```

| Груз | Правило |
|---|---|
| O₂ Filters | в тени батарея жрёт сильнее |
| Cryo Samples | idle на солнце → порча |
| Crew Pod | crater/ridge → panic (медленнее) |
| Reactor Rod | reroute после бесплатного → overheat |
| Helium-3 | много earth, мало colony |
| Med Seeds | жёсткий deadline |
| Comm Relay | бонус +8 колонии, если сдан до кризиса |

На карте ещё **3 кассеты**: +8 колонии за заход на клетку, без сдачи и без слота. Тень забирает неподобранные. Если заказы кончились — это запасной заработок до конца смены.

**Статусы:** `queued → accepted → in_transit → delivered | failed | expired | lost_to_shadow`

Имена и тексты: [COPY.md](./COPY.md).

---

## 9. Маршрут

- Тап/клик → A* только по проходимым клеткам.
- Preview: пунктир, ETA, **predicted battery** (тот же код, что tick).
- Predicted < 0 → маршрут красный, DEPLOY всё равно можно.
- **Reroute:** 1 раз бесплатно, дальше −15% батареи (Reactor штрафует сильнее).
- Mobile: undo кнопкой, не только right-click.
- **Команды без задержки.**

---

## 10. Кризис (1 на забег, из сида)

Срабатывает примерно на **40%** расчётной длины миссии.

| Кризис | Эффект |
|---|---|
| Dust Storm | drift на Dust Field |
| Solar Flare | plateau опасен, зарядка скачет |
| Cave-in | один crater становится impassable |
| VIP Override | новый срочный контракт mid-run |
| Comm Blackout | ghost скрыт ~15s, HUD упрощён |

---

## 11. Ghost Run

После забега на сиде сохраняется ghost (последняя попытка **или** лучший colony score — решить в Phase 6, по умолчанию **лучший colony**).

На карте: полупрозрачный ровер + след. **Не влияет на физику.**

Данные: `ghost_runs.replay_json` + `game_events`.

---

## 12. Победа

**Colony Score** — главный порог.  
**Earth Score** — вторичный.

| Исход | Условие |
|---|---|
| 🟢 Colony Saved | colony ≥ threshold (100) |
| 🟡 Pyrrhic Victory | colony ≥ threshold, earth низкий |
| 🔴 Signal Lost | colony < threshold |

Тексты вердиктов: [COPY.md](./COPY.md).  
Не показывать голый `score=340` как главный итог.

---

## 13. Autonomy (3 заряда)

Кнопка **AUTONOMY**: мгновенный микро-обход (до 2 гексов), если новый path безопаснее (меньше crater/shadow). Иначе no-op.

Тема: локальный AI ровера. В README можно честно указать: алгоритм, не LLM в рантайме v1.

---

## 14. Wow (обязательный juice v1)

- Earth в небе, parallax
- Terminator: gradient + пыль + появление звёзд
- Rover: dust trail, фары в тени, bounce на ridge
- Delivery: pulse ring
- Battery low: vignette
- Lost to shadow: клетка «замерзает»
- Black Box: cinematic scroll лога
- Звук: ветер, click, rising tone тени, thud+ping доставки, mute всегда виден
- `prefers-reduced-motion` отключает shake/particles
- Haptic на mobile при deliver (если Vibration API доступен)

**Не в v1:** WebGL, Phaser, 3D, Pixi.

---

## 15. UX

### Mobile portrait (primary)

```
┌─────────────────────┐
│  Phase ETA | Battery│
├─────────────────────┤
│      HEX MAP        │  ~70vh
├─────────────────────┤
│ Contracts carousel  │
├─────────────────────┤
│ [Undo] [DEPLOY]     │  thumb zone
└─────────────────────┘
```

### Desktop landscape

Карта слева, контракты / батарея / лог / DEPLOY справа.

| Действие | Mobile | Desktop |
|---|---|---|
| Гекс | tap | click |
| Undo | кнопка | right-click / Backspace |
| Deploy | большая кнопка | кнопка / Space |
| Контракт | swipe card | click list |
| Autonomy | FAB | `A` |

Touch target ≥ 44×44. `touch-action: manipulation`. Без pinch-zoom в v1.

**Перф:** 60 FPS mid-range phone, Canvas 2D, first load < 3s на 4G.

PWA (manifest + иконки) — nice-to-have, не блокер MVP.

---

## 16. Стек (зафиксирован)

| Слой | Выбор |
|---|---|
| Backend | Go, Gin, **pgx + sqlc**, goose, WebSocket, slog |
| Frontend | Next.js App Router, React, TypeScript, Tailwind, Zustand, TanStack Query, Canvas 2D |
| DB | PostgreSQL 16 |
| DevOps | Docker Compose, Caddy |

**Не используем в v1:** GORM, Redis, Kafka, MinIO, Kubernetes, gRPC, микросервисы, Phaser.

---

## 17. MVP scope

### В v1

- [x] Seed-карта + 5–7 контрактов
- [x] Swift / Hauler
- [x] Terminator real-time + Phase ETA
- [x] Preview + deploy + reroute
- [x] Грузы со special rules
- [x] 1 кризис на забег
- [x] Ghost на сиде
- [x] Share `/s/[seed]`
- [x] Black Box
- [x] Кассеты (запасной заработок)
- [x] Mobile + desktop
- [x] REST + WebSocket
- [x] `docker compose up` → игра на :80
- [x] Юнит-тесты `internal/sim`
- [x] README: запуск, AI usage, решения

### Не в v1

Мультиплеер, глобальный лидерборд, Redis/Kafka/MinIO, 3D, редактор карт, магазин, оффлайн-игра, LLM в рантайме, мета-прокачка, F2P-энергия, сезон-пасс, «What if» rewind.

---

## 18. Критерии приёмки

1. `docker compose up --build` открывает игру на телефоне и PC (через хост :80).
2. Новый игрок понимает цель без README за ≤ 30 сек (4 слайда, skip).
3. Забег 3–5 минут.
4. Три разных сида ощущаются по-разному.
5. Повтор того же сида показывает ghost.
6. Share seed открывает играбельный сид.
7. Black Box: лог + вердикт.
8. README честно описывает AI.
9. Скриншоты: [desktop](../screenshots/desktop.png), [mobile](../screenshots/mobile.png), [tutorial](../screenshots/tutorial.png).

---

## 19. Туториал

4 слайда, полный текст в [COPY.md](./COPY.md). Skip в `localStorage`.

---

## 20. Баланс (стартовые константы)

Актуальные числа — `apps/api/internal/sim/constants.go` (тюнить после playtest).

```yaml
tick_hz: 10
game_duration_target_sec: 180
rover_swift_speed_hex_per_sec: 0.52
rover_hauler_speed_hex_per_sec: 0.36
battery_swift: 80
battery_hauler: 105
colony_win_threshold: 100
earth_pyrrhic_below: 40
autonomy_charges: 3
reroute_free_count: 1
max_active_contracts: 2
salvage_count: 3
salvage_value: 8
map_hex_count: 30   # 6×5
```

---

## 21. Product Vision (v2+, не кодить)

- Daily seed + ghost-challenge по ссылке
- Косметика следа / палитры HUD
- Третий chassis

Без paywall на забеги и без мета-апгрейдов «+10% батареи» до сдачи тестового.

---

## 22. Принципы, которые нельзя ломать

1. Сервер — источник правды. Клиент шлёт intent.
2. Тень и стоимость хода считаются в гексах, не в пикселях.
3. Preview батареи = тот же код, что tick.
4. Один бинарь API. Нет сервиса заказов / сервиса роверов.
5. Сначала `sim/` зелёный, потом Canvas.
6. Мобилка — основной layout, desktop — расширение.
7. Скоуп v1 не раздувать «ещё одной гениальной фичей» до работающего забега.
