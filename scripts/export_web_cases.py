from __future__ import annotations

import json
import pathlib
import sys
from dataclasses import dataclass

ROOT = pathlib.Path(__file__).resolve().parents[1]
SRC = ROOT / "src"
if str(SRC) not in sys.path:
    sys.path.insert(0, str(SRC))

from game_coloring.graphs import (
    Graph,
    build_square_graph,
    make_binary_tree_graph,
    make_caterpillar_graph,
    make_cycle_graph,
    make_path_graph,
)
from game_coloring.solver import build_solver


PLAYER_ALICE = "Alice"
PLAYER_BOB = "Bob"


@dataclass(frozen=True)
class Preset:
    case_id: str
    title: str
    family: str
    size: int
    color_count: int
    difficulty: str
    summary: str
    challenge: bool
    graph: Graph
    positions: list[dict[str, float]]


def path_positions(size: int) -> list[dict[str, float]]:
    if size <= 0:
        return []

    start_x = 80
    end_x = 440
    y = 150
    if size == 1:
        return [{"x": (start_x + end_x) / 2, "y": y}]

    spacing = (end_x - start_x) / (size - 1)
    return [{"x": round(start_x + spacing * index, 2), "y": y} for index in range(size)]


def cycle_positions(size: int) -> list[dict[str, float]]:
    if size == 0:
        return []

    center_x = 260
    center_y = 170
    radius = 110 if size <= 5 else 130
    positions: list[dict[str, float]] = []
    for index in range(size):
        angle = (-0.5 + 2 * index / size) * 3.141592653589793
        positions.append(
            {
                "x": round(center_x + radius * __import__("math").cos(angle), 2),
                "y": round(center_y + radius * __import__("math").sin(angle), 2),
            }
        )
    return positions


