// The page. All game reasoning happens in the Go solver compiled to
// WebAssembly; nothing here knows the rules.
//
// This file used to contain a hand-written JavaScript reimplementation of the
// solver, kept in step with the Python and Go versions by hand. That is exactly
// the kind of duplication that drifts silently, so it is gone.
//
// The one cost of moving the solver into a worker is that every question about
// the position is asynchronous. The page cannot ask "who is winning?" while
// rendering; it asks, waits, stores the answer in state.analysis, then renders.

import { analyzeState, describeGraph, chromaticNumber } from "./solver-client.js";
import { positionsFor, vertexLabel, FAMILIES, orderOf } from "./layout.js";
import { LEVELS, levelById } from "./levels.js";

const SWATCHES = ["#4c956c", "#d68c45", "#8d5a97", "#2d728f", "#c75146", "#3f7d8c", "#a8763e", "#6b5b95", "#417b5a", "#b5533c", "#57606f", "#8d6e63"];
const STORAGE_KEY = "distance-d-game-progress";
const INTRO_KEY = "distance-d-game-seen-intro";

const state = {
  mode: "levels",
  spec: null,
  graph: null,
  positions: [],
  colors: [],
  owners: [],
  humanRole: "Alice",
  currentPlayer: "Alice",
  analysis: null,
  thinking: false,
  selected: null,
  message: "",
  progress: {},
  openDrawer: null,
  moveToken: 0,
  // Captured from the empty board before anyone moves: whether the position is
  // winnable at all from the player's side, and how forgiving it is. Without
  // this, losing a deliberately unwinnable level is indistinguishable from
  // playing badly, and the player has no idea whether to try again.
  verdict: null,
};

const $ = (id) => document.getElementById(id);
const el = {
  title: $("case-title"), summary: $("case-summary"), chip: $("progress-chip"),
  turn: $("turn-label"), stateLine: $("state-label"), board: $("graph-board"),
  colors: $("color-buttons"), selection: $("selection-label"), message: $("message-label"),
  explain: $("explain-panel"), overlay: $("result-overlay"), overlayEyebrow: $("overlay-eyebrow"),
  verdict: $("verdict-line"), intro: $("intro"),
  overlayTitle: $("overlay-title"), overlayBody: $("overlay-body"),
  levelList: $("level-list"), backdrop: $("drawer-backdrop"),
  sandboxFamily: $("sandbox-family"), sandboxSize: $("sandbox-size"),
  sandboxSizeLabel: $("sandbox-size-label"), sandboxDistance: $("sandbox-distance"),
  sandboxColors: $("sandbox-colors"), sandboxInfo: $("sandbox-info"),
};

// ---------------------------------------------------------------- progress

function loadProgress() {
  try {
    state.progress = JSON.parse(window.localStorage.getItem(STORAGE_KEY) || "{}");
  } catch {
    state.progress = {};
  }
}

function saveProgress() {
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(state.progress));
  } catch {
    // A private window or blocked storage is fine; progress is a convenience.
  }
}

const progressKey = (id, role) => `${id}:${role}`;
const isDone = (id, role) => Boolean(state.progress[progressKey(id, role)]);

// ---------------------------------------------------------------- loading

async function loadSpec(spec) {
  state.spec = spec;
  state.selected = null;
  state.message = "";
  state.analysis = null;
  state.thinking = true;
  el.title.textContent = spec.title;
  el.summary.textContent = spec.summary || "";
  render();

  try {
    state.graph = await describeGraph(spec.family, spec.n, spec.d);
  } catch (error) {
    state.thinking = false;
    el.title.textContent = "Could not build that graph";
    el.summary.textContent = String(error.message || error);
    return;
  }

  state.positions = positionsFor(spec.family, spec.n);
  state.colors = new Array(state.graph.order).fill(0);
  state.owners = new Array(state.graph.order).fill(null);
  state.currentPlayer = "Alice";
  state.verdict = await openingVerdict(spec, state.graph.order);
  await refresh();
}

// openingVerdict asks the solver about the untouched board. optimalMoves there
// are exactly Alice's winning openings, so counting the distinct vertices says
// how much room for error the level allows.
async function openingVerdict(spec, order) {
  try {
    const start = await analyzeState(spec.family, spec.n, spec.d, spec.k,
      new Array(order).fill(0), true);
    const vertices = new Set(start.optimalMoves.map((m) => m.vertex));
    return { aliceWins: start.aliceWins, winningOpenings: vertices.size, order };
  } catch {
    return null;
  }
}

