"""Check the properties of chi_g,d that the solver deliberately does not assume.

Two questions, both measured rather than taken on faith:

  Q1  monotone in k -- can Alice lose a game she could have won with fewer
      colours? If she never can, a binary search over k becomes *legal* (though
      see the note on cost below, which is why the code still scans upward).

  Q2  monotone in d -- can reaching further ever make Alice's job easier? This
      one looks obvious and is not: G^d is a subgraph of G^(d+1), and the game
      chromatic number is famously NOT monotone under subgraphs.

Exits non-zero if any violation is found, so it can be run in CI.

    python3 scripts/verify_properties.py
    python3 scripts/verify_properties.py --max-vertices 11 --distances 1 2 3 4
"""

from __future__ import annotations

import argparse
import pathlib
import sys
import time

ROOT = pathlib.Path(__file__).resolve().parents[1]
SRC = ROOT / "src"
if str(SRC) not in sys.path:
    sys.path.insert(0, str(SRC))

from game_coloring import graphs as G
from game_coloring.solver import build_solver


def win_vector(graph: G.Graph, distance: int) -> tuple[list[bool], str]:
    """Alice's verdict for every k from 0 to |V|, and how it was obtained."""
    reach = G.power_graph(graph, distance)
    size = len(reach)

    if G.is_complete(reach):
        return [k >= size for k in range(size + 1)], "shortcut"

    outcomes = [size == 0]
    for color_count in range(1, size + 1):
        outcomes.append(build_solver(reach, color_count)((0,) * size, True))
    return outcomes, "search"


def candidate_graphs(limit: int):
    for n in range(2, limit + 1):
        yield f"P_{n}", G.make_path_graph(n)
    for n in range(3, limit + 1):
        yield f"C_{n}", G.make_cycle_graph(n)
    for n in range(1, limit):
        yield f"K_1,{n}", G.make_star_graph(n)
    for n in range(3, limit):
        yield f"W_{n}", G.make_wheel_graph(n)
    for n in range(3, limit // 2 + 1):
        yield f"H_{n}", G.make_helm_graph(n)
    for n in range(3, limit // 2 + 1):
        yield f"S_{n}", G.make_sunlet_graph(n)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--max-vertices", type=int, default=8)
    parser.add_argument("--distances", type=int, nargs="+", default=[1, 2, 3])
    args = parser.parse_args()

    distances = sorted(args.distances)
    chi_by_graph: dict[str, dict[int, int]] = {}
    k_violations: list[str] = []

    header = f"{'case':8s} {'d':>2s} {'|V|':>4s} {'how':>9s} {'chi':>4s}  outcome by k"
    print(header)
    print("-" * len(header))

    for label, graph in candidate_graphs(args.max_vertices):
        if len(graph) > args.max_vertices:
            continue
        for distance in distances:
            started = time.time()
            outcomes, how = win_vector(graph, distance)
            elapsed = time.time() - started

            monotone = all(
                outcomes[i] <= outcomes[i + 1] for i in range(len(outcomes) - 1)
            )
            chi = outcomes.index(True) if True in outcomes else None
            if chi is not None:
                chi_by_graph.setdefault(label, {})[distance] = chi
            if not monotone:
                k_violations.append(f"{label} d={distance}: {outcomes}")

            shown = "".join("A" if value else "b" for value in outcomes)
            flag = "" if monotone else "  *** VIOLATION ***"
            print(
                f"{label:8s} {distance:2d} {len(graph):4d} {how:>9s} "
                f"{-1 if chi is None else chi:4d}  {shown}{flag}  ({elapsed:.1f}s)",
                flush=True,
            )

    d_violations = []
    for label, per_distance in sorted(chi_by_graph.items()):
        ordered = [per_distance[d] for d in sorted(per_distance)]
        if any(ordered[i] > ordered[i + 1] for i in range(len(ordered) - 1)):
            d_violations.append(f"{label}: {ordered}")

    print()
    print(f"Q1  monotone in k: {len(k_violations)} violation(s)")
    for line in k_violations:
        print(f"      {line}")
    print(f"Q2  chi non-decreasing in d: {len(d_violations)} violation(s)")
    for line in d_violations:
        print(f"      {line}")

    if k_violations or d_violations:
        print("\nFAILED: a property the code relies on does not hold.")
        return 1

    print("\nAll checked properties hold. Note this is evidence, not proof.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
