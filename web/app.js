const COLOR_SWATCHES = ["#4c956c", "#d68c45", "#8d5a97", "#2d728f", "#c75146"];
const STORAGE_KEY = "distance2-game-progress";

const state = {
  data: null,
  caseMap: new Map(),
  caseOrder: [],
  currentCaseId: null,
  humanRole: "Alice",
  currentPlayer: "Alice",
  colors: [],
  owners: [],
  selectedVertex: null,
  message: "",
  progress: {},
  openDrawerId: null,
};

const roleAliceButton = document.querySelector("#role-alice");
const roleBobButton = document.querySelector("#role-bob");
const restartButton = document.querySelector("#restart-button");
const openLevelsButton = document.querySelector("#open-levels");
const openGamePanelButton = document.querySelector("#open-game-panel");
const openNextMoveButton = document.querySelector("#open-next-move");
const openHowToButton = document.querySelector("#open-how-to");
const levelList = document.querySelector("#level-list");
const titleLabel = document.querySelector("#case-title");
const summaryLabel = document.querySelector("#case-summary");
const progressChip = document.querySelector("#progress-chip");
const turnLabel = document.querySelector("#turn-label");
const stateLabel = document.querySelector("#state-label");
const selectionLabel = document.querySelector("#selection-label");
const messageLabel = document.querySelector("#message-label");
const graphBoard = document.querySelector("#graph-board");
const colorButtons = document.querySelector("#color-buttons");
const resultOverlay = document.querySelector("#result-overlay");
const overlayEyebrow = document.querySelector("#overlay-eyebrow");
const overlayTitle = document.querySelector("#overlay-title");
const overlayBody = document.querySelector("#overlay-body");
const overlayRestart = document.querySelector("#overlay-restart");
const overlayNext = document.querySelector("#overlay-next");
const overlaySwitchRole = document.querySelector("#overlay-switch-role");
const drawerBackdrop = document.querySelector("#drawer-backdrop");
const drawers = Array.from(document.querySelectorAll(".drawer"));
const drawerCloseButtons = Array.from(document.querySelectorAll(".drawer-close"));

function encodeState(colors, currentPlayer) {
  return `${colors.join(".")}|${currentPlayer[0]}`;
}

function getCurrentCase() {
  return state.caseMap.get(state.currentCaseId);
}

function getCurrentStateData() {
  const currentCase = getCurrentCase();
  return currentCase.states[encodeState(state.colors, state.currentPlayer)];
}

function isGameOver() {
  const currentState = getCurrentStateData();
  return currentState.finished || currentState.deadVertices.length > 0;
}

function didHumanWin() {
  const currentState = getCurrentStateData();
  if (currentState.finished) {
    return state.humanRole === "Alice";
  }
  if (currentState.deadVertices.length > 0) {
    return state.humanRole === "Bob";
  }
  return false;
}

function loadProgress() {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    state.progress = raw ? JSON.parse(raw) : {};
  } catch {
    state.progress = {};
  }
}

function saveProgress() {
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(state.progress));
  } catch {
    state.progress = state.progress;
  }
}

function progressKey(caseId, role) {
  return `${role}:${caseId}`;
}

function isLevelDone(caseId, role) {
  return Boolean(state.progress[progressKey(caseId, role)]);
}

function markCurrentLevelDone() {
  state.progress[progressKey(state.currentCaseId, state.humanRole)] = true;
  saveProgress();
}

function maybeRecordCompletion() {
  if (!isGameOver()) {
    return;
  }

  const currentCase = getCurrentCase();
  const levelIsSolvableForHuman = currentCase.initialWinner === state.humanRole;
  if (levelIsSolvableForHuman) {
    if (didHumanWin()) {
      markCurrentLevelDone();
    }
    return;
  }
  markCurrentLevelDone();
}

function setMessage(message) {
  state.message = message;
  messageLabel.textContent = message;
}

function openDrawer(drawerId) {
  state.openDrawerId = drawerId;
  for (const drawer of drawers) {
    drawer.classList.toggle("hidden", drawer.id !== drawerId);
  }
  drawerBackdrop.classList.toggle("hidden", !drawerId);
}