// Whether the side the player has chosen can win at all, with perfect play.
function winnableForPlayer() {
  if (!state.verdict) return null;
  return state.humanRole === "Alice" ? state.verdict.aliceWins : !state.verdict.aliceWins;
}

// refresh asks the solver about the current position, then lets the computer
// move if it is its turn. The token guards against a stale reply landing after
// the player has restarted or switched level.
async function refresh() {
  const token = ++state.moveToken;
  state.thinking = true;
  render();

  let analysis;
  try {
    analysis = await analyzeState(
      state.spec.family, state.spec.n, state.spec.d, state.spec.k,
      state.colors, state.currentPlayer === "Alice",
    );
  } catch (error) {
    if (token !== state.moveToken) return;
    state.thinking = false;
    state.message = `Solver error: ${error.message || error}`;
    render();
    return;
  }
  if (token !== state.moveToken) return;

  state.analysis = analysis;
  state.thinking = false;
  recordCompletion();
  render();

  if (!isOver() && state.currentPlayer !== state.humanRole) {
    window.setTimeout(() => computerMove(token), 350);
  }
}

const isOver = () =>
  Boolean(state.analysis) && (state.analysis.finished || state.analysis.dead.length > 0);

function humanWon() {
  if (!state.analysis) return false;
  if (state.analysis.finished) return state.humanRole === "Alice";
  if (state.analysis.dead.length > 0) return state.humanRole === "Bob";
  return false;
}

function recordCompletion() {
  if (!state.spec.id || !isOver() || !humanWon()) return;
  state.progress[progressKey(state.spec.id, state.humanRole)] = true;
  saveProgress();
}

// ---------------------------------------------------------------- moving

function applyMove(vertex, color, player) {
  state.colors[vertex] = color;
  state.owners[vertex] = player;
  state.selected = null;
  state.currentPlayer = player === "Alice" ? "Bob" : "Alice";
}

async function humanMove(vertex, color) {
  if (state.thinking || isOver() || state.currentPlayer !== state.humanRole) return;
  applyMove(vertex, color, state.humanRole);
  state.message = "";
  await refresh();
}

async function computerMove(token) {
  if (token !== state.moveToken || isOver() || !state.analysis) return;
  const pick = state.analysis.optimalMoves[0] || state.analysis.legalMoves[0];
  if (!pick) return;
  const label = nameOf(pick.vertex);
  applyMove(pick.vertex, pick.color, state.currentPlayer);
  state.message = `Computer coloured ${label} with colour ${pick.color}.`;
  await refresh();
}

const nameOf = (vertex) =>
  state.spec ? vertexLabel(state.spec.family, state.spec.n, vertex) : `v${vertex}`;

// ---------------------------------------------------------------- rendering

function render() {
  renderStatus();
  renderBoard();
  renderColorButtons();
  renderExplanation();
  renderOverlay();
  renderLevelList();
}

function renderStatus() {
  const spec = state.spec;
  if (!spec) return;

  renderVerdict();

  const done = spec.id && isDone(spec.id, state.humanRole);
  el.chip.textContent = done ? "Cleared" : `d = ${spec.d}, k = ${spec.k}`;
  el.chip.className = done ? "chip is-good" : "chip";

  if (state.thinking) {
    el.turn.textContent = "Thinking…";
    el.stateLine.textContent = "";
    return;
  }
  if (!state.analysis) {
    el.turn.textContent = "";
    el.stateLine.textContent = "";
    return;
  }
  if (isOver()) {
    el.turn.textContent = state.analysis.finished
      ? "Every vertex is coloured — Alice wins."
      : "A vertex has no colour left — Bob wins.";
  } else {
    const whose = state.currentPlayer === state.humanRole ? "Your turn" : "Computer's turn";
    el.turn.textContent = `${whose} (${state.currentPlayer}).`;
  }

  const left = state.colors.filter((c) => c === 0).length;
  el.stateLine.textContent =
    `${state.graph.order - left} of ${state.graph.order} coloured` +
    (state.graph.complete ? " · every pair blocks every other at this distance" : "");
}