def caterpillar_positions(spine_length: int, leaf_counts: tuple[int, ...]) -> list[dict[str, float]]:
    positions: list[dict[str, float]] = []
    start_x = 90
    end_x = 430
    base_y = 185
    spacing = 0 if spine_length <= 1 else (end_x - start_x) / (spine_length - 1)

    for vertex in range(spine_length):
        positions.append({"x": round(start_x + spacing * vertex, 2), "y": base_y})

    for spine_vertex, leaf_count in enumerate(leaf_counts):
        x = positions[spine_vertex]["x"]
        for leaf_index in range(leaf_count):
            y = 95 - 48 * leaf_index if leaf_index % 2 == 0 else 275 + 48 * (leaf_index // 2)
            positions.append({"x": x, "y": y})

    return positions


def binary_tree_positions(depth: int) -> list[dict[str, float]]:
    positions: list[dict[str, float]] = []
    width = 360
    center_x = 260
    start_y = 70
    row_gap = 90

    for level in range(depth + 1):
        nodes_in_level = 2**level
        step = 0 if nodes_in_level == 1 else width / (nodes_in_level - 1)
        level_start_x = center_x if nodes_in_level == 1 else center_x - width / 2
        for index in range(nodes_in_level):
            positions.append(
                {
                    "x": round(level_start_x + step * index, 2),
                    "y": start_y + row_gap * level,
                }
            )

    return positions


def make_presets() -> list[Preset]:
    return [
        Preset(
            case_id="path-p2-k1",
            title="Path P2 with 1 color",
            family="path",
            size=2,
            color_count=1,
            difficulty="tutorial",
            summary="A fast Bob win that teaches what a dead vertex is.",
            challenge=False,
            graph=make_path_graph(2),
            positions=path_positions(2),
        ),
        Preset(
            case_id="path-p3-k3",
            title="Path P3 with 3 colors",
            family="path",
            size=3,
            color_count=3,
            difficulty="core",
            summary="A short success case where Alice can finish the coloring.",
            challenge=False,
            graph=make_path_graph(3),
            positions=path_positions(3),
        ),
        Preset(
            case_id="path-p4-k3",
            title="Path P4 with 3 colors",
            family="path",
            size=4,
            color_count=3,
            difficulty="core",
            summary="A small but less obvious path case.",
            challenge=False,
            graph=make_path_graph(4),
            positions=path_positions(4),
        ),
        Preset(
            case_id="path-p5-k3",
            title="Path P5 with 3 colors",
            family="path",
            size=5,
            color_count=3,
            difficulty="core",
            summary="A longer path where planning ahead starts to matter.",
            challenge=False,
            graph=make_path_graph(5),
            positions=path_positions(5),
        ),
        Preset(
            case_id="path-p6-k4",
            title="Path P6 with 4 colors",
            family="path",
            size=6,
            color_count=4,
            difficulty="challenge",
            summary="Optional challenge: a larger path with more room for mistakes.",
            challenge=True,
            graph=make_path_graph(6),
            positions=path_positions(6),
        ),
        Preset(
            case_id="cycle-c3-k3",
            title="Cycle C3 with 3 colors",
            family="cycle",
            size=3,
            color_count=3,
            difficulty="core",
            summary="The triangle shows cycle pressure immediately.",
            challenge=False,
            graph=make_cycle_graph(3),
            positions=cycle_positions(3),
        ),
        Preset(
            case_id="cycle-c4-k4",
            title="Cycle C4 with 4 colors",
            family="cycle",
            size=4,
            color_count=4,
            difficulty="core",
            summary="A clean contrast with path behavior on a small cycle.",
            challenge=False,
            graph=make_cycle_graph(4),
            positions=cycle_positions(4),
        ),
        Preset(
            case_id="cycle-c5-k4",
            title="Cycle C5 with 4 colors",
            family="cycle",
            size=5,
            color_count=4,
            difficulty="core",
            summary="A Bob-winning cycle that teaches why enough colors matter.",
            challenge=False,
            graph=make_cycle_graph(5),
            positions=cycle_positions(5),
        ),
        Preset(
            case_id="cycle-c6-k5",
            title="Cycle C6 with 5 colors",
            family="cycle",
            size=6,
            color_count=5,
            difficulty="challenge",
            summary="Optional challenge: a larger cycle that needs careful play.",
            challenge=True,
            graph=make_cycle_graph(6),
            positions=cycle_positions(6),
        ),
        Preset(
            case_id="cycle-c7-k4",
            title="Cycle C7 with 4 colors",
            family="cycle",
            size=7,
            color_count=4,
            difficulty="challenge",
            summary="Optional challenge: odd cycle pressure with a longer board.",
            challenge=True,
            graph=make_cycle_graph(7),
            positions=cycle_positions(7),
        ),
        Preset(
            case_id="cycle-c8-k5",
            title="Cycle C8 with 5 colors",
            family="cycle",
            size=8,
            color_count=5,
            difficulty="challenge",
            summary="Optional challenge: a larger even cycle with more room to plan.",
            challenge=True,
            graph=make_cycle_graph(8),
            positions=cycle_positions(8),
        ),
        Preset(
            case_id="caterpillar-t1-k3",
            title="Caterpillar T1 with 3 colors",
            family="tree",
            size=6,
            color_count=3,
            difficulty="core",
            summary="A tree trap case: Bob can still create a dead vertex here.",
            challenge=False,
            graph=make_caterpillar_graph(3, (1, 1, 1)),
            positions=caterpillar_positions(3, (1, 1, 1)),
        ),
        Preset(
            case_id="caterpillar-t1-k4",
            title="Caterpillar T1 with 4 colors",
            family="tree",
            size=6,
            color_count=4,
            difficulty="core",
            summary="The same caterpillar becomes finishable once one more color is available.",
            challenge=False,
            graph=make_caterpillar_graph(3, (1, 1, 1)),
            positions=caterpillar_positions(3, (1, 1, 1)),
        ),
        Preset(
            case_id="caterpillar-t2-k4",
            title="Caterpillar T2 with 4 colors",
            family="tree",
            size=8,
            color_count=4,
            difficulty="core",
            summary="A longer caterpillar where Bob can still force a trap with only 4 colors.",
            challenge=False,
            graph=make_caterpillar_graph(4, (1, 1, 1, 1)),
            positions=caterpillar_positions(4, (1, 1, 1, 1)),
        ),
        Preset(
            case_id="caterpillar-t3-k4",
            title="Caterpillar T3 with 4 colors",
            family="tree",
            size=8,
            color_count=4,
            difficulty="challenge",
            summary="A longer caterpillar path where Alice can still finish if she plans the order well.",
            challenge=True,
            graph=make_caterpillar_graph(5, (1, 0, 1, 0, 1)),
            positions=caterpillar_positions(5, (1, 0, 1, 0, 1)),
        ),
        Preset(
            case_id="binary-b1-k3",
            title="Binary Tree B1 with 3 colors",
            family="tree",
            size=3,
            color_count=3,
            difficulty="core",
            summary="A tiny binary tree that introduces the branching pattern without much pressure.",
            challenge=False,
            graph=make_binary_tree_graph(1),
            positions=binary_tree_positions(1),
        ),
        Preset(
            case_id="binary-b2-k3",
            title="Binary Tree B2 with 3 colors",
            family="tree",
            size=7,
            color_count=3,
            difficulty="core",
            summary="A binary-tree trap case where Bob can still force a dead vertex.",
            challenge=False,
            graph=make_binary_tree_graph(2),
            positions=binary_tree_positions(2),
        ),
        Preset(
            case_id="binary-b2-k4",
            title="Binary Tree B2 with 4 colors",
            family="tree",
            size=7,
            color_count=4,
            difficulty="challenge",
            summary="A small binary tree where Alice can still force a full coloring.",
            challenge=True,
            graph=make_binary_tree_graph(2),
            positions=binary_tree_positions(2),
        ),
    ]


def serialize_graph(graph: Graph) -> list[list[int]]:
    return [sorted(neighbors) for neighbors in graph]


def export_case(preset: Preset) -> dict[str, object]:
    square_graph = build_square_graph(preset.graph)
    solve = build_solver(square_graph, preset.color_count)
    start_colors = (0,) * preset.size

    return {
        "id": preset.case_id,
        "title": preset.title,
        "family": preset.family,
        "size": preset.size,
        "k": preset.color_count,
        "difficulty": preset.difficulty,
        "challenge": preset.challenge,
        "summary": preset.summary,
        "graph": serialize_graph(preset.graph),
        "squareGraph": serialize_graph(square_graph),
        "positions": preset.positions,
        "initialWinner": PLAYER_ALICE if solve(start_colors, True) else PLAYER_BOB,
    }


def main() -> int:
    cases = [export_case(preset) for preset in make_presets()]
    payload = {
        "title": "Distance-2 Coloring Game",
        "description": "Preset teaching cases for a beginner-friendly web demo.",
        "cases": cases,
    }

    output_path = ROOT / "web" / "data" / "cases.json"
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(payload, indent=2), encoding="utf-8")
    print(f"Wrote {output_path.relative_to(ROOT)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
