# Findings

Measured results for the distance-`d` colouring game, `chi_g,d`. Everything here was
computed with the tools in this repo and cross-checked between two independent
implementations; nothing is proved. Reproduce any of it with the commands in each section.

Notation and vertex numbering are fixed in `CLAUDE.md`. Briefly: `W_n`, `H_n` and `S_n`
take `n` as the **rim** size, so `W_6` has 7 vertices, `H_6` has 13, and `S_6` has 12.

---

## 1. The whole grid

```
./gamecolor -sweep -max-order 14 -distances 1,2,3 -format table
```

```
distance 1
    family   1   2   3  4  5  6  7  8  9  10  11  12  13  14
      path  1*  2*   2  3  3  3  3  3  3   3   3   3   3   3
     cycle   .   .  3*  3  3  3  3  3  3   3   3   3   3   3
      star  2*   2   2  2  2  2  2  2  2   2   2   2   2   .
     wheel   .   .  4*  3  4  3  4  4  4   4   4   4   4   .
      helm   .   .   4  4  4  4  .  .  .   .   .   .   .   .
    sunlet   .   .   3  4  3  4  3  .  .   .   .   .   .   .

distance 2
    family   1   2   3   4   5   6   7   8    9   10   11   12   13  14
      path  1*  2*  3*   3   3   4   4   4    4    4    4    4    4   5
     cycle   .   .  3*  4*  5*   5   4   5    5    5    5    5    5   5
      star  2*  3*  4*  5*  6*  7*  8*  9*  10*  11*  12*  13*  14*   .
     wheel   .   .  4*  5*  6*  7*  8*  9*  10*  11*  12*  13*  14*   .
      helm   .   .   6   5   6   7   .   .    .    .    .    .    .   .
    sunlet   .   .   5   5   6   6   6   .    .    .    .    .    .   .

distance 3
    family   1   2   3   4   5   6   7   8    9   10   11   12   13  14
      path  1*  2*  3*  4*   4   5   4   5    5    6    5    6    6   6
     cycle   .   .  3*  4*  5*  6*  7*   7    5    6    6    7    7   7
      star  2*  3*  4*  5*  6*  7*  8*  9*  10*  11*  12*  13*  14*   .
     wheel   .   .  4*  5*  6*  7*  8*  9*  10*  11*  12*  13*  14*   .
      helm   .   .  7*   7   9  10   .   .    .    .    .    .    .   .
    sunlet   .   .  6*   7   8   8   8   .    .    .    .    .    .   .
```

`*` marks cells where the distance-`d` power graph is complete, so the answer is forced.

---

## 2. When `d` reaches the diameter, the game stops being a game

**If `d >= diam(G)` then `chi_g,d(G) = |V|`.** Checked across 156 cases, no exceptions.

The reason is stronger than "it's easy to compute". Once every vertex blocks every other,
each move burns one fresh colour regardless of who plays it or where. With `k >= |V|` a
fresh colour always remains and Alice cannot lose; with `k < |V|` the colours run out after
`k` moves and whatever is left is dead. **Neither player has a decision that affects the
outcome** — the game becomes strategy-free.

This retires whole families rather than isolated cases:

| Family | diameter | trivial from |
|---|---|---|
| Wheel `W_n` | 2 | `d = 2` |
| Star `K_1,n` | 2 | `d = 2` |
| `C_3`, `C_4`, `C_5` | ≤ 2 | `d = 2` |
| `C_6`, `C_7` | 3 | `d = 3` |
| Helm `H_5` | 4 | `d = 4` |

A wheel's hub is one step from every rim vertex, so any two rim vertices are two steps apart
through it. **Wheels and stars therefore only carry information at `d = 1`.** Helms and
sunlets have pendants reaching past the hub, so their diameters grow and they stay
interesting across a range of `d`.

The solver checks for this before searching, which is why `W_17` at `d = 2` is instant while
at `d = 1` it takes 31 seconds.