// renderVerdict says up front whether this position is winnable from the
// player's side, and how forgiving it is. Stated before play, not after, so
// nobody spends the game wondering whether they are the problem.
function renderVerdict() {
  const winnable = winnableForPlayer();
  if (winnable === null) {
    el.verdict.textContent = "";
    el.verdict.className = "verdict-line";
    return;
  }

  if (!winnable) {
    el.verdict.textContent =
      `This one cannot be won as ${state.humanRole} — no matter how you play. ` +
      `Seeing why is the point; play it out, or switch sides.`;
    el.verdict.className = "verdict-line is-lost-cause";
    return;
  }

  const { winningOpenings, order } = state.verdict;
  let room = "";
  if (state.humanRole === "Alice") {
    if (winningOpenings === 1) {
      room = " Exactly one opening move works, out of " + order + " vertices — choose carefully.";
    } else if (winningOpenings < order) {
      room = ` Only ${winningOpenings} of ${order} vertices work as an opening.`;
    } else {
      room = " Any opening move works; the care is needed later.";
    }
  }
  el.verdict.textContent = `Winnable as ${state.humanRole}.${room}`;
  el.verdict.className = "verdict-line is-winnable";
}

function renderBoard() {
  el.board.innerHTML = "";
  if (!state.graph || !state.positions.length) return;

  const svg = (tag) => document.createElementNS("http://www.w3.org/2000/svg", tag);
  const dead = new Set(state.analysis ? state.analysis.dead : []);

  // Conflict-only pairs first, so real edges draw on top of them.
  const realEdge = new Set();
  state.graph.neighbors.forEach((list, v) => list.forEach((u) => realEdge.add(`${Math.min(u, v)}-${Math.max(u, v)}`)));

  state.graph.reach.forEach((list, v) => {
    list.forEach((u) => {
      if (u < v) return;
      const key = `${v}-${u}`;
      const line = svg("line");
      line.setAttribute("x1", state.positions[v].x);
      line.setAttribute("y1", state.positions[v].y);
      line.setAttribute("x2", state.positions[u].x);
      line.setAttribute("y2", state.positions[u].y);
      line.setAttribute("class", realEdge.has(key) ? "edge" : "edge conflict");
      el.board.appendChild(line);
    });
  });

  state.positions.forEach((point, vertex) => {
    const group = svg("g");
    group.setAttribute("class", "vertex");
    const color = state.colors[vertex];
    if (color) group.classList.add("is-colored");
    if (dead.has(vertex)) group.classList.add("is-dead");
    if (state.selected === vertex) group.classList.add("is-selected");

    const circle = svg("circle");
    circle.setAttribute("cx", point.x);
    circle.setAttribute("cy", point.y);
    circle.setAttribute("r", 19);
    circle.setAttribute("fill", color ? SWATCHES[(color - 1) % SWATCHES.length] : "#ffffff");
    group.appendChild(circle);

    const text = svg("text");
    text.setAttribute("x", point.x);
    text.setAttribute("y", point.y + 4);
    text.setAttribute("class", "vertex-label");
    text.textContent = nameOf(vertex);
    group.appendChild(text);

    group.addEventListener("click", () => selectVertex(vertex));
    el.board.appendChild(group);
  });
}

function selectVertex(vertex) {
  if (state.thinking || isOver()) return;
  if (state.currentPlayer !== state.humanRole) {
    state.message = "Wait for the computer to move.";
    render();
    return;
  }
  if (state.colors[vertex] !== 0) {
    state.message = `${nameOf(vertex)} is already coloured.`;
    render();
    return;
  }
  state.selected = state.selected === vertex ? null : vertex;
  state.message = "";
  render();
}

function renderColorButtons() {
  el.colors.innerHTML = "";
  if (!state.spec || isOver() || state.thinking) return;

  const allowed = new Set();
  if (state.selected !== null && state.analysis) {
    state.analysis.legalMoves
      .filter((m) => m.vertex === state.selected)
      .forEach((m) => allowed.add(m.color));
  }

  for (let color = 1; color <= state.spec.k; color += 1) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "color-button";
    button.style.background = SWATCHES[(color - 1) % SWATCHES.length];
    button.textContent = color;
    const usable = state.selected !== null && allowed.has(color);
    button.disabled = !usable;
    if (state.selected !== null && !usable) button.classList.add("is-blocked");
    button.addEventListener("click", () => humanMove(state.selected, color));
    el.colors.appendChild(button);
  }

  el.selection.textContent =
    state.selected === null
      ? "Pick a vertex, then pick a colour."
      : `${nameOf(state.selected)} selected — ${allowed.size} of ${state.spec.k} colours still available.`;
  el.message.textContent = state.message;
}

