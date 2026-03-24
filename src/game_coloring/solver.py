from __future__ import annotations

from dataclasses import dataclass
from functools import lru_cache

from .graphs import Graph, build_square_graph


ColorState = tuple[int, ...]


@dataclass(frozen=True)
class Move:
    player: str
    vertex: int
    color: int


@dataclass(frozen=True)
class Analysis:
    winner: str
    winning_opening_move: Move | None
    example_line: tuple[Move, ...]
    dead_vertices: tuple[int, ...]
    final_colors: ColorState


def legal_colors(square_graph: Graph, colors: ColorState, vertex: int, color_count: int) -> tuple[int, ...]:
    if colors[vertex] != 0:
        return ()

    blocked = {colors[neighbor] for neighbor in square_graph[vertex] if colors[neighbor] != 0}
    return tuple(color for color in range(1, color_count + 1) if color not in blocked)


def has_dead_vertex(square_graph: Graph, colors: ColorState, color_count: int) -> bool:
    for vertex, color in enumerate(colors):
        if color == 0 and not legal_colors(square_graph, colors, vertex, color_count):
            return True
    return False


def dead_vertices(square_graph: Graph, colors: ColorState, color_count: int) -> tuple[int, ...]:
    blocked_vertices: list[int] = []
    for vertex, color in enumerate(colors):
        if color == 0 and not legal_colors(square_graph, colors, vertex, color_count):
            blocked_vertices.append(vertex)
    return tuple(blocked_vertices)


def ordered_moves(square_graph: Graph, colors: ColorState, color_count: int) -> tuple[tuple[int, tuple[int, ...]], ...]:
    moves: list[tuple[int, tuple[int, ...]]] = []

    for vertex, color in enumerate(colors):
        if color != 0:
            continue
        options = legal_colors(square_graph, colors, vertex, color_count)
        if options:
            moves.append((vertex, options))

    moves.sort(key=lambda item: (len(item[1]), item[0]))
    return tuple(moves)


def build_solver(square_graph: Graph, color_count: int):
    @lru_cache(maxsize=None)
    def solve(colors: ColorState, alice_turn: bool) -> bool:
        if all(color != 0 for color in colors):
            return True

        if has_dead_vertex(square_graph, colors, color_count):
            return False

        moves = ordered_moves(square_graph, colors, color_count)

        if alice_turn:
            for vertex, options in moves:
                for chosen_color in options:
                    next_colors = list(colors)
                    next_colors[vertex] = chosen_color
                    if solve(tuple(next_colors), False):
                        return True
            return False

        for vertex, options in moves:
            for chosen_color in options:
                next_colors = list(colors)
                next_colors[vertex] = chosen_color
                if not solve(tuple(next_colors), True):
                    return False
        return True

    return solve


def alice_wins(graph: Graph, color_count: int) -> bool:
    if color_count <= 0:
        return len(graph) == 0

    square_graph = build_square_graph(graph)
    solve = build_solver(square_graph, color_count)
    start_colors = (0,) * len(square_graph)
    return solve(start_colors, True)


def game_distance_2_chromatic_number(graph: Graph) -> int:
    for color_count in range(1, len(graph) + 1):
        if alice_wins(graph, color_count):
            return color_count
    return len(graph) + 1


def analyze_game(graph: Graph, color_count: int) -> Analysis:
    if color_count <= 0:
        return Analysis(
            winner="Alice" if len(graph) == 0 else "Bob",
            winning_opening_move=None,
            example_line=(),
            dead_vertices=tuple(range(len(graph))) if len(graph) != 0 else (),
            final_colors=(0,) * len(graph),
        )

    square_graph = build_square_graph(graph)
    solve = build_solver(square_graph, color_count)
    colors: ColorState = (0,) * len(square_graph)
    alice_turn = True
    line: list[Move] = []
    opening_move: Move | None = None
    alice_can_win = solve(colors, True)

    while True:
        if all(color != 0 for color in colors):
            return Analysis(
                winner="Alice",
                winning_opening_move=opening_move,
                example_line=tuple(line),
                dead_vertices=(),
                final_colors=colors,
            )

        blocked_vertices = dead_vertices(square_graph, colors, color_count)
        if blocked_vertices:
            return Analysis(
                winner="Bob",
                winning_opening_move=opening_move if alice_can_win else None,
                example_line=tuple(line),
                dead_vertices=blocked_vertices,
                final_colors=colors,
            )

        moves = ordered_moves(square_graph, colors, color_count)
        chosen_vertex = -1
        chosen_color = -1

        if alice_turn:
            if alice_can_win:
                for vertex, options in moves:
                    for color in options:
                        next_colors = list(colors)
                        next_colors[vertex] = color
                        if solve(tuple(next_colors), False):
                            chosen_vertex = vertex
                            chosen_color = color
                            break
                    if chosen_vertex != -1:
                        break
            else:
                chosen_vertex, options = moves[0]
                chosen_color = options[0]
        else:
            if alice_can_win:
                chosen_vertex, options = moves[0]
                chosen_color = options[0]
            else:
                for vertex, options in moves:
                    for color in options:
                        next_colors = list(colors)
                        next_colors[vertex] = color
                        if not solve(tuple(next_colors), True):
                            chosen_vertex = vertex
                            chosen_color = color
                            break
                    if chosen_vertex != -1:
                        break

        next_colors = list(colors)
        next_colors[chosen_vertex] = chosen_color
        move = Move(
            player="Alice" if alice_turn else "Bob",
            vertex=chosen_vertex,
            color=chosen_color,
        )
        if opening_move is None and alice_turn:
            opening_move = move
        line.append(move)
        colors = tuple(next_colors)
        alice_turn = not alice_turn