---

## 3. Bigger graphs can need fewer colours

`chi_g,d` is **not** monotone in `n`, in any family. One example from each:

| Comparison | smaller | larger |
|---|---|---|
| `C_6` vs `C_7` at `d=2` | 5 | **4** |
| `W_5` vs `W_6` at `d=1` | 4 | **3** |
| `H_3` vs `H_4` at `d=2` | 6 | **5** |
| `S_4` vs `S_5` at `d=1` | 4 | **3** |
| `P_6` vs `P_7` at `d=3` | 5 | **4** |

This is a real property, not noise. Parity of the structure matters more than its size.
Never smooth or interpolate a results table on the assumption that values climb with `n`.

---

## 4. Monotone in `k` and in `d`, everywhere measured

```
python3 scripts/verify_properties.py --max-vertices 8
```

- **In `k`**: Alice never loses a game she could win with fewer colours. Every win vector is
  a clean block of Bob wins followed by a block of Alice wins. 95 cases, no violations.
- **In `d`**: `chi_g,d` never decreases as `d` grows. 32 graphs, no violations.

The second is less obvious than it sounds. `G^d` is a subgraph of `G^(d+1)`, and the game
chromatic number is famously **not** monotone under subgraphs, so "more edges must be
harder" is exactly the kind of reasoning that fails here. It had to be measured.

Both are evidence, not proof.

### Why the k-scan still counts upward

Monotonicity in `k` would make a binary search over `k` *legal*. It would also be *slower*.
Binary search assumes every probe costs the same, and here they differ by two orders of
magnitude — cost grows steeply with `k`, not with closeness to the answer:

```
C_9 at d=2, answer is 5
   k=4  Bob     1.32s   11.1%      k=8  Alice   2.90s   24.2%
   k=5  Alice   0.25s    2.1%      k=9  Alice   5.28s   44.1%
```

Counting up stops at 5 and never touches 6–9. A binary search over `[0,9]` would probe
above the answer, which is the expensive half: about 3.7s against 1.6s.

The cost grows with `k` for two compounding reasons. More colours means more distinct
positions — `(k+1)^n` rather than `2^n`. And with many colours **no vertex can ever be
trapped**, so no branch terminates early and every game is played to full depth. With few
colours most branches die within a few moves.

---

## 5. Wheels at `d = 1`: Alice's hub move buys tempo

`chi_g,1(W_n)` for `n = 3..18`: `4 3 4 3 4 4 4 4 4 4 4 4 4 4 4 4`

Constant 4, except rims of 4 and 6. Against colouring the same graph alone:

| rim | parity | alone | with Bob | penalty |
|---|---|---|---|---|
| 3, 5, 7, 9, 11 | odd | 4 | 4 | none |
| 4, 6 | even | 3 | 3 | none |
| **8, 10, 12, 14…** | even | 3 | **4** | **+1** |

**Odd rims are never penalised.** An odd cycle needs 3 colours by itself and the hub needs a
4th, so 4 is unavoidable and Bob cannot make it worse.

**Even rims are where the game bites**, and the mechanism is turn order. On `W_4` with 3
colours:

```
Alice opens by colouring the HUB   ->  Alice wins
Alice opens on a RIM vertex        ->  Bob wins
```

Taking the hub does two things: it fixes the hub's colour, confining every rim vertex to the
other two, and it **passes the turn**. What remains is exactly "2-colour a cycle, Bob to
move first", and that game goes:

```
2 colours on a bare cycle:
        Alice first    Bob first
  C_4    Bob wins      Alice wins
  C_6    Bob wins      Alice wins
  C_8    Bob wins      Bob wins     <- the break
```

The "Bob first" column is Alice winning at 4 and 6 and losing from 8 — **exactly the wheel's
pattern**. The hub is not merely an extra colour to pay for; it is a move Alice can spend to
hand Bob the disadvantage of opening. That gift is worth a colour on rims of 4 and 6, and
stops being enough from 8.

