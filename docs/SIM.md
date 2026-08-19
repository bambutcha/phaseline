# Simulation formulas

Пакет: `apps/api/internal/sim`  
Tick: 10 Hz, `dt = 0.1`  
Без SQL, без Gin.

## Порядок тика

1. Сдвинуть terminator (`pos += speed * dt`).
2. Пересчитать `inShadow` каждой клетки.
3. Если rover `moving` — сдвинуть `progress`, при `>= 1` commit гекса.
4. Idle drain / solar gain.
5. Cargo rules (cryo idle, panic timeout).
6. Deadlines контрактов.
7. Кризис, если `t >= firesAt` и ещё не сработал.
8. Кассеты: подбор / потеря в тени.
9. Проверка stranded / карта в тени / нечего везти / таймер.
10. Собрать events.

## MoveCost

```
MoveCost = Base(type) * WeightMod * TerrainMod * ShadowMod
impassable → +Inf
```

| Type | Base |
|---|---|
| base | 0.8 |
| regolith | 1.0 |
| solar_plateau | 0.9 |
| dust_field | 1.1 |
| cold_sink | 1.2 |
| crater | 1.4 |
| ridge | 1.5 |

**WeightMod** по сумме активного груза:

- только light → 1.00
- medium в сумме → 1.15
- есть heavy → 1.35

**TerrainMod:** Swift на ridge ×1.1; Hauler на crater ×1.1; иначе 1.0.

**ShadowMod:** `inShadow` → клетка непроходима (`+Inf`). В тень войти нельзя.

## Батарея

```
on hex enter:  battery -= MoveCost
each tick:     battery -= idleDrain * dt * (inShadow ? 2 : 1)
if sun && solar_plateau && not flare-danger:
               battery += solarGain * dt
clamp 0 .. maxBattery
if battery <= 0 && was moving → stranded
```

| | Swift | Hauler |
|---|---|---|
| maxBattery | 80 | 105 |
| idleDrain / s | 1.38 | 1.05 |
| speed hex/s | 0.52 | 0.36 |

`solarGain = 3 / s` на solar plateau. На базе, если стоишь: `baseRecharge = 4.6 / s`.  
Ход жрёт `MoveCost × MoveDrainMult` (`MoveDrainMult = 1.72`).

Движение: `progress += speed / MoveCost * dt` — либо эквивалент: время на ребро = `MoveCost / speed`. Зафиксировать **один** вариант в коде и тестах. Рекомендация:

```
edgeTime = MoveCost / rover.Speed
progress += dt / edgeTime
```

## Cargo rules

| Груз | Когда | Эффект |
|---|---|---|
| O₂ | уже в ShadowMod + idle×2 | fail если stranded с грузом |
| Cryo | idle ≥ 1s на солнце | status failed (spoil) |
| Crew | вход в crater/ridge | panic: speed ×0.7 до следующего гекса |
| Reactor | reroute после free | battery −15 (overheat) |
| Helium-3 | scoring | earthValue высокий |
| Med Seeds | deadline ≤ 0 | expired |
| Comm Relay | deliver до crisis | +bonus colony |

## Кассеты

3 штуки из сида, не на базе и не на pickup. Заход на клетку: `+8` colony, слот не занимает. Тень: статус `lost`. Смена не закрывается, пока есть живые заказы, живые кассеты или ещё не сработал VIP-кризис.

## Preview

`Predict(path)` прогоняет **тот же** Tick в копии стейта (без записи events в DB). Красный path, если `battery < 0` в любой точке **или** predicted end battery < 0.

## Кризис

`firesAt = 0.28 * duration` сек по умолчанию, плюс джиттер из сида в диапазоне ±10s, но детерминированный.

## Порог победы

`colony_win_threshold = 100`  
Pyrrhic: colony ≥ 100 и earth < 40.  
Длительность смены: 180 секунд. Тень непроходима: в погасший гекс войти нельзя, накрытие ровера = stranded. Пины забора уникальны. Смена не кончается от набранных 100 очков. Если живых заказов нет, остаются кассеты (+8 колонии за заход, без сдачи). Смена закрывается, когда нечего везти и кассет не осталось. 100 колонии достижимы (потолок сдаваемых заказов ≥ 100, dual-rover solver в тестах).
