# game-distance2-coloring

A solver for the **distance-d colouring game** on small graph families, used to compute
`chi_g,d` and look for patterns in it.

## The game

Alice and Bob take turns, Alice first. A move colours one uncoloured vertex with a colour
not already used by any vertex within distance `d`. Alice (the *maker*) wins if every
vertex ends up coloured. Bob (the *breaker*) wins the moment some uncoloured vertex has no
legal colour left — a **dead vertex**.

`chi_g,d(G)` is the least number of colours for which Alice wins. `d` defaults to 2.

There are two implementations. **Go** is the one to actually use. **Python** is kept
deliberately slow and simple as the reference the Go version is graded against.

## Quick start

```bash
export PATH="$HOME/.local/go/bin:$PATH"   # wherever your Go lives
go build -o gamecolor ./cmd/gamecolor
./gamecolor -family cycle -n 7
```

```
C_7: chi_g,2 = 4
  k=1: Bob
  k=2: Bob
  k=3: Bob
  k=4: Alice
```

A 7-cycle needs 4 colours. The lines below are the working: 1–3 colours let Bob trap a
vertex, 4 is the first count Alice can win with. It stops there rather than checking
larger counts, which are far more expensive (see *Why the scan counts upward*).

## Options

| Flag | Meaning | Values |
|---|---|---|
| `-family` | which family | `path` `cycle` `star` `wheel` `helm` `sunlet` |
| `-n` | size | vertices for path/cycle, **rim count** otherwise |
| `-d` | colouring distance | default `2` |
| `-k` | test one colour count instead of finding the least | optional |
| `-explain` | with `-k`, replay one concrete line of play | optional |

```bash
./gamecolor -family sunlet -n 5 -d 3      # S_5: chi_g,3 = 8
./gamecolor -family helm -n 4 -k 5        # H_4 with k=5 at d=2: Alice
./gamecolor -family helm -n 4 -k 4        # H_4 with k=4 at d=2: Bob
```

### Seeing a game played out

`-explain` replays one line and, when Bob wins, says exactly why the trapped vertex is
trapped. Vertices are named by their role, so the reason reads directly:

```
$ ./gamecolor -family helm -n 4 -k 3 -d 1 -explain
H_4 with k=3 at d=1: Bob

one line of play:
  Alice: hub   -> 1
  Bob:   rim1  -> 2
  Alice: rim2  -> 3
  Bob:   stub1 -> 1
  Alice: rim3  -> 2
  Bob:   stub4 -> 3
dead vertex: rim4
  blocked by hub=1, rim1=2, rim3=2, stub4=3
  between them that is all 3 colours -- nothing left for it
(one line of play, not a proof of the whole strategy tree)
```

`rim4`'s two rim neighbours are both colour 2 — the rim's alternating pattern was never
broken. Bob used the pendant to supply the third colour. That single line is why a helm
needs a colour more than the wheel it is built from.

### Families, sizes, and numbering

`-n` means different things per family, and the vertex numbering is a fixed contract that
the declared graph symmetries depend on.

| Family | Notation | `-n` | Vertices | Numbering |
|---|---|---|---|---|
| Path | `P_n` | n | n | `0..n-1` along the path |
| Cycle | `C_n` | n | n | `0..n-1` around the cycle |
| Star | `K_1,n` | leaves | n+1 | centre `0`, leaves `1..n` |
| Wheel | `W_n` | rim | n+1 | hub `0`, rim `1..n` |
| Helm | `H_n` | rim | 2n+1 | hub `0`, rim `1..n`, pendant of rim `i` at `n+i` |
| Sunlet | `S_n` | rim | 2n | rim `0..n-1`, pendant of rim `i` at `n+i` |

`-family wheel -n 5` is a wheel with **5 rim vertices, 6 in total**. Sunlet is also written
`C_n ⊙ K_1`; some papers call it a crown graph, but that name usually means something else.

## Sweeps

Every family, every size, every distance, in one go — fanned out across cores:

```bash
./gamecolor -sweep -max-order 12 -distances 1,2,3 -format table
```