So the low points in the trend are precisely the sizes where a good opening exists for
Alice, and the shortcut is narrow: one wrong first move and 3 colours fails.

---

## 6. Sunlets at `d = 1`: Alice loses by being forced to move

`chi_g,1(S_n)` for `n = 3..9`: `3 4 3 4 3 4 3` — perfectly alternating, **odd is easy and
even is hard**, which is backwards from the usual intuition about odd cycles.

| rim | parity | alone | with Bob | penalty |
|---|---|---|---|---|
| 3, 5, 7 | odd | 3 | 3 | none |
| **4, 6** | even | 2 | **4** | **+2** |

`+2` is the largest gap anywhere in this project.

**A tested-and-rejected explanation.** The obvious guess is that the pendants hand Bob spare
harmless moves which steal Alice's tempo on the rim. Measured, and false — the bare rim at 3
colours does not care who moves first:

```
Bare cycle with 3 colours, n = 4..9:  Alice wins whether she moves first or second.
```

**What is actually happening.** The sunlet *as a whole* is turn-order sensitive at even `n`:

```
Sunlet with 3 colours:
  n     Alice first   Bob first
  4         Bob         Alice     <- first player loses
  5        Alice        Alice
  6         Bob         Alice     <- first player loses
  7        Alice        Alice
```

On an even sunlet, **whoever moves first loses**. Were Bob forced to open, Alice would win.
The rules make Alice open, so she loses. She is not being outplayed; she is losing because
she is compelled to move at all when every available move damages her position — zugzwang.
The fourth colour buys enough slack that her forced opening stops being fatal.

Odd sunlets have no such problem: Alice wins whoever starts, confirmed at n = 5 and 7.

### The two families are opposite cases of the same idea

Both come down to who is forced to move, and they land on opposite sides:

- **Wheel** — Alice has a genuinely good opening, the hub, which both constrains the rim and
  dumps the awkward first rim move onto Bob. She hands the problem away.
- **Sunlet** — there is no hub and no equivalent move. Every option is poisoned and she must
  pick one.

Helms show neither effect: `chi_g,1(H_n) = 4` flat for `n = 3..6`, no wobble.

---

## 7. Open questions

- **Does the sunlet alternation continue?** Confirmed to `n = 9`. `S_10` and beyond are
  within reach of the Go solver but were not run.
- **Is the wheel really 4 forever?** Confirmed to `n = 18` (19 vertices). Cost roughly
  triples per extra rim vertex, so `n = 20` is about 15 minutes.
- **Why is even-rim zugzwang specific to sunlets?** Helms have pendants too and show no
  wobble at all. The pendants alone do not explain it.
- **Do monotonicity in `k` and `d` hold in general,** or only on the sizes reachable here?
  Both are measured, neither is proved.

---

## Reproducing

```bash
go build -o gamecolor ./cmd/gamecolor

# section 1
./gamecolor -sweep -max-order 14 -distances 1,2,3 -format table

# a single cell
./gamecolor -family sunlet -n 6 -d 1
./gamecolor -family wheel -n 4 -k 3 -d 1

# sections 2 and 4 (this one runs on the naive Python reference, so it is slow;
# --max-vertices 8 takes a couple of minutes, 10 takes far longer)
python3 scripts/verify_properties.py --max-vertices 8

# grading Go against the naive Python reference
python3 scripts/conformance.py --max-order 8
```

Turn-order probes (sections 5 and 6) use the Python reference directly, since they ask about
a position mid-game rather than a whole graph:

```python
import sys; sys.path.insert(0, "src")
from game_coloring import graphs as G
from game_coloring.solver import build_solver

solve = build_solver(G.power_graph(G.make_sunlet_graph(6), 1), 3)
solve((0,) * 12, True)    # Alice to move first
solve((0,) * 12, False)   # Bob to move first
```
