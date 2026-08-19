# API contract

Base: `/api/v1`  
Health вне версии: `GET /health`, `GET /ready`  
WS: `/ws/game/{id}`

Ошибки:

```json
{ "error": { "code": "not_found", "message": "game not found" } }
```

Коды: `invalid`, `not_found`, `conflict` (нельзя принять третий контракт), `game_over`.

---

## REST

### `POST /api/v1/games`

Создать партию.

```json
{ "seed": "MCC-7F2A", "rover": "swift" }
```

`seed` и `rover` опциональны. Rover по умолчанию `swift`. Значения: `swift` | `hauler`.

Ответ `201`:

```json
{
  "id": "uuid",
  "seed": "MCC-7F2A",
  "status": "lobby",
  "rover": "swift",
  "map": {
    "hexes": [
      { "q": 0, "r": 0, "type": "base", "id": "0,0" }
    ],
    "terminator": { "pos": 0, "speed": 0.015, "direction": "east" }
  },
  "contracts": [],
  "crisis": { "kind": "dust_storm", "firesAt": 96 },
  "autonomyCharges": 3
}
```

Превью публичного сида **не** отдаёт список контрактов (см. `GET /seeds/{seed}`). Snapshot партии включает `rovers`, `salvage`, `colonyScore`, `earthScore`.

### `GET /api/v1/games/{id}`

Полный snapshot (lobby / active / finished).

### `POST /api/v1/games/{id}/contracts/{cid}/accept`

Принять контракт, если слотов < 2 и статус `queued`.

### `POST /api/v1/games/{id}/route`

```json
{ "hexPath": ["0,0", "1,0", "1,1"] }
```

Сервер валидирует соседей и проходимость. Считает `predictedBattery`, `etaSec`, `inShadowAt`.

Ответ: тот же preview.

### `POST /api/v1/games/{id}/deploy`

Старт движения по текущему path. `lobby` → `active` при первом deploy.

### `POST /api/v1/games/{id}/reroute`

Как `route`, но во время `moving`. Применяет штраф, если бесплатный reroute уже использован.

### `POST /api/v1/games/{id}/autonomy`

Тратит заряд. Ответ: новый path или `{ "applied": false, "reason": "no_safer_path" }`.

### `GET /api/v1/games/{id}/blackbox`

Только если `status=finished`. Events + verdict + scores + ghost seed URL.

### `GET /api/v1/seeds/{seed}`

Публично, без спойлеров: `seed`, `layout`, размер карты, направление тени. **Не** контракты, не кризис, не кассеты.

---

## WebSocket

`GET /ws/game/{id}`

Client → server:

```json
{ "type": "select_rover", "rover": "swift" }
{ "type": "dispatch", "contractId": "c0", "rover": "hauler" }
{ "type": "goto", "hexId": "2,1", "rover": "swift" }
{ "type": "accept_contract", "contractId": "c0" }
{ "type": "set_route", "hexPath": ["0,0", "1,0"] }
{ "type": "deploy" }
{ "type": "reroute", "hexPath": [] }
{ "type": "autonomy" }
```

Игровой клиент ходит в основном через `dispatch` / `goto` / `select_rover`. REST и WS — одинаковые intent.

Server → client:

```json
{
  "type": "tick",
  "t": 12.4,
  "rover": {
    "hex": "1,0",
    "progress": 0.4,
    "battery": 0.62,
    "state": "moving",
    "cargo": []
  },
  "terminator": { "pos": 0.18, "direction": "east" },
  "contracts": [],
  "colonyScore": 0,
  "earthScore": 0,
  "events": [{ "t": 12.1, "kind": "entered_shadow", "payload": { "hexId": "1,0" } }]
}
```

`type: snapshot` — полный стейт после reconnect.  
`type: game_over` — финал, клиент открывает Black Box.

---

## События (`kind`)

`game_start`, `contract_accepted`, `dispatch`, `goto`, `hex_entered`, `entered_shadow`, `deliver`, `salvage`, `salvage_lost`, `contract_failed`, `crisis`, `stranded`, `game_over`.
