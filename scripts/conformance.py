"""Grade the Go implementation against the Python reference.

Go emits a JSON grid (`gamecolor -sweep`); this recomputes every field with the
Python implementation and reports any disagreement. Exits non-zero on the first
sign of divergence, so it can gate CI.

Python is the reference here not because it is better but because it is the
implementation that has been cross-checked against an independently written
brute force. Any mismatch is a bug in the Go port until proven otherwise.

    go run ./cmd/gamecolor -sweep -max-order 8 > grid.json
    python3 scripts/conformance.py grid.json

    # or let it build and run Go itself
    python3 scripts/conformance.py --max-order 8
"""

from __future__ import annotations

import argparse
import json
import pathlib
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parents[1]
SRC = ROOT / "src"
if str(SRC) not in sys.path:
    sys.path.insert(0, str(SRC))

from game_coloring import graphs as G
from game_coloring.solver import alice_wins, game_chromatic_number
from game_coloring.symmetry import close_group

# Mirrors sweepFamilies in cmd/gamecolor/sweep.go. Divergence here is itself a
# thing worth catching, so the harness checks the case list matches too.
FAMILIES = {
    "path": (G.make_path_graph, G.path_symmetries),
    "cycle": (G.make_cycle_graph, G.cycle_symmetries),
    "star": (G.make_star_graph, lambda n: ()),
    "wheel": (G.make_wheel_graph, G.wheel_symmetries),
    "helm": (G.make_helm_graph, G.helm_symmetries),
    "sunlet": (G.make_sunlet_graph, G.sunlet_symmetries),
}


def reference_case(family: str, n: int, distance: int) -> dict:
    build, symmetries = FAMILIES[family]
    graph = build(n)
    reach = G.power_graph(graph, distance)
    span = G.diameter(graph)
    connected = span != float("inf")
    generators = [list(p) for p in symmetries(n)]
    chi = game_chromatic_number(graph, distance)

    return {
        "order": len(graph),
        "edges": sum(len(neighbors) for neighbors in graph) // 2,
        "diameter": int(span) if connected else -1,
        "connected": connected,
        "neighbors": [sorted(neighbors) for neighbors in graph],
        "generators": generators,
        "groupSize": len(close_group(len(graph), symmetries(n))),
        "powerEdges": sum(len(neighbors) for neighbors in reach) // 2,
        "complete": G.is_complete(reach),
        "chi": chi,
        "verdicts": [alice_wins(graph, k, distance) for k in range(chi + 1)],
    }


# Ordered so the most diagnostic difference is reported first: if the graphs
# themselves differ, the game values differing is a consequence, not a finding.
FIELDS = (
    "order",
    "edges",
    "neighbors",
    "diameter",
    "connected",
    "generators",
    "groupSize",
    "powerEdges",
    "complete",
    "chi",
    "verdicts",
)


def compare(case: dict) -> list[str]:
    label = f"{case['family']}_{case['n']} d={case['d']}"
    if case["family"] not in FAMILIES:
        return [f"{label}: Go emitted a family Python does not know"]

    expected = reference_case(case["family"], case["n"], case["d"])
    problems = []
    for field in FIELDS:
        if case[field] != expected[field]:
            problems.append(
                f"{label}: {field}\n"
                f"      go     = {case[field]}\n"
                f"      python = {expected[field]}"
            )
    return problems


def load_grid(args: argparse.Namespace) -> dict:
    if args.grid:
        return json.loads(pathlib.Path(args.grid).read_text(encoding="utf-8"))

    command = [
        args.go,
        "run",
        "./cmd/gamecolor",
        "-sweep",
        "-max-order",
        str(args.max_order),
        "-distances",
        args.distances,
    ]
    completed = subprocess.run(command, cwd=ROOT, capture_output=True, text=True)
    if completed.returncode != 0:
        raise SystemExit(f"go sweep failed:\n{completed.stderr}")
    return json.loads(completed.stdout)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("grid", nargs="?", help="a JSON grid; omit to run Go directly")
    parser.add_argument("--max-order", type=int, default=8)
    parser.add_argument("--distances", default="1,2,3")
    parser.add_argument("--go", default="go", help="path to the go binary")
    args = parser.parse_args()

    grid = load_grid(args)
    cases = grid["cases"]
    if not cases:
        raise SystemExit("the grid is empty; nothing was compared")

    problems: list[str] = []
    for case in cases:
        problems.extend(compare(case))

    checked = len(cases) * len(FIELDS)
    print(f"compared {len(cases)} cases x {len(FIELDS)} fields = {checked} points")

    if problems:
        print(f"\n{len(problems)} MISMATCH(ES):\n")
        for problem in problems:
            print(f"  {problem}")
        print("\nFAILED: the implementations disagree.")
        return 1

    print("All fields agree.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
