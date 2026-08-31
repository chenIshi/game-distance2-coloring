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

### Why the pendant matters: it is a third neighbour

The reason the bare rim is indifferent to turn order is stronger than "Alice plays well".
With 3 colours on a bare cycle, **no vertex can ever be trapped by anyone.** Every vertex
has exactly two neighbours, two neighbours block at most two colours, and there are three,
so one is always left. Alice cannot lose that game even playing deliberately badly, which is
why it looks turn-order insensitive: the outcome was never in doubt.

A pendant raises each rim vertex from two neighbours to three — exactly the number of
colours. That is the entire structural change:

```
bare cycle:  rim vertex has 2 neighbours  ->  at most 2 colours blocked  ->  can never die
sunlet:      rim vertex has 3 neighbours  ->  all 3 colours can be blocked  ->  can die
```

One stub per vertex takes the game from *impossible to lose* to *losable*, and only then
does being forced to move first start to cost anything. A concrete kill on `S_4` with 3
colours:

```
  Alice: colours rim0 with 1
  Bob  : colours rim2 with 2
  Alice: colours rim1 with 3
  Bob  : colours stub7 (hanging off rim3) with 3

  DEAD: rim3's neighbours are rim0=1, rim2=2, stub7=3 -- all three colours, nothing left.
```

`rim3` died to its two ring neighbours plus its own stub. Strip the stub and the same
position is harmless.

**What is actually happening.** Once losing is possible, the even sunlet becomes turn-order
sensitive and the first move is a liability:

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

This is not a matter of Alice missing a clever opening. On `S_4` with 3 colours **every one
of her eight possible opening moves loses**, rim and pendant alike, while on the bare `C_4`
every opening wins. There is nothing to find.

Odd sunlets have no such problem: Alice wins whoever starts, confirmed at n = 5 and 7.

### The two families are opposite cases of the same idea

Both come down to who is forced to move, and they land on opposite sides:

- **Wheel** — Alice has a genuinely good opening, the hub, which both constrains the rim and
  dumps the awkward first rim move onto Bob. She hands the problem away.
- **Sunlet** — there is no hub and no equivalent move. Every option is poisoned and she must
  pick one.

Helms show neither effect: `chi_g,1(H_n) = 4` flat for `n = 3..6`, no wobble.

---

## 7. Helms at `d = 1`: the pendant takes back the wheel's trick

`chi_g,1(H_n) = 4` flat for `n = 3..7`. No wobble at all, where the wheel dips and the
sunlet alternates. The flatness is not "nothing interesting happens" — it is a third
distinct behaviour.

### Three families, three different games

At 3 colours, the count that decides every even rim:

```
             Alice first   Bob first
  W_4           Alice        Bob        <- first player WINS
  H_4            Bob         Bob        <- turn order is irrelevant
  S_4            Bob         Alice      <- first player LOSES
```

On the wheel, moving first is an advantage and Alice cashes it in. On the sunlet it is a
liability and she is stuck with it. On the helm **turn order does not matter at all**: Bob
wins whoever starts. Alice is not unlucky about move order, she is simply beaten.

That is why helms never wobble. Wobble comes from turn-order effects, and helms have none.

### The pendant destroys the hub trick

`W_4` and `W_6` get away with 3 colours by the tempo argument in section 5. Hang one pendant
on each rim vertex and `H_4` and `H_6` need 4. The tool shows exactly why:

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

`rim4`'s two rim neighbours are **both colour 2**. The alternating rim pattern was never
broken — Alice defended the rim perfectly, and it did not help. Bob ignored the rim and used
the pendant to deliver the third colour. On a bare wheel `rim4` would see only `{1, 2}` and
colour 3 would still be free.

So the hub hands Alice a trick and the pendants take it straight back. Compare with section
6: on a sunlet the pendant is what makes a trap possible at all; on a helm it is what
defeats the one defence the hub had provided.

### At `d = 2` the game costs helms nothing

```
n   |V|  hub+rim clique  chi(H_n^2)  chi_g,2  penalty
3     7        4              5          6      +1
4     9        5              5          5     none
5    11        6              6          6     none
6    13        7              7          7     none
7    15        8              8          8     none
8    17        9              9          9     none
```

**From `n = 4` on, Bob is powerless.** Alice achieves the best possible colouring every time.

The answer is forced by structure. At `d = 2` the hub is within two steps of everything, and
any two rim vertices are two steps apart through it, so the hub together with the entire rim
is a **clique of size `n+1`** in the power graph. That alone forces `n+1` colours, and Alice
always reaches it:

> `chi_g,2(H_n) = n + 1` for `n >= 4`.

This also explains an apparent violation of section 3. The row `6 5 6 7 8` looks like another
case of a larger graph needing fewer colours, but it is not a dip in the underlying trend:
`n = 3` is the single exceptional case carrying a `+1` penalty, sitting above an otherwise
smooth line. Everything from `n = 4` is exactly `n+1`.

---

## 8. Open questions

- **Does the sunlet alternation continue?** Confirmed to `n = 9`. `S_10` and beyond are
  within reach of the Go solver but were not run.
- **Is the wheel really 4 forever?** Confirmed to `n = 18` (19 vertices). Cost roughly
  triples per extra rim vertex, so `n = 20` is about 15 minutes.
- **Why is even worse than odd for sunlets?** Section 6 explains why sunlets are losable at
  all — the pendant gives each rim vertex three neighbours, matching the three colours. That
  argument says nothing about parity, and odd sunlets have exactly the same degree-3 rim
  vertices, yet `S_3`, `S_5`, `S_7` are fine at 3 colours while `S_4`, `S_6`, `S_8` are not.
  Unexplained.
- **Why do helms show no wobble at all?** Partly answered in section 7: turn order is
  irrelevant for helms, and wobble comes from turn-order effects. What remains open is why
  the hub removes the sensitivity that the sunlet has, given both families carry the same
  pendants.
- **Is `chi_g,2(H_n) = n + 1` exact for all `n >= 4`?** Confirmed to `n = 8`, and the
  hub-plus-rim clique gives the lower bound for free. Whether Alice always attains it is
  measured, not proved.
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

Traced games come straight from the tool:

```bash
./gamecolor -family helm -n 4 -k 3 -d 1 -explain     # section 7's kill
./gamecolor -family sunlet -n 4 -k 3 -d 1 -explain   # section 6's kill
./gamecolor -family wheel -n 4 -k 3 -d 1 -explain    # section 5's hub opening
```

Turn-order probes (sections 5, 6 and 7) still use the Python reference directly, because
they ask who wins when **Bob** moves first, and the CLI always starts Alice:

```python
import sys; sys.path.insert(0, "src")
from game_coloring import graphs as G
from game_coloring.solver import build_solver

solve = build_solver(G.power_graph(G.make_sunlet_graph(6), 1), 3)
solve((0,) * 12, True)    # Alice to move first
solve((0,) * 12, False)   # Bob to move first
```