function closeDrawer() {
  state.openDrawerId = null;
  for (const drawer of drawers) {
    drawer.classList.add("hidden");
  }
  drawerBackdrop.classList.add("hidden");
}

function toggleDrawer(drawerId) {
  if (state.openDrawerId === drawerId) {
    closeDrawer();
    return;
  }
  openDrawer(drawerId);
}

function getNextCaseId() {
  const currentIndex = state.caseOrder.indexOf(state.currentCaseId);
  if (currentIndex === -1) {
    return state.caseOrder[0];
  }
  return state.caseOrder[(currentIndex + 1) % state.caseOrder.length];
}

function resetGame() {
  const currentCase = getCurrentCase();
  state.colors = new Array(currentCase.size).fill(0);
  state.owners = new Array(currentCase.size).fill(null);
  state.currentPlayer = "Alice";
  state.selectedVertex = null;
  setMessage("Select a vertex to begin this round.");
  render();

  if (state.humanRole !== state.currentPlayer) {
    window.setTimeout(playComputerTurn, 380);
  }
}

function setRole(role) {
  state.humanRole = role;
  document.body.classList.toggle("theme-bob", role === "Bob");
  roleAliceButton.classList.toggle("active", role === "Alice");
  roleBobButton.classList.toggle("active", role === "Bob");
  resetGame();
}

function chooseCase(caseId) {
  state.currentCaseId = caseId;
  closeDrawer();
  resetGame();
}

function renderLevelCards() {
  levelList.innerHTML = "";

  for (const caseId of state.caseOrder) {
    const item = state.caseMap.get(caseId);
    const card = document.createElement("button");
    card.type = "button";
    card.className = "level-card";
    if (item.id === state.currentCaseId) {
      card.classList.add("active");
    }
    if (isLevelDone(item.id, state.humanRole)) {
      card.classList.add("done");
    }

    const topline = document.createElement("div");
    topline.className = "level-topline";

    const title = document.createElement("strong");
    title.textContent = item.title;

    const tag = document.createElement("span");
    tag.className = "level-tag";
    tag.textContent = isLevelDone(item.id, state.humanRole)
      ? "Done"
      : item.challenge
        ? "Challenge"
        : "Core";

    const teaser = document.createElement("p");
    teaser.textContent = item.summary;
    teaser.className = "drawer-copy";

    topline.appendChild(title);
    topline.appendChild(tag);
    card.appendChild(topline);
    card.appendChild(teaser);
    card.addEventListener("click", () => chooseCase(item.id));
    levelList.appendChild(card);
  }
}

function renderCaseControls() {
  const currentCase = getCurrentCase();
  titleLabel.textContent = currentCase.title;
  summaryLabel.textContent = currentCase.summary;
  progressChip.className = "chip";
  progressChip.textContent = currentCase.challenge ? "Challenge level" : "Core level";
}

function renderStatus() {
  const currentState = getCurrentStateData();
  progressChip.classList.remove("is-safe", "is-danger", "is-finished", "is-win");

  turnLabel.textContent = `Current turn: ${state.currentPlayer}`;

  if (currentState.finished) {
    if (state.humanRole === "Alice") {
      progressChip.classList.add("is-win");
      progressChip.textContent = "Level cleared";
    } else {
      progressChip.classList.add("is-finished");
      progressChip.textContent = "Alice finished it";
    }
    stateLabel.textContent =
      state.humanRole === "Alice"
        ? "Every vertex is colored. You cleared the level."
        : "Every vertex is colored, so Alice completed the puzzle.";
    return;
  }

  if (currentState.deadVertices.length > 0) {
    progressChip.classList.add("is-danger");
    progressChip.textContent = state.humanRole === "Bob" ? "Trap complete" : "You got trapped";
    stateLabel.textContent =
      state.humanRole === "Bob"
        ? "A dead vertex appeared, so Bob wins this level."
        : "A dead vertex appeared, so Bob wins this level. Replay and try another line.";
    return;
  }

  if (state.humanRole === "Alice") {
    if (currentState.aliceWins) {
      progressChip.classList.add("is-safe");
      progressChip.textContent = "Still alive";
      stateLabel.textContent = "Alice can still force a full coloring from here.";
    } else {
      progressChip.classList.add("is-danger");
      progressChip.textContent = "Danger";
      stateLabel.textContent = "Bob can force a trap from this position.";
    }
    return;
  }

  if (currentState.aliceWins) {
    progressChip.classList.add("is-danger");
    progressChip.textContent = "Alice can recover";
    stateLabel.textContent = "Alice still has a route to finish the graph.";
  } else {
    progressChip.classList.add("is-safe");
    progressChip.textContent = "Trap pressure";
    stateLabel.textContent = "Bob still has enough pressure to force a dead vertex.";
  }
}