```
distance 2
    family   1   2   3   4   5   6   7   8    9   10   11  12
      path  1*  2*  3*   3   3   4   4   4    4    4    4   4
     cycle   .   .  3*  4*  5*   5   4   5    5    5    5   5
      star  2*  3*  4*  5*  6*  7*  8*  9*  10*  11*  12*   .
     wheel   .   .  4*  5*  6*  7*  8*  9*  10*  11*  12*   .
      helm   .   .   6   5   6   .   .   .    .    .    .   .
    sunlet   .   .   5   5   6   6   .   .    .    .    .   .
```

`*` marks a cell where the distance-`d` power graph is complete, so `chi = |V|` with no
search needed. `.` means that size does not exist for the family, or exceeds `-max-order`.

Three formats: `-format table` to read, `-format csv` to load into a spreadsheet, and
`-format json` (the default) which carries the full detail the conformance harness diffs.

Measured results, and the reasoning behind them, are collected in
[`FINDINGS.md`](FINDINGS.md).

## Things that look like bugs and are not

**A bigger graph can need fewer colours.** `C_6` needs 5 but `C_7` needs 4. This happens in
every family — `W_5`→`W_6`, `H_3`→`H_4`, `S_4`→`S_5`, `P_6`→`P_7` at d=3. Parity of the
structure matters more than its size. Never smooth or interpolate a results table on the
assumption that values climb with `n`.

**Wheels and stars are trivial at every `d >= 2`.** Both have diameter 2, so their power
graph is complete and the answer is forced to `|V|`. They only carry information at `d = 1`.
That is the entire starred region of the table above.

**Why the scan counts upward.** Alice never loses a game she could win with fewer colours
(measured over 90+ cases, no violations), so a binary search over `k` would be *legal*. It
would also be slower. Solve cost grows steeply with `k`, not with closeness to the answer —
on `C_9` at d=2 the largest `k` costs 44% of the total while the answer itself costs 2%.
Counting up stops at the answer and never touches the expensive part.

## Python reference

```bash
python run_cli.py cycle 7
python run_cli.py path 2 --k 1 --explain
```

Only knows path, cycle, and star, only at distance 2, and is thousands of times slower.
It stays naive on purpose. It is the oracle the Go implementation is graded against, and an
oracle that shared Go's optimisations could not detect a flaw in them.

## Tests

```bash
python3 -m unittest discover -s tests             # ~5s, the everyday suite
GAME_COLORING_SLOW=1 python3 -m unittest discover -s tests   # ~2min, adds the big grids
go test ./...                                     # ~5s
python3 scripts/conformance.py --max-order 8      # grades Go against Python
python3 scripts/verify_properties.py              # re-measures the claimed properties
```

The last two exit non-zero on disagreement. CI runs all of them on every push.

Every value asserted in the Python suite was cross-checked against a separately written
brute-force solver before being written down, so the suite states what is true rather than
recording what the code happens to print.

## Web demo

A playable demo lives in `web/`, running **the same Go solver** compiled to WebAssembly —
not a reimplementation of it.

```bash
./scripts/build_wasm.sh          # compiles cmd/wasmsolver to web/solver.wasm
python3 -m http.server 8000      # then open http://localhost:8000/web/
```

It has curated levels drawn from the findings (several in pairs differing by one colour or
one vertex), and a sandbox where you pick family, size, distance and colour count freely.
Nothing is precomputed: the solver runs in a Web Worker in your browser, so any combination
works. When a vertex is trapped the page names the neighbours blocking each colour, the same
way `-explain` does.

`web/solver.wasm` and `web/wasm_exec.js` are build artifacts and are gitignored; the Pages
workflow builds them before deploying.

## Layout

```
cmd/gamecolor/      Go CLI and sweep
internal/graph/     graph builders, declared symmetries, power graphs
internal/game/      the solver
src/game_coloring/  Python reference
scripts/            conformance harness, property verifier, web export
tests/              Python test suite
```

`CLAUDE.md` carries the working conventions and the reasoning behind the design decisions.
`FINDINGS.md` carries the measured results.
