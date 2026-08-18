# Hex math (pointy-top, axial)

Источник правды для Go (`internal/sim/hex.go`) и TS (`lib/hex.ts`).  
Пиксели — только рендер и hit-test. Симуляция тени — axial.

## Axial и cube

```
axial: (q, r)
cube:  x = q,  z = r,  y = -x - z
```

**Расстояние:**

```
dist(a, b) = (|aq-bq| + |aq+ar - bq-br| + |ar-br|) / 2
```

**ID:** `"{q},{r}"` — `3,-1`.

## Соседи

```
(+1, 0), (+1, -1), (0, -1), (-1, 0), (-1, +1), (0, +1)
```

## Hex → pixel (центры)

`size` = радиус (центр → угол).

```
x = size * (√3 * q + √3/2 * r)
y = size * (3/2 * r)
```

## Pixel → hex

```
q = (√3/3 * x - 1/3 * y) / size
r = (2/3 * y) / size
→ cubeRound → axial
```

### cubeRound

1. `rx, ry, rz = round(x,y,z)`
2. Посчитать ошибки `|rx-x|` и т.д.
3. Компонент с **максимальной** ошибкой восстановить: он равен `-два других`.

Без cubeRound тапы будут промахиваться.

## Тень (не pixelX)

Терминатор — скаляр `pos` в проекции на ось направления сида.

Пример direction `east`: клетка в тени, если `q < pos` (уточнить знак при генерации карты; зафиксировать тестом).

`PhaseETA(hex) = (project(hex) - pos) / speed` если ещё на свету, иначе `0`.

## A*

- Граф: 6 соседей, skip `impassable`.
- `g` += `MoveCost(hex, rover, inShadowAtArrival)` — shadow на момент **прибытия**, не на момент клика, если preview считает будущее терминатора.
- Heuristic: `dist * minBaseCost`.
- Preview обязан вызывать тот же `MoveCost`, что tick.

## Фикстуры для золотых тестов

| q | r | pixel (size=1) x,y примерно |
|---|---|---|
| 0 | 0 | 0, 0 |
| 1 | 0 | 1.732, 0 |
| 0 | 1 | 0.866, 1.5 |

`dist((0,0),(1,-1)) = 1`  
`neighbors((0,0))` длина 6.
