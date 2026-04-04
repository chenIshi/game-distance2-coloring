const COLOR_SWATCHES = ["#4c956c", "#d68c45", "#8d5a97", "#2d728f", "#c75146"];

const state = {
  data: null,
  caseMap: new Map(),
  currentCaseId: null,
  humanRole: "Alice",
  currentPlayer: "Alice",
  colors: [],
  owners: [],
  selectedVertex: null,
  message: "",
};

const caseSelect = document.querySelector("#case-select");
const roleAliceButton = document.querySelector("#role-alice");
const roleBobButton = document.querySelector("#role-bob");
const restartButton = document.querySelector("#restart-button");
const titleLabel = document.querySelector("#case-title");
const summaryLabel = document.querySelector("#case-summary");
const progressChip = document.querySelector("#progress-chip");
const turnLabel = document.querySelector("#turn-label");
const stateLabel = document.querySelector("#state-label");
const selectionLabel = document.querySelector("#selection-label");
const messageLabel = document.querySelector("#message-label");
const graphBoard = document.querySelector("#graph-board");
const colorButtons = document.querySelector("#color-buttons");
const statusBox = turnLabel.closest(".status-box");

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

function setMessage(message) {
  state.message = message;
  messageLabel.textContent = message;
}

function resetGame() {
  const currentCase = getCurrentCase();
  state.colors = new Array(currentCase.size).fill(0);
  state.owners = new Array(currentCase.size).fill(null);
  state.currentPlayer = "Alice";
  state.selectedVertex = null;
  setMessage("Select a vertex to begin.");
  render();

  if (state.humanRole !== state.currentPlayer) {
    window.setTimeout(playComputerTurn, 350);
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
  resetGame();
}

function renderCaseControls() {
  const currentCase = getCurrentCase();
  titleLabel.textContent = currentCase.title;
  summaryLabel.textContent = currentCase.summary;
  progressChip.textContent = currentCase.challenge ? "Optional challenge" : "Main lesson";
  progressChip.className = "chip";
  caseSelect.value = state.currentCaseId;
}

function renderStatus() {
  const currentState = getCurrentStateData();
  statusBox.classList.remove("state-safe", "state-danger", "state-finished", "state-win");
  progressChip.classList.remove("is-safe", "is-danger", "is-finished", "is-win");

  turnLabel.textContent = `Current turn: ${state.currentPlayer}`;
  if (currentState.finished) {
    if (state.humanRole === "Alice") {
      statusBox.classList.add("state-win");
      progressChip.classList.add("is-win");
      progressChip.textContent = "You won";
    } else {
      statusBox.classList.add("state-finished");
      progressChip.classList.add("is-finished");
      progressChip.textContent = "Alice won";
    }
    stateLabel.textContent =
      state.humanRole === "Alice"
        ? "All dots are colored. You finished the game successfully."
        : "All dots are colored, so Alice completed the game.";
    return;
  }
  if (currentState.deadVertices.length > 0) {
    statusBox.classList.add("state-danger");
    progressChip.classList.add("is-danger");
    progressChip.textContent = state.humanRole === "Bob" ? "You won" : "You got stuck";
    stateLabel.textContent =
      state.humanRole === "Bob"
        ? "A dead dot appeared, so Bob won this round."
        : "A dead vertex appeared, so Bob won this round. Restart and try a different choice.";
    return;
  }

  if (state.humanRole === "Alice") {
    if (currentState.aliceWins) {
      statusBox.classList.add("state-safe");
      progressChip.classList.add("is-safe");
      progressChip.textContent = "You are still on track";
      stateLabel.textContent = "Alice can still reach a full coloring from here.";
    } else {
      statusBox.classList.add("state-danger");
      progressChip.classList.add("is-danger");
      progressChip.textContent = "Warning";
      stateLabel.textContent = "This position is slipping away. Bob can force a win unless you go back and change the line.";
    }
    return;
  }

  if (currentState.aliceWins) {
    statusBox.classList.add("state-danger");
    progressChip.classList.add("is-danger");
    progressChip.textContent = "Warning";
    stateLabel.textContent = "Alice can still recover from this position.";
  } else {
    statusBox.classList.add("state-safe");
    progressChip.classList.add("is-safe");
    progressChip.textContent = "You are pressuring Alice";
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

  const selectedLegalColors = state.selectedVertex === null ? new Set() : legalByVertex.get(state.selectedVertex) ?? new Set();

  for (let color = 1; color <= currentCase.k; color += 1) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "color-button";
    button.textContent = `Color ${color}`;
    button.style.backgroundColor = COLOR_SWATCHES[(color - 1) % COLOR_SWATCHES.length];
    button.style.color = "#fff";
    button.disabled = state.humanRole !== state.currentPlayer || state.selectedVertex === null || !selectedLegalColors.has(color);
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
    inner.setAttribute("fill", color === 0 ? "#f3ede5" : COLOR_SWATCHES[(color - 1) % COLOR_SWATCHES.length]);

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
  if (state.selectedVertex === null) {
    if (state.currentPlayer === state.humanRole && suggestedMove) {
      selectionLabel.textContent = `Suggested move: pick Vertex ${suggestedMove.vertex + 1}, then choose color ${suggestedMove.color}.`;
    } else {
      selectionLabel.textContent = "Select a vertex to see which colors are allowed.";
    }
    return;
  }

  const moves = currentState.legalMoves.filter((move) => move.vertex === state.selectedVertex);
  if (moves.length === 0) {
    selectionLabel.textContent = `Vertex ${state.selectedVertex + 1} is dead. No color can be used there now.`;
    return;
  }

  selectionLabel.textContent = `Vertex ${state.selectedVertex + 1} can use color${moves.length > 1 ? "s" : ""} ${moves.map((move) => move.color).join(", ")}.`;
}

function render() {
  renderCaseControls();
  renderStatus();
  renderSelectionHint();
  renderColorButtons();
  renderBoard();
}

function selectVertex(vertex) {
  const currentState = getCurrentStateData();
  if (state.currentPlayer !== state.humanRole) {
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
    window.setTimeout(playComputerTurn, 450);
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

async function init() {
  const response = await fetch("./data/cases.json");
  state.data = await response.json();
  state.caseMap = new Map(state.data.cases.map((item) => [item.id, item]));

  for (const item of state.data.cases) {
    const option = document.createElement("option");
    option.value = item.id;
    option.textContent = item.challenge ? `${item.title} (challenge)` : item.title;
    caseSelect.appendChild(option);
  }

  state.currentCaseId = state.data.cases[0].id;
  roleAliceButton.addEventListener("click", () => setRole("Alice"));
  roleBobButton.addEventListener("click", () => setRole("Bob"));
  restartButton.addEventListener("click", resetGame);
  caseSelect.addEventListener("change", (event) => chooseCase(event.target.value));

  setRole("Alice");
}

init().catch((error) => {
  titleLabel.textContent = "Failed to load";
  setMessage(String(error));
});