function renderColorButtons() {
  const currentCase = getCurrentCase();
  const currentState = getCurrentStateData();
  colorButtons.innerHTML = "";

  const legalByVertex = new Map();
  for (const move of currentState.legalMoves) {
    if (!legalByVertex.has(move.vertex)) {
      legalByVertex.set(move.vertex, new Set());
    }
    legalByVertex.get(move.vertex).add(move.color);
  }

  const selectedLegalColors =
    state.selectedVertex === null ? new Set() : legalByVertex.get(state.selectedVertex) ?? new Set();

  for (let color = 1; color <= currentCase.k; color += 1) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "color-button";
    button.textContent = `Color ${color}`;
    button.style.backgroundColor = COLOR_SWATCHES[(color - 1) % COLOR_SWATCHES.length];
    button.style.color = "#fff";
    button.disabled =
      state.humanRole !== state.currentPlayer ||
      state.selectedVertex === null ||
      isGameOver() ||
      !selectedLegalColors.has(color);
    button.classList.toggle("active", selectedLegalColors.has(color));
    button.addEventListener("click", () => makeHumanMove(state.selectedVertex, color));
    colorButtons.appendChild(button);
  }
}

function renderBoard() {
  const currentCase = getCurrentCase();
  const currentState = getCurrentStateData();
  const deadSet = new Set(currentState.deadVertices);
  const highlightSet =
    state.selectedVertex === null ? new Set() : new Set(currentCase.squareGraph[state.selectedVertex]);

  graphBoard.innerHTML = "";

  currentCase.graph.forEach((neighbors, vertex) => {
    neighbors
      .filter((neighbor) => neighbor > vertex)
      .forEach((neighbor) => {
        const line = document.createElementNS("http://www.w3.org/2000/svg", "line");
        line.setAttribute("class", "edge");
        line.setAttribute("x1", currentCase.positions[vertex].x);
        line.setAttribute("y1", currentCase.positions[vertex].y);
        line.setAttribute("x2", currentCase.positions[neighbor].x);
        line.setAttribute("y2", currentCase.positions[neighbor].y);
        graphBoard.appendChild(line);
      });
  });

  currentCase.positions.forEach((position, vertex) => {
    const group = document.createElementNS("http://www.w3.org/2000/svg", "g");
    group.setAttribute("class", "vertex");
    if (vertex === state.selectedVertex) {
      group.classList.add("selected");
    }
    if (deadSet.has(vertex)) {
      group.classList.add("dead");
    }
    if (highlightSet.has(vertex)) {
      group.classList.add("distance-two");
    }

    const outer = document.createElementNS("http://www.w3.org/2000/svg", "circle");
    outer.setAttribute("class", "outer");
    outer.setAttribute("cx", position.x);
    outer.setAttribute("cy", position.y);
    outer.setAttribute("r", 26);
    outer.setAttribute("fill", "#fff");
    outer.setAttribute("stroke", "#7d6a58");
    outer.setAttribute("stroke-width", "3");

    const inner = document.createElementNS("http://www.w3.org/2000/svg", "circle");
    inner.setAttribute("cx", position.x);
    inner.setAttribute("cy", position.y);
    inner.setAttribute("r", 21);
    const color = state.colors[vertex];
    inner.setAttribute("fill", color === 0 ? "rgba(235, 228, 216, 0.95)" : COLOR_SWATCHES[(color - 1) % COLOR_SWATCHES.length]);

    const vertexLabel = document.createElementNS("http://www.w3.org/2000/svg", "text");
    vertexLabel.setAttribute("class", "vertex-label");
    vertexLabel.setAttribute("x", position.x);
    vertexLabel.setAttribute("y", position.y + 44);
    vertexLabel.textContent = `Vertex ${vertex + 1}`;

    group.appendChild(outer);
    group.appendChild(inner);

    if (color !== 0) {
      const colorLabel = document.createElementNS("http://www.w3.org/2000/svg", "text");
      colorLabel.setAttribute("class", "color-label");
      colorLabel.setAttribute("x", position.x);
      colorLabel.setAttribute("y", position.y + 7);
      colorLabel.textContent = String(color);
      group.appendChild(colorLabel);

      const ownerBadge = document.createElementNS("http://www.w3.org/2000/svg", "circle");
      ownerBadge.setAttribute("cx", position.x + 18);
      ownerBadge.setAttribute("cy", position.y - 18);
      ownerBadge.setAttribute("r", 10);
      ownerBadge.setAttribute("fill", state.owners[vertex] === "Alice" ? "#2d6a4f" : "#aa2e25");
      group.appendChild(ownerBadge);

      const ownerLabel = document.createElementNS("http://www.w3.org/2000/svg", "text");
      ownerLabel.setAttribute("class", "owner-label");
      ownerLabel.setAttribute("x", position.x + 18);
      ownerLabel.setAttribute("y", position.y - 14);
      ownerLabel.textContent = state.owners[vertex] === "Alice" ? "A" : "B";
      group.appendChild(ownerLabel);
    }

    group.appendChild(vertexLabel);
    group.addEventListener("click", () => selectVertex(vertex));
    graphBoard.appendChild(group);
  });
}

