import assert from "node:assert/strict";
import test from "node:test";
import { cubeDistance, hexToPixel, NEIGHBORS } from "./hex";

test("hexToPixel fixtures size=1", () => {
  const a = hexToPixel(0, 0, 1);
  assert.ok(Math.abs(a.x) < 1e-9 && Math.abs(a.y) < 1e-9);
  const b = hexToPixel(1, 0, 1);
  assert.ok(Math.abs(b.x - Math.sqrt(3)) < 1e-6 && Math.abs(b.y) < 1e-9);
  const c = hexToPixel(0, 1, 1);
  assert.ok(Math.abs(c.x - Math.sqrt(3) / 2) < 1e-6 && Math.abs(c.y - 1.5) < 1e-9);
});

test("cubeDistance (0,0) to (1,-1) is 1", () => {
  assert.equal(cubeDistance({ q: 0, r: 0 }, { q: 1, r: -1 }), 1);
});

test("neighbors length 6", () => {
  assert.equal(NEIGHBORS.length, 6);
});