// renderExplanation is the part a static picture cannot do: when a vertex dies,
// say which neighbours took which colours, in the same terms the CLI uses.
function renderExplanation() {
  const analysis = state.analysis;
  if (!analysis || analysis.dead.length === 0) {
    el.explain.classList.add("hidden");
    el.explain.innerHTML = "";
    return;
  }

  const parts = analysis.dead.map((vertex) => {
    const blockers = (analysis.blockers && analysis.blockers[String(vertex)]) || [];
    const listed = blockers
      .map((b) => `<span class="swatch" style="background:${SWATCHES[(b.color - 1) % SWATCHES.length]}"></span>${nameOf(b.vertex)} = ${b.color}`)
      .join(", ");
    return `
      <p class="explain-head">${nameOf(vertex)} has no colour left.</p>
      <p class="explain-body">Blocked by ${listed}.</p>
      <p class="explain-foot">Between them that is all ${state.spec.k} colours, so nothing can go here.</p>`;
  });

  el.explain.innerHTML = parts.join("");
  el.explain.classList.remove("hidden");
}

function renderOverlay() {
  if (!isOver()) {
    el.overlay.classList.add("hidden");
    return;
  }
  const won = humanWon();
  const winnable = winnableForPlayer();

  el.overlayTitle.textContent = state.analysis.finished
    ? "Every vertex coloured"
    : `${nameOf(state.analysis.dead[0])} was trapped`;

  if (won) {
    el.overlayEyebrow.textContent = "You win";
    el.overlayBody.textContent = winnable
      ? "That is the winning side of this position."
      : "You won a position that should have been lost — the opponent slipped.";
  } else if (winnable === false) {
    // Losing was the only outcome. Say so plainly, or the player cannot tell
    // whether to try again.
    el.overlayEyebrow.textContent = "Lost — as it had to be";
    el.overlayBody.textContent =
      `This position is a loss for ${state.humanRole} however it is played, so there was ` +
      `nothing you could have done differently. Play again to look at it more closely, ` +
      `or switch sides to see the winning half.`;
  } else {
    el.overlayEyebrow.textContent = "You lose";
    el.overlayBody.textContent =
      "This one was winnable, so a move somewhere went wrong — usually earlier than it feels. " +
      "Try again.";
  }
  el.overlay.classList.remove("hidden");
}

function renderLevelList() {
  if (el.levelList.dataset.built === "yes") {
    el.levelList.querySelectorAll("[data-level]").forEach((node) => {
      const done = isDone(node.dataset.level, state.humanRole);
      node.classList.toggle("is-done", done);
      node.classList.toggle("is-current", state.spec && state.spec.id === node.dataset.level);
    });
    return;
  }
  el.levelList.innerHTML = "";
  LEVELS.forEach((level) => {
    const card = document.createElement("button");
    card.type = "button";
    card.className = "level-card";
    card.dataset.level = level.id;
    card.innerHTML = `<strong>${level.title}</strong><span class="level-meta">${level.difficulty} · d = ${level.d}, k = ${level.k}</span><span class="level-summary">${level.summary}</span>`;
    card.addEventListener("click", () => {
      closeDrawer();
      state.mode = "levels";
      loadSpec({ ...level });
    });
    el.levelList.appendChild(card);
  });
  el.levelList.dataset.built = "yes";
  renderLevelList();
}

// ---------------------------------------------------------------- sandbox

function sandboxSpec() {
  const family = el.sandboxFamily.value;
  const n = Number(el.sandboxSize.value);
  const d = Number(el.sandboxDistance.value);
  const k = Number(el.sandboxColors.value);
  const meta = FAMILIES.find((f) => f.id === family);
  return {
    family, n, d, k,
    title: `${meta.label} of ${n}, ${k} colours, d = ${d}`,
    summary: `${orderOf(family, n)} vertices. Built in your browser — nothing about this case was precomputed.`,
  };
}

function syncSandboxLabels() {
  const meta = FAMILIES.find((f) => f.id === el.sandboxFamily.value);
  el.sandboxSizeLabel.textContent = `Size (${meta.sizeLabel})`;
  el.sandboxSize.min = meta.min;
  el.sandboxSize.max = meta.max;
  if (Number(el.sandboxSize.value) < meta.min) el.sandboxSize.value = meta.min;
  if (Number(el.sandboxSize.value) > meta.max) el.sandboxSize.value = meta.max;
  el.sandboxInfo.textContent = `${orderOf(meta.id, Number(el.sandboxSize.value))} vertices`;
}

