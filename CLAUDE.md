# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Conventions

Fixed vocabulary for this project; do not re-derive or vary these.

**The game.** Alice and Bob alternate, Alice moves first. A move colors one uncolored
vertex with a color not used by any vertex within distance `d`. Alice (the maker) wins if
every vertex gets colored; Bob (the breaker) wins the moment some uncolored vertex has no
legal color left — that vertex is called a *dead vertex*. `chi_g,d(G)` is the least number
of colors for which Alice wins. `d` defaults to 2.

**Families, sizes, and vertex numbering.** The numbering is part of the contract: graph
symmetries are expressed as permutations of these indices, so changing the numbering
silently invalidates them.

| Family | Notation | Vertices | Numbering |
|---|---|---|---|
| Path | `P_n` | n | `0..n-1` along the path |
| Cycle | `C_n` | n | `0..n-1` around the cycle |
| Wheel | `W_n` | n+1 | hub `0`, rim `1..n` |
| Helm | `H_n` | 2n+1 | hub `0`, rim `1..n`, pendant of rim `i` at `n+i` |
| Sunlet | `S_n` | 2n | rim `0..n-1`, pendant of rim `i` at `n+i` |

`W_n` has *n rim vertices*, not n total. Sunlet is also written `C_n` ⊙ `K_1`; some papers
call it a crown graph, but "crown graph" more usually means K_{n,n} minus a perfect
matching — prefer "sunlet" and state the vertex count.

**Degenerate cases worth short-circuiting.** If `d >= diam(G)` then `G^d` is complete and
`chi_g,d(G) = |V|` with no search needed. This catches whole families: `W_n` and `K_{1,n}`
both have diameter 2, so they are trivial at every `d >= 2` and only carry information at
`d = 1`. `C_3`, `C_4`, `C_5` are likewise complete under `d = 2`.

**Monotonicity in k and d: measured, and holding.** Both were open questions when the
distance-d generalization landed, and both are now checked over 90+ (graph, distance)
combinations across all six families with zero violations:

- *In k*: Alice never loses a game she could win with fewer colours. The win vector is
  always a clean block of Bob wins followed by a block of Alice wins.
- *In d*: `chi_g,d` never decreases as `d` grows. This one is less obvious than it looks —
  `G^d` is a subgraph of `G^(d+1)`, and the game chromatic number is famously **not**
  monotone under subgraphs, so a subgraph can need strictly more colours than the graph
  containing it. It had to be measured.

This is evidence, not proof. Re-run it with `python3 scripts/verify_properties.py`
(exits non-zero on violation, so it is CI-able); `tests/test_monotonicity.py` guards a
smaller grid continuously.

**The k-sweep still scans upward, and should stay that way.** Monotonicity would make a
binary search *legal*, but measurement says it would be a *pessimization*: solve cost grows
steeply with k, not with closeness to the threshold. On `C_9` at d=2, k=9 costs 44% of the
total while the threshold k=5 costs 2%. Scanning upward stops at the threshold and never
pays for the expensive large-k solves; a binary search over `[0, |V|]` would probe above it.
The linear scan is the right algorithm here, not a placeholder.

Measured results live in `FINDINGS.md` — keep new findings there rather than growing
this file, which is for working conventions.

**Monotonicity in n definitely fails, in every family.** A larger graph can need strictly
fewer colours. This is a real property of the problem, not a bug, and it is the single
most common thing to mistake for one:

| Comparison | Smaller | Larger |
|---|---|---|
| `C_6` vs `C_7` at d=2 | 5 | **4** |
| `W_5` vs `W_6` at d=1 | 4 | **3** |
| `H_3` vs `H_4` at d=2 | 6 | **5** |
| `S_4` vs `S_5` at d=1 | 4 | **3** |
| `P_6` vs `P_7` at d=3 | 5 | **4** |

`NonMonotonicityInSizeTests` pins one case per family. Treat a dip in a sweep as data, not
as a defect — and never "smooth" or interpolate a results table on the assumption that the
values climb with n. Parity of the structure matters more than its size.

**Symmetries.** A symmetry is a permutation `p` of vertex indices where every edge stays an
edge. Under-reporting symmetries costs speed only; over-reporting them corrupts answers —
so builders should declare the obvious rotations/reflections and a validity test should
confirm them, rather than anyone trying to prove completeness. Symmetries of `G` are
automatically symmetries of `G^d`, so they are computed once per graph and reused for all
`d`.

## Commands

