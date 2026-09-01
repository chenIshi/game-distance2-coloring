// The curated levels.
//
// Defined here rather than generated into a data file: every level is just
// (family, n, d, k) plus prose, and the solver in the page can work out
// everything else. That removes a build step and a generated JSON blob.
//
// The list is ordered as a path through the findings, not by size. Several
// levels come in pairs that differ by one colour or one vertex, because the
// contrast is the point.

export const LEVELS = [
  {
    id: "path-p3-k3",
    title: "Start here — three vertices, three colours",
    family: "path", n: 3, d: 2, k: 3,
    difficulty: "tutorial",
    summary: "A gentle first game. At distance 2 all three vertices block each other, so every colour gets used exactly once and you cannot go wrong. Click a circle, then a colour.",
  },
  {
    id: "path-p2-k1",
    title: "What losing looks like",
    family: "path", n: 2, d: 2, k: 1,
    difficulty: "tutorial",
    summary: "Two vertices, one colour. This is unwinnable for Alice by design — colour either vertex and the other has nothing left. Play it once to see how a trap is reported.",
  },
  {
    id: "path-p6-k4",
    title: "Path of 6, four colours",
    family: "path", n: 6, d: 2, k: 4,
    difficulty: "core",
    summary: "Four is exactly enough for a path this long. Three is not.",
  },
  {
    id: "cycle-c6-k4",
    title: "Cycle of 6, four colours",
    family: "cycle", n: 6, d: 2, k: 4,
    difficulty: "core",
    summary: "Not enough. A 6-cycle needs five colours at distance 2 — try to see why four fails.",
  },
  {
    id: "cycle-c7-k4",
    title: "Cycle of 7, four colours",
    family: "cycle", n: 7, d: 2, k: 4,
    difficulty: "core",
    summary: "The same four colours, one more vertex, and now Alice wins. Bigger is not always harder.",
  },
  {
    id: "wheel-w4-k3-d1",
    title: "Wheel of 4, three colours",
    family: "wheel", n: 4, d: 1, k: 3,
    difficulty: "core",
    summary: "Alice can win, but only with one particular opening. Colour the hub first and the rim is forced; start anywhere else and Bob wins.",
  },
  {
    id: "wheel-w8-k3-d1",
    title: "Wheel of 8, three colours",
    family: "wheel", n: 8, d: 1, k: 3,
    difficulty: "challenge",
    summary: "The same hub trick, a longer rim. It stops working here: Bob has room to make trouble far from wherever Alice just played.",
  },
  {
    id: "helm-h4-k3-d1",
    title: "Helm of 4, three colours",
    family: "helm", n: 4, d: 1, k: 3,
    difficulty: "challenge",
    summary: "A wheel of 4 with a stub on each rim vertex. The hub trick no longer saves Alice — watch how a rim vertex dies to its own stub.",
  },
  {
    id: "helm-h4-k4-d1",
    title: "Helm of 4, four colours",
    family: "helm", n: 4, d: 1, k: 4,
    difficulty: "core",
    summary: "One more colour and the same board is comfortable. That extra colour is exactly what the stubs cost.",
  },
  {
    id: "sunlet-s5-k3-d1",
    title: "Sunlet of 5, three colours",
    family: "sunlet", n: 5, d: 1, k: 3,
    difficulty: "core",
    summary: "A ring of 5 with a stub on every vertex. Odd rings are fine on three colours.",
  },
  {
    id: "sunlet-s4-k3-d1",
    title: "Sunlet of 4, three colours",
    family: "sunlet", n: 4, d: 1, k: 3,
    difficulty: "challenge",
    summary: "The even ring is harder, which is backwards from usual. Alice loses here mostly because she is forced to move first at all.",
  },
  {
    id: "sunlet-s4-k4-d1",
    title: "Sunlet of 4, four colours",
    family: "sunlet", n: 4, d: 1, k: 4,
    difficulty: "core",
    summary: "The same even ring with one more colour. Now her opening move is no longer fatal.",
  },
  {
    id: "wheel-w5-k6-d2",
    title: "Wheel of 5 at distance 2",
    family: "wheel", n: 5, d: 2, k: 6,
    difficulty: "tutorial",
    summary: "At distance 2 the hub puts every pair within reach of each other, so every vertex needs its own colour. Six vertices, six colours, no decisions to make.",
  },
  {
    id: "helm-h4-k5-d2",
    title: "Helm of 4 at distance 2",
    family: "helm", n: 4, d: 2, k: 5,
    difficulty: "challenge",
    summary: "Distance 2 on a helm. The hub and the whole rim now block each other, but the stubs stay far enough apart to share.",
  },
];

export function levelById(id) {
  return LEVELS.find((level) => level.id === id);
}
