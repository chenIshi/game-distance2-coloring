// Runs the Go solver off the main thread.
//
// WebAssembly will happily let the solver think for many seconds on a hard
// position, and on the main thread that freezes the page: no scrolling, no
// clicking, nothing. A worker keeps the interface alive while it works.
//
// The protocol is one request in, one reply out, with the caller's id echoed
// back so several queries can be in flight without getting confused.

let ready = null;

function start() {
  if (ready) return ready;
  ready = (async () => {
    importScripts("./wasm_exec.js");
    const go = new Go();
    const source = await WebAssembly.instantiateStreaming(fetch("./solver.wasm"), go.importObject)
      .catch(async () => {
        // instantiateStreaming needs the right Content-Type; some static
        // servers do not send it for .wasm. Fall back to fetching the bytes.
        const bytes = await (await fetch("./solver.wasm")).arrayBuffer();
        return WebAssembly.instantiate(bytes, go.importObject);
      });
    go.run(source.instance);
    if (typeof gameColoringSolve !== "function") {
      throw new Error("solver did not register itself");
    }
  })();
  return ready;
}

self.onmessage = async (event) => {
  const { id, request } = event.data;
  try {
    await start();
    self.postMessage({ id, result: gameColoringSolve(request) });
  } catch (error) {
    self.postMessage({ id, result: { error: String(error && error.message ? error.message : error) } });
  }
};
