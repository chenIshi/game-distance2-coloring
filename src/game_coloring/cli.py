from __future__ import annotations

import argparse

from .graphs import make_cycle_graph, make_path_graph, make_star_graph
from .solver import alice_wins, analyze_game, game_distance_2_chromatic_number


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="game_coloring",
        description="Solve the distance-2 coloring game on small graph families.",
    )
    subparsers = parser.add_subparsers(dest="graph_type", required=True)

    path_parser = subparsers.add_parser("path", help="solve a path graph P_n")
    path_parser.add_argument("n", type=int, help="number of vertices")
    path_parser.add_argument("--k", type=int, help="number of colors to test")
    path_parser.add_argument("--explain", action="store_true", help="show a sample line of play")

    cycle_parser = subparsers.add_parser("cycle", help="solve a cycle graph C_n")
    cycle_parser.add_argument("n", type=int, help="number of vertices")
    cycle_parser.add_argument("--k", type=int, help="number of colors to test")
    cycle_parser.add_argument("--explain", action="store_true", help="show a sample line of play")

    star_parser = subparsers.add_parser("star", help="solve a star graph K_1,n")
    star_parser.add_argument("leaves", type=int, help="number of leaves")
    star_parser.add_argument("--k", type=int, help="number of colors to test")
    star_parser.add_argument("--explain", action="store_true", help="show a sample line of play")

    return parser


def build_graph(args: argparse.Namespace):
    if args.graph_type == "path":
        return make_path_graph(args.n), f"P_{args.n}"
    if args.graph_type == "cycle":
        return make_cycle_graph(args.n), f"C_{args.n}"
    return make_star_graph(args.leaves), f"K_1,{args.leaves}"


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()
    graph, label = build_graph(args)

    if args.k is not None:
        winner = "Alice" if alice_wins(graph, args.k) else "Bob"
        print(f"{label} with k={args.k}: {winner}")
        if args.explain:
            analysis = analyze_game(graph, args.k)
            if analysis.winner == "Alice":
                if analysis.winning_opening_move is not None:
                    move = analysis.winning_opening_move
                    print(f"winning opening move: {move.player} colors v{move.vertex} with color {move.color}")
                print("sample continuation:")
            else:
                print("one failed line:")

            for move in analysis.example_line:
                print(f"  {move.player}: v{move.vertex} -> color {move.color}")

            if analysis.dead_vertices:
                dead_list = ", ".join(f"v{vertex}" for vertex in analysis.dead_vertices)
                print(f"dead vertex set: {dead_list}")
            else:
                print("final coloring completed successfully")
        return 0

    chromatic_number = game_distance_2_chromatic_number(graph)
    print(f"{label}: chi_g,2 = {chromatic_number}")
    for color_count in range(1, chromatic_number + 1):
        winner = "Alice" if alice_wins(graph, color_count) else "Bob"
        print(f"  k={color_count}: {winner}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