function renderSelectionHint() {
  const currentState = getCurrentStateData();
  const suggestedMove = currentState.optimalMoves[0] ?? currentState.legalMoves[0];

  if (isGameOver()) {
    selectionLabel.textContent = "Round over. Replay this level or move to the next one.";
    return;
  }

  if (state.selectedVertex === null) {
    if (state.currentPlayer === state.humanRole && suggestedMove) {
      selectionLabel.textContent = `Suggested move: choose Vertex ${suggestedMove.vertex + 1}, then use color ${suggestedMove.color}.`;
    } else {
      selectionLabel.textContent = "Wait for the computer move, or inspect the board.";
    }
    return;
  }

  const moves = currentState.legalMoves.filter((move) => move.vertex === state.selectedVertex);
  if (moves.length === 0) {
    selectionLabel.textContent = `Vertex ${state.selectedVertex + 1} is dead. No legal color remains there.`;
    return;
  }

  selectionLabel.textContent = `Vertex ${state.selectedVertex + 1} can use color${moves.length > 1 ? "s" : ""} ${moves.map((move) => move.color).join(", ")}.`;
}

function renderOverlay() {
  const currentState = getCurrentStateData();
  if (!currentState.finished && currentState.deadVertices.length === 0) {
    resultOverlay.classList.add("hidden");
    return;
  }

  resultOverlay.classList.remove("hidden");

  if (currentState.finished) {
    overlayEyebrow.textContent = "Level Cleared";
    overlayTitle.textContent =
      state.humanRole === "Alice" ? "You finished the graph." : "Alice completed the graph.";
    overlayBody.textContent =
      state.humanRole === "Alice"
        ? "Every vertex stayed alive long enough to be colored. Replay to try a cleaner route, or move on."
        : "Bob could not create a trap in time. Try another level or switch sides.";
  } else {
    overlayEyebrow.textContent = "Trap Triggered";
    overlayTitle.textContent =
      state.humanRole === "Bob" ? "You trapped a vertex." : "Bob trapped the graph.";
    overlayBody.textContent =
      state.humanRole === "Bob"
        ? "One uncolored vertex ran out of legal colors. Replay to sharpen the trap, or move on."
        : "At least one uncolored vertex had no legal color left. Replay and try a different order.";
  }
}