```bash
# Solve from the repo root (no install needed; run_cli.py injects src/ into sys.path)
python run_cli.py path 6
python run_cli.py cycle 5 --k 4
python run_cli.py star 4 --k 3 --explain

# Everyday test suite (~5s)
python3 -m unittest discover -s tests

# Including the expensive grids (~2min): 11-vertex helms, distance-3 sweeps
GAME_COLORING_SLOW=1 python3 -m unittest discover -s tests

# A single test case
python3 -m unittest tests.test_paths.PathGraphTests.test_small_known_values

# Go (requires go on PATH; this machine has it at ~/.local/go/bin)
go build ./... && go vet ./... && gofmt -l .
go test -short ./...          # ~2s
go test ./...                 # ~16s, includes the expensive grid
go run ./cmd/gamecolor -family helm -n 4 -d 3
go run ./cmd/gamecolor -family cycle -n 7 -k 4
go run ./cmd/gamecolor -family helm -n 4 -k 3 -d 1 -explain

# The research artifact: every family, size, and distance, parallel across cells
go run ./cmd/gamecolor -sweep -max-order 12 -distances 1,2,3 -format table
go run ./cmd/gamecolor -sweep -max-order 12 -distances 1,2,3 -format csv > grid.csv
go run ./cmd/gamecolor -sweep -max-order 12 -distances 1,2,3 > grid.json  # json is the default

# Grade Go against the Python reference (exits non-zero on any disagreement)
python3 scripts/conformance.py --max-order 8
python3 scripts/conformance.py grid.json

# Build the browser solver, then serve the repo statically
./scripts/build_wasm.sh
python3 -m http.server 8000   # then open http://localhost:8000/web/
```

Optional editable install (`pip install -r requirements.txt`) additionally provides the
`game-coloring` console script and `python -m game_coloring`.

## Architecture

Two independent implementations of the same game solver, plus a generator that bridges them.

### Python core (`src/game_coloring/`)

- `graphs.py` — a `Graph` is `tuple[frozenset[int], ...]` (adjacency indexed by vertex),
  intentionally hashable and immutable so game states can be memoized. Builders: path,
  cycle, star, wheel, helm, sunlet, caterpillar, binary tree. `power_graph(g, d)` joins
  every pair within distance `d`; **all solving happens on that power graph**, so the
  solver never reasons about distance itself and generalizing to distance-d touched only
  this one function. `build_square_graph` is the `d=2` alias. Also `diameter` and
  `is_complete`, which drive the degeneracy shortcut.

  Each family's symmetries are declared immediately after its builder (`path_symmetries`,
  `cycle_symmetries`, `wheel_symmetries`, `helm_symmetries`, `sunlet_symmetries`) because
  the builder is the only thing that knows the recipe — `make_cycle_graph`'s `% n` *is*
  the rotation. Keeping them adjacent makes it obvious that changing the numbering means
  changing the symmetries.
- `symmetry.py` — permutation machinery (`compose`, `invert`, `close_group`) plus
  `is_symmetry`, which confirms a claimed relabelling leaves every edge intact. This is
  what makes hand-declared symmetries safe to rely on.
- `solver.py` — the game. Entry points take a `distance` argument defaulting to 2.
  A state is `ColorState = tuple[int, ...]` where `0` means uncolored.
  Alice tries to complete a full proper coloring of G²; Bob tries to reach a *dead vertex*
  (an uncolored vertex with no legal color left). `build_solver(square_graph, color_count)`
  returns an `lru_cache`d `solve(colors, alice_turn) -> bool` closure — the cache is per
  (power graph, k) pair, so build one solver and reuse it rather than calling `alice_wins`
  in a loop when performance matters. `game_chromatic_number(graph, d)` scans k = 0, 1, 2, …
  for the first Alice win, with `game_distance_2_chromatic_number` kept as the `d=2` alias.
  Both short-circuit when the power graph is complete: every move then burns a distinct
  colour, so the answer is `|V|` with no search. `analyze_game` replays one concrete line of play for the CLI's
  `--explain`; that line is illustrative, not a proof of the whole strategy tree.
- `cli.py` — argparse front end. Note it only exposes path/cycle/star, and `__init__.py`
  only re-exports those three, even though `graphs.py` also builds caterpillars and binary
  trees. Those extra families are reachable only through the web export script.

### Web demo (`web/`)

Static files on GitHub Pages, but **not a static solver**: `cmd/wasmsolver` compiles the real
Go solver to WebAssembly and the page calls into it. The browser is just another
cross-compilation target (`GOOS=js GOARCH=wasm`), so `internal/graph` and `internal/game`
are used unchanged.

This replaced a hand-written JavaScript reimplementation of the rules that had to be kept in
step with Python and Go by hand — the last place three copies of the game could silently
disagree. Do not reintroduce one.

- `solver-worker.js` runs the WASM off the main thread. WASM will happily let the solver
  think for seconds on a hard position, and on the main thread that freezes the page.
- `solver-client.js` wraps the worker in promises. Every question about the position is
  therefore **asynchronous**: the page asks, waits, stores the answer in `state.analysis`,
  then renders. It cannot ask "who is winning?" while rendering.
- `layout.js` computes vertex positions for arbitrary `n`, because the sandbox has no fixed
  list of sizes to precompute. It also holds `vertexLabel`, which mirrors the CLI's naming.
- `levels.js` is the curated level list, plain data. There is no generated JSON and no
  export script any more; every level is `(family, n, d, k)` plus prose, and the solver in
  the page works out the rest.

The values crossing the JS boundary are plain objects, not JSON strings: `encoding/json`
costs about 476 KB gzipped because it pulls in reflection, and the visitor downloads that.

