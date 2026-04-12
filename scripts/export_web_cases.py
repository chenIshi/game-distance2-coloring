from __future__ import annotations

import json
import pathlib
import sys
from dataclasses import asdict, dataclass

ROOT = pathlib.Path(__file__).resolve().parents[1]
SRC = ROOT / "src"
if str(SRC) not in sys.path:
    sys.path.insert(0, str(SRC))

from game_coloring.graphs import Graph, build_square_graph, make_cycle_graph, make_path_graph
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
        ),
    ]


def serialize_graph(graph: Graph) -> list[list[int]]:
    return [sorted(neighbors) for neighbors in graph]


def export_case(preset: Preset) -> dict[str, object]:
    square_graph = build_square_graph(preset.graph)
    solve = build_solver(square_graph, preset.color_count)
    start_colors = (0,) * preset.size
    positions = path_positions(preset.size) if preset.family == "path" else cycle_positions(preset.size)

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
        "positions": positions,
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