async function sandboxSolve() {
  const spec = sandboxSpec();
  el.sandboxInfo.textContent = "Solving…";
  try {
    const { chi } = await chromaticNumber(spec.family, spec.n, spec.d);
    el.sandboxInfo.textContent =
      `Needs ${chi} colour${chi === 1 ? "" : "s"} at d = ${spec.d}. ` +
      `With k = ${spec.k}, ${spec.k >= chi ? "Alice" : "Bob"} should win.`;
  } catch (error) {
    el.sandboxInfo.textContent = String(error.message || error);
  }
}

// ---------------------------------------------------------------- drawers

function openDrawer(id) {
  document.querySelectorAll(".drawer").forEach((d) => d.classList.add("hidden"));
  $(id).classList.remove("hidden");
  el.backdrop.classList.remove("hidden");
  state.openDrawer = id;
}

function closeDrawer() {
  document.querySelectorAll(".drawer").forEach((d) => d.classList.add("hidden"));
  el.backdrop.classList.add("hidden");
  state.openDrawer = null;
}

const toggleDrawer = (id) => (state.openDrawer === id ? closeDrawer() : openDrawer(id));

// ---------------------------------------------------------------- wiring

async function setRole(role) {
  state.humanRole = role;
  $("role-alice").classList.toggle("is-active", role === "Alice");
  $("role-bob").classList.toggle("is-active", role === "Bob");
  await restart();
}

async function restart() {
  if (!state.spec) return;
  state.colors = new Array(state.graph.order).fill(0);
  state.owners = new Array(state.graph.order).fill(null);
  state.currentPlayer = "Alice";
  state.selected = null;
  state.message = "";
  await refresh();
}

function nextLevel() {
  if (!state.spec || !state.spec.id) return;
  const index = LEVELS.findIndex((l) => l.id === state.spec.id);
  const next = LEVELS[(index + 1) % LEVELS.length];
  loadSpec({ ...next });
}

function showIntro() {
  el.intro.classList.remove("hidden");
}

function dismissIntro() {
  el.intro.classList.add("hidden");
  try {
    window.localStorage.setItem(INTRO_KEY, "yes");
  } catch {
    // Not being able to remember is fine; it just shows again next time.
  }
}

async function init() {
  loadProgress();

  let seenIntro = false;
  try {
    seenIntro = window.localStorage.getItem(INTRO_KEY) === "yes";
  } catch {
    seenIntro = false;
  }
  if (!seenIntro) showIntro();

  FAMILIES.forEach((family) => {
    const option = document.createElement("option");
    option.value = family.id;
    option.textContent = family.label;
    el.sandboxFamily.appendChild(option);
  });
  el.sandboxFamily.value = "sunlet";
  el.sandboxSize.value = 4;
  el.sandboxDistance.value = 1;
  el.sandboxColors.value = 3;
  syncSandboxLabels();

  $("open-levels").addEventListener("click", () => toggleDrawer("levels-drawer"));
  $("open-sandbox").addEventListener("click", () => toggleDrawer("sandbox-drawer"));
  $("open-game-panel").addEventListener("click", () => toggleDrawer("game-drawer"));
  $("open-how-to").addEventListener("click", () => toggleDrawer("how-to-drawer"));
  el.backdrop.addEventListener("click", closeDrawer);
  document.querySelectorAll(".drawer-close").forEach((b) => b.addEventListener("click", closeDrawer));

  $("role-alice").addEventListener("click", () => setRole("Alice"));
  $("role-bob").addEventListener("click", () => setRole("Bob"));
  $("restart-button").addEventListener("click", restart);
  $("overlay-restart").addEventListener("click", restart);
  $("overlay-next").addEventListener("click", nextLevel);
  $("overlay-switch-role").addEventListener("click", () =>
    setRole(state.humanRole === "Alice" ? "Bob" : "Alice"));

  el.sandboxFamily.addEventListener("change", syncSandboxLabels);
  el.sandboxSize.addEventListener("input", syncSandboxLabels);
  $("sandbox-solve").addEventListener("click", sandboxSolve);
  $("sandbox-play").addEventListener("click", () => {
    closeDrawer();
    state.mode = "sandbox";
    loadSpec(sandboxSpec());
  });

  $("intro-start").addEventListener("click", dismissIntro);
  $("show-intro").addEventListener("click", () => { closeDrawer(); showIntro(); });

  $("role-alice").classList.add("is-active");
  await loadSpec({ ...LEVELS[0] });
}

init().catch((error) => {
  el.title.textContent = "Failed to start";
  el.summary.textContent = String(error && error.message ? error.message : error);
});
