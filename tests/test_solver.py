from __future__ import annotations

import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1] / "src"))
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from game_coloring.graphs import (
    build_square_graph,
    is_complete,
    make_cycle_graph,
    make_helm_graph,
    make_path_graph,
    make_star_graph,
    make_sunlet_graph,
    make_wheel_graph,
    normalize_graph,
    power_graph,
)
from support import slow

from game_coloring.solver import (
    alice_wins,
    analyze_game,
    build_solver,
    game_chromatic_number,
    game_distance_2_chromatic_number,
    legal_colors,
)


def complete_graph(n: int):
    return normalize_graph([{u for u in range(n) if u != v} for v in range(n)])


class KnownValueTests(unittest.TestCase):
    """Values cross-checked against an independently written brute-force solver."""

    def test_paths(self) -> None:
        expected = {1: 1, 2: 2, 3: 3, 4: 3, 5: 3, 6: 4, 7: 4}
        for n, value in expected.items():
            with self.subTest(n=n):
                self.assertEqual(game_distance_2_chromatic_number(make_path_graph(n)), value)

    def test_cycles(self) -> None:
        expected = {3: 3, 4: 4, 5: 5, 6: 5, 7: 4, 8: 5}
        for n, value in expected.items():
            with self.subTest(n=n):
                self.assertEqual(game_distance_2_chromatic_number(make_cycle_graph(n)), value)

    def test_cycle_values_are_not_monotone_in_n(self) -> None:
        # Guards a genuine property of the problem that looks like a bug when
        # first encountered: a longer cycle can need fewer colors.
        self.assertEqual(game_distance_2_chromatic_number(make_cycle_graph(6)), 5)
        self.assertEqual(game_distance_2_chromatic_number(make_cycle_graph(7)), 4)

    def test_stars_need_one_colour_per_vertex(self) -> None:
        # K_1,n has diameter 2, so its square is complete on n+1 vertices.
        for leaves in range(1, 6):
            with self.subTest(leaves=leaves):
                graph = make_star_graph(leaves)
                self.assertEqual(game_distance_2_chromatic_number(graph), leaves + 1)

    def test_complete_graphs_need_one_colour_per_vertex(self) -> None:
        for n in range(2, 6):
            with self.subTest(n=n):
                self.assertEqual(game_distance_2_chromatic_number(complete_graph(n)), n)


class DegenerateCaseTests(unittest.TestCase):
    def test_complete_square_forces_chi_equal_to_vertex_count(self) -> None:
        # Whenever the square graph is complete every move burns a distinct
        # colour, so Alice wins exactly when k >= |V|. No search required.
        for label, graph in [
            ("C_3", make_cycle_graph(3)),
            ("C_4", make_cycle_graph(4)),
            ("C_5", make_cycle_graph(5)),
            ("K_1,4", make_star_graph(4)),
        ]:
            with self.subTest(graph=label):
                self.assertEqual(build_square_graph(graph), complete_graph(len(graph)))
                self.assertEqual(game_distance_2_chromatic_number(graph), len(graph))

    def test_zero_colours_never_wins_on_a_nonempty_graph(self) -> None:
        self.assertFalse(alice_wins(make_path_graph(1), 0))
        self.assertFalse(alice_wins(make_path_graph(4), 0))

    def test_negative_colour_count_behaves_like_zero(self) -> None:
        self.assertFalse(alice_wins(make_path_graph(1), -1))


class SpotCheckedWinnerTests(unittest.TestCase):
    """Individually verified (graph, k) verdicts.

    Deliberately spelled out one k at a time rather than asserting a threshold,
    because monotonicity in k is still unverified for this game and must not be
    baked into the tests.
    """

    CASES = {
        "C_6": (make_cycle_graph(6), {3: False, 4: False, 5: True, 6: True, 7: True}),
        "C_7": (make_cycle_graph(7), {3: False, 4: True, 5: True, 6: True, 7: True}),
        "P_6": (make_path_graph(6), {2: False, 3: False, 4: True, 5: True, 6: True}),
        "K_1,3": (make_star_graph(3), {2: False, 3: False, 4: True, 5: True, 6: True}),
    }

    def test_spot_checked_verdicts(self) -> None:
        for label, (graph, verdicts) in self.CASES.items():
            for color_count, expected in verdicts.items():
                with self.subTest(graph=label, k=color_count):
                    self.assertEqual(alice_wins(graph, color_count), expected)