`web/solver.wasm` and `web/wasm_exec.js` are build artifacts (`scripts/build_wasm.sh`),
gitignored, rebuilt by the Pages workflow. CI builds them too, because the `js && wasm`
build tag means an ordinary `go build ./...` never compiles that package.

### Go implementation

A port of the Python solver, intended to become the tool actually used for
sweeps while Python stays the reference oracle. Layout is standard Go: root
`go.mod`, `cmd/gamecolor` for the CLI, `internal/graph` and `internal/game`.

- `internal/graph` — `Graph` carries `Neighbors` **and** `Generators`, the
  symmetry generators its builder declared. That slot is the whole reason the
  package exists in this shape: the builder knows the recipe, and discovering
  symmetries from a bare adjacency list is a hard general problem. `Power`
  carries the generators through unchanged, since relabelling cannot change
  distances. `symmetry.go` mirrors the Python module, `IsSymmetry` included.
- `internal/game` — the solver, plus `explain.go`, which replays one concrete
  line of play. Explaining a result is kept separate from computing it: `Analyze`
  mirrors Python's `analyze_game` move for move (verified over 87 cases), and the
  CLI names vertices by role (`hub`, `rim4`, `stub4`) rather than by index, which
  is what makes a reported kill readable. `-explain` requires `-k`; without a
  colour count there is no single game to narrate.

**The Go port must agree with Python exactly.** As of the initial port, 84
(family, n, d) combinations match with zero mismatches. Any divergence is a bug
in the port, not a new result.

**The language was never the win.** The unoptimized port ran about 3x faster
than Python; the real gains came from searching fewer positions:

| Change | H_5 at d=2 | Note |
|---|---|---|
| Python reference | 10.0 s | |
| naive Go port | 3.4 s | ~3x, all of it constant factor |
| packed integer memo key | 1.40 s | positions unchanged; bookkeeping only |
| colour canonicalization | 26.7 ms | 511,269 positions → 5,493 |
| symmetry folding | 8.2 ms | 5,493 → 630 |

`helm_5` at d=1 went from **29 minutes** in Python to **0.079 s**. Cases that
were previously unreachable now finish: `H_7`, `S_8`, `C_14`, `W_16`.

Paths gain almost nothing from symmetry folding — a path has exactly two
symmetries, so the cost nearly cancels the saving. The rim families have 2n.

**Do not port these optimizations back to Python.** The reference is valuable
precisely because it is naive: an oracle that shared a trick could not detect a
flaw in that trick. Conformance being bounded by Python's speed is the intended
trade, not a defect. Optimizations are additionally graded inside Go by
`canonical_test.go`, which runs plain / colours-only / colours+symmetry side by
side and compares every verdict.

**Parallelism is across sweep cells, never across k within a cell** — running a
cell's colour counts concurrently would compute the expensive above-threshold
solves the upward scan deliberately never reaches, the same trap as binary
searching k. Sweep output is written by index so it stays byte-identical
regardless of scheduling, since the conformance harness diffs it.

`-format` picks the sweep's output: `json` (default, the full detail the harness diffs),
`table` (families down, sizes across, `*` marking cells the degeneracy shortcut decided),
or `csv`. Only JSON is contractual — the harness parses it, so changing its shape means
changing `scripts/conformance.py` too.

### Tests

| File | Covers |
|---|---|
| `test_graphs.py` | builder sizes/degrees/error cases, adjacency invariants, `power_graph`, `diameter` |
| `test_symmetry.py` | every declared symmetry passes `is_symmetry`, group sizes, numbering contract, reuse across powers |
| `test_solver.py` | known values across families and distances, degenerate cases, spot-checked verdicts, `analyze_game` consistency |
| `test_paths.py` | original path-only checks |
| `test_cli.py` | exact CLI output strings and argparse exit codes |
| `test_monotonicity.py` | measured properties of `chi_g,d` in k and in d, and why the scan runs upward |
| `support.py` | the `@slow` marker (not collected by discovery) |

Go-side: `internal/graph/{graph,symmetry}_test.go` mirror the Python structural
and symmetry tests; `internal/game/solver_test.go` pins the same known values;
`internal/game/canonical_test.go` is the differential check on the
optimizations; `internal/game/bench_test.go` holds the benchmark cases and
`TestPositionCount`, which reports the number that actually matters when
optimizing.

CI (`.github/workflows/ci.yml`) runs the Python suite, Go fmt/vet/test, the
conformance harness, and `verify_properties.py` on every push.

Every value asserted in `test_solver.py` was cross-checked against a separately written
brute-force solver before being written down — the point is that this suite is the
*reference oracle* a faster implementation gets graded against, so it must not simply
enshrine whatever the current code happens to print. Extend it the same way: derive or
independently verify a value first, then assert it.

`SpotCheckedWinnerTests` deliberately lists each k separately instead of asserting a
threshold, because monotonicity in k is still unverified (see Conventions).

Tests are split by *cost*, not by confidence: the `@slow` grids in `HelmTests` and
`SunletTests` are just as verified as the rest, they simply cost ~2 minutes. Keep the
default suite under about five seconds so it stays worth running on every edit.
