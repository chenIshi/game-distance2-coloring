// Thin promise wrapper around the solver worker.
//
// Every call is asynchronous, which is the one real cost of moving the solver
// off the main thread: the page can no longer ask "who is winning?" in the
// middle of rendering. It has to ask, wait, then render.

const worker = new Worker("./solver-worker.js");
const pending = new Map();
let nextId = 1;

worker.onmessage = (event) => {
  const { id, result } = event.data;
  const waiting = pending.get(id);
  if (!waiting) return;
  pending.delete(id);
  if (result && result.error) waiting.reject(new Error(result.error));
  else waiting.resolve(result);
};

worker.onerror = (event) => {
  for (const [, waiting] of pending) waiting.reject(new Error(event.message || "solver failed"));
  pending.clear();
};

function ask(request) {
  return new Promise((resolve, reject) => {
    const id = nextId++;
    pending.set(id, { resolve, reject });
    worker.postMessage({ id, request });
  });
}

export function describeGraph(family, n, d) {
  return ask({ op: "graph", family, n, d });
}

export function chromaticNumber(family, n, d) {
  return ask({ op: "chi", family, n, d });
}

export function analyzeState(family, n, d, k, colors, aliceTurn) {
  return ask({ op: "state", family, n, d, k, colors, aliceTurn });
}