class AnalysisConsistencyTests(unittest.TestCase):
    """analyze_game must report a line of play that is actually legal."""

    CASES = [
        ("P_2 k=1", make_path_graph(2), 1),
        ("P_4 k=3", make_path_graph(4), 3),
        ("P_6 k=4", make_path_graph(6), 4),
        ("C_6 k=4", make_cycle_graph(6), 4),
        ("C_6 k=5", make_cycle_graph(6), 5),
        ("C_7 k=4", make_cycle_graph(7), 4),
        ("K_1,3 k=3", make_star_graph(3), 3),
        ("K_1,3 k=4", make_star_graph(3), 4),
    ]

    def test_winner_agrees_with_alice_wins(self) -> None:
        for label, graph, color_count in self.CASES:
            with self.subTest(case=label):
                analysis = analyze_game(graph, color_count)
                expected = "Alice" if alice_wins(graph, color_count) else "Bob"
                self.assertEqual(analysis.winner, expected)

    def test_example_line_is_a_legal_alternating_game(self) -> None:
        for label, graph, color_count in self.CASES:
            with self.subTest(case=label):
                analysis = analyze_game(graph, color_count)
                square = build_square_graph(graph)
                colors = [0] * len(graph)

                for index, move in enumerate(analysis.example_line):
                    self.assertEqual(move.player, "Alice" if index % 2 == 0 else "Bob")
                    self.assertEqual(colors[move.vertex], 0, "recoloured a vertex")
                    allowed = legal_colors(square, tuple(colors), move.vertex, color_count)
                    self.assertIn(move.color, allowed, "played an illegal colour")
                    colors[move.vertex] = move.color

                self.assertEqual(tuple(colors), analysis.final_colors)

    def test_alice_wins_produce_a_complete_proper_colouring(self) -> None:
        for label, graph, color_count in self.CASES:
            analysis = analyze_game(graph, color_count)
            if analysis.winner != "Alice":
                continue
            with self.subTest(case=label):
                square = build_square_graph(graph)
                self.assertEqual(analysis.dead_vertices, ())
                self.assertTrue(all(color != 0 for color in analysis.final_colors))
                for vertex, neighbors in enumerate(square):
                    for neighbor in neighbors:
                        self.assertNotEqual(
                            analysis.final_colors[vertex],
                            analysis.final_colors[neighbor],
                            "final colouring is not proper on the square graph",
                        )
                self.assertIsNotNone(analysis.winning_opening_move)
                self.assertEqual(analysis.winning_opening_move, analysis.example_line[0])

    def test_bob_wins_produce_a_genuinely_dead_vertex(self) -> None:
        for label, graph, color_count in self.CASES:
            analysis = analyze_game(graph, color_count)
            if analysis.winner != "Bob":
                continue
            with self.subTest(case=label):
                square = build_square_graph(graph)
                self.assertNotEqual(analysis.dead_vertices, ())
                for vertex in analysis.dead_vertices:
                    self.assertEqual(analysis.final_colors[vertex], 0)
                    self.assertEqual(
                        legal_colors(square, analysis.final_colors, vertex, color_count),
                        (),
                        "reported a dead vertex that still has a legal colour",
                    )