function render() {
  maybeRecordCompletion();
  renderLevelCards();
  renderCaseControls();
  renderStatus();
  renderSelectionHint();
  renderColorButtons();
  renderBoard();
  renderOverlay();
}

function selectVertex(vertex) {
  const currentState = getCurrentStateData();
  if (state.currentPlayer !== state.humanRole || isGameOver()) {
    return;
  }

  state.selectedVertex = vertex;
  const availableMoves = currentState.legalMoves.filter((move) => move.vertex === vertex);
  if (availableMoves.length === 0) {
    setMessage(`Vertex ${vertex + 1} cannot be colored right now.`);
  } else {
    const suggestedMove = currentState.optimalMoves.find((move) => move.vertex === vertex);
    if (suggestedMove) {
      setMessage(`Good choice. Try color ${suggestedMove.color} on Vertex ${vertex + 1}.`);
    } else {
      setMessage(`Vertex ${vertex + 1} can use color${availableMoves.length > 1 ? "s" : ""} ${availableMoves.map((move) => move.color).join(", ")}.`);
    }
  }
  render();
}

function advanceTurn(vertex, color) {
  state.colors[vertex] = color;
  state.owners[vertex] = state.currentPlayer;
  state.currentPlayer = state.currentPlayer === "Alice" ? "Bob" : "Alice";
  state.selectedVertex = null;
}

function makeHumanMove(vertex, color) {
  const currentState = getCurrentStateData();
  const move = currentState.legalMoves.find((item) => item.vertex === vertex && item.color === color);

  if (!move) {
    setMessage(`That move is not allowed on Vertex ${vertex + 1}.`);
    render();
    return;
  }

  advanceTurn(vertex, color);
  setMessage(`You used color ${color} on Vertex ${vertex + 1}.`);
  render();

  const nextState = getCurrentStateData();
  if (!nextState.finished && nextState.deadVertices.length === 0 && state.currentPlayer !== state.humanRole) {
    window.setTimeout(playComputerTurn, 420);
  }
}

function playComputerTurn() {
  const currentState = getCurrentStateData();
  if (currentState.finished || currentState.deadVertices.length > 0 || state.currentPlayer === state.humanRole) {
    render();
    return;
  }

  const move = currentState.optimalMoves[0] ?? currentState.legalMoves[0];
  advanceTurn(move.vertex, move.color);
  setMessage(`Computer used color ${move.color} on Vertex ${move.vertex + 1}.`);
  render();
}

function playNextLevel() {
  chooseCase(getNextCaseId());
}

function switchRole() {
  setRole(state.humanRole === "Alice" ? "Bob" : "Alice");
}

async function init() {
  const response = await fetch("./data/cases.json");
  state.data = await response.json();
  state.caseMap = new Map(state.data.cases.map((item) => [item.id, item]));
  state.caseOrder = state.data.cases.map((item) => item.id);
  loadProgress();
  state.currentCaseId = state.caseOrder[0];

  roleAliceButton.addEventListener("click", () => setRole("Alice"));
  roleBobButton.addEventListener("click", () => setRole("Bob"));
  restartButton.addEventListener("click", resetGame);
  overlayRestart.addEventListener("click", resetGame);
  overlayNext.addEventListener("click", playNextLevel);
  overlaySwitchRole.addEventListener("click", switchRole);
  openLevelsButton.addEventListener("click", () => toggleDrawer("levels-drawer"));
  openGamePanelButton.addEventListener("click", () => toggleDrawer("game-drawer"));
  openNextMoveButton.addEventListener("click", () => toggleDrawer("next-move-drawer"));
  openHowToButton.addEventListener("click", () => toggleDrawer("how-to-drawer"));
  drawerBackdrop.addEventListener("click", closeDrawer);
  for (const button of drawerCloseButtons) {
    button.addEventListener("click", () => closeDrawer());
  }

  setRole("Alice");
}

init().catch((error) => {
  titleLabel.textContent = "Failed to load";
  setMessage(String(error));
});