class EmptyGraphTests(unittest.TestCase):
    """The empty graph needs zero colours, and both entry points must agree.

    A disagreement here is harmless in isolation but would surface later as a
    spurious conformance failure against a second implementation.
    """

    def test_alice_wins_with_no_colours(self) -> None:
        self.assertTrue(alice_wins((), 0))

    def test_chromatic_number_is_zero(self) -> None:
        self.assertEqual(game_distance_2_chromatic_number(()), 0)
        self.assertEqual(game_distance_2_chromatic_number(make_path_graph(0)), 0)

    def test_entry_points_agree(self) -> None:
        value = game_distance_2_chromatic_number(())
        self.assertTrue(alice_wins((), value))

    def test_analysis_reports_alice(self) -> None:
        analysis = analyze_game((), 0)
        self.assertEqual(analysis.winner, "Alice")
        self.assertEqual(analysis.dead_vertices, ())
        self.assertEqual(analysis.example_line, ())


class VertexCountAlwaysSufficesTests(unittest.TestCase):
    """With one colour per vertex Alice cannot lose.

    A vertex has at most |V| - 1 neighbours, so it can never see all |V|
    colours and no vertex can ever go dead. This is what makes the fallback
    branch of game_distance_2_chromatic_number unreachable.
    """

    def test_alice_wins_with_one_colour_per_vertex(self) -> None:
        graphs = [
            ("P_5", make_path_graph(5)),
            ("C_6", make_cycle_graph(6)),
            ("C_7", make_cycle_graph(7)),
            ("K_1,4", make_star_graph(4)),
            ("K_5", complete_graph(5)),
        ]
        for label, graph in graphs:
            with self.subTest(graph=label):
                self.assertTrue(alice_wins(graph, len(graph)))

    def test_chromatic_number_never_exceeds_vertex_count(self) -> None:
        graphs = [make_path_graph(n) for n in range(1, 7)]
        graphs += [make_cycle_graph(n) for n in range(3, 8)]
        graphs += [make_star_graph(n) for n in range(1, 5)]
        for graph in graphs:
            with self.subTest(size=len(graph)):
                self.assertLessEqual(game_distance_2_chromatic_number(graph), len(graph))


class DistanceGeneralizationTests(unittest.TestCase):
    """All values below were cross-checked against an independent brute force."""

    def test_distance_two_matches_the_original_entry_point(self) -> None:
        graphs = [make_path_graph(n) for n in range(1, 7)]
        graphs += [make_cycle_graph(n) for n in range(3, 8)]
        graphs += [make_star_graph(n) for n in range(1, 5)]
        for graph in graphs:
            with self.subTest(size=len(graph)):
                self.assertEqual(
                    game_chromatic_number(graph, 2),
                    game_distance_2_chromatic_number(graph),
                )

    def test_distance_defaults_to_two(self) -> None:
        graph = make_cycle_graph(7)
        self.assertEqual(game_chromatic_number(graph), game_chromatic_number(graph, 2))
        self.assertEqual(alice_wins(graph, 4), alice_wins(graph, 4, 2))

    def test_paths_at_distance_one(self) -> None:
        for n in range(4, 8):
            with self.subTest(n=n):
                self.assertEqual(game_chromatic_number(make_path_graph(n), 1), 3)

    def test_paths_at_distance_three(self) -> None:
        expected = {4: 4, 5: 4, 6: 5, 7: 4}
        for n, value in expected.items():
            with self.subTest(n=n):
                self.assertEqual(game_chromatic_number(make_path_graph(n), 3), value)

    def test_cycles_at_distance_one(self) -> None:
        for n in range(5, 8):
            with self.subTest(n=n):
                self.assertEqual(game_chromatic_number(make_cycle_graph(n), 1), 3)


class WheelTests(unittest.TestCase):
    def test_distance_one(self) -> None:
        expected = {3: 4, 4: 3, 5: 4, 6: 3, 7: 4, 8: 4}
        for rim, value in expected.items():
            with self.subTest(rim=rim):
                self.assertEqual(game_chromatic_number(make_wheel_graph(rim), 1), value)

    def test_trivial_at_every_distance_of_two_or_more(self) -> None:
        # A wheel has diameter 2, so its power graph is complete from d=2 on and
        # the answer is forced to be one colour per vertex.
        for rim in range(3, 9):
            for distance in (2, 3, 4):
                with self.subTest(rim=rim, d=distance):
                    graph = make_wheel_graph(rim)
                    self.assertTrue(is_complete(power_graph(graph, distance)))
                    self.assertEqual(game_chromatic_number(graph, distance), rim + 1)


class HelmTests(unittest.TestCase):
    """Grids split by cost, not by confidence: every value here was verified."""

    QUICK = {1: {3: 4, 4: 4}, 2: {3: 6, 4: 5}, 3: {3: 7}}
    EXPENSIVE = {1: {5: 4}, 2: {5: 6}, 3: {4: 7, 5: 9}}

    def check(self, grid) -> None:
        for distance, expected in grid.items():
            for rim, value in expected.items():
                with self.subTest(rim=rim, d=distance):
                    graph = make_helm_graph(rim)
                    self.assertEqual(game_chromatic_number(graph, distance), value)

    def test_small_grid(self) -> None:
        self.check(self.QUICK)

    @slow
    def test_large_grid(self) -> None:
        self.check(self.EXPENSIVE)


class SunletTests(unittest.TestCase):
    QUICK = {1: {3: 3, 4: 4, 5: 3}, 2: {3: 5, 4: 5}, 3: {3: 6}}
    EXPENSIVE = {2: {5: 6}, 3: {4: 7, 5: 8}}

    def check(self, grid) -> None:
        for distance, expected in grid.items():
            for rim, value in expected.items():
                with self.subTest(rim=rim, d=distance):
                    graph = make_sunlet_graph(rim)
                    self.assertEqual(game_chromatic_number(graph, distance), value)

    def test_small_grid(self) -> None:
        self.check(self.QUICK)

    @slow
    def test_large_grid(self) -> None:
        self.check(self.EXPENSIVE)


class NonMonotonicityInSizeTests(unittest.TestCase):
    """Bigger graphs can need fewer colours. Real, and easy to mistake for a bug."""

    def test_examples_across_families(self) -> None:
        cases = [
            ("C_6 vs C_7 at d=2", make_cycle_graph(6), make_cycle_graph(7), 2, 5, 4),
            ("W_5 vs W_6 at d=1", make_wheel_graph(5), make_wheel_graph(6), 1, 4, 3),
            ("H_3 vs H_4 at d=2", make_helm_graph(3), make_helm_graph(4), 2, 6, 5),
            ("S_4 vs S_5 at d=1", make_sunlet_graph(4), make_sunlet_graph(5), 1, 4, 3),
            ("P_6 vs P_7 at d=3", make_path_graph(6), make_path_graph(7), 3, 5, 4),
        ]
        for label, smaller, larger, distance, low, high in cases:
            with self.subTest(case=label):
                self.assertEqual(game_chromatic_number(smaller, distance), low)
                self.assertEqual(game_chromatic_number(larger, distance), high)
                self.assertGreater(low, high, "expected the larger graph to need fewer")


class CompleteShortcutTests(unittest.TestCase):
    """The shortcut skips the search, so prove the search agrees with it."""

    def test_search_agrees_with_the_shortcut(self) -> None:
        cases = [
            ("W_4 d=2", make_wheel_graph(4), 2),
            ("W_5 d=2", make_wheel_graph(5), 2),
            ("C_5 d=2", make_cycle_graph(5), 2),
            ("K_1,4 d=2", make_star_graph(4), 2),
            ("S_3 d=3", make_sunlet_graph(3), 3),
        ]
        for label, graph, distance in cases:
            reach = power_graph(graph, distance)
            self.assertTrue(is_complete(reach), label)
            size = len(reach)
            for color_count in range(max(size - 2, 1), size + 2):
                with self.subTest(case=label, k=color_count):
                    searched = build_solver(reach, color_count)((0,) * size, True)
                    self.assertEqual(alice_wins(graph, color_count, distance), searched)
                    self.assertEqual(searched, color_count >= size)


if __name__ == "__main__":
    unittest.main()
