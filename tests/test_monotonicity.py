"""Measured properties of chi_g,d that the code must NOT assume.

Both questions here were open when the distance-d generalization landed. They
are checked rather than taken on faith, because each one, if false, would break
an "obvious" optimization:

  * monotone in k -- would license replacing the k-sweep with a binary search
  * monotone in d -- looks obvious but is not, because G^d is a subgraph of
    G^(d+1) and the game chromatic number is famously NOT monotone under
    subgraphs: a subgraph can need strictly more colours than the graph
    containing it.
"""

from __future__ import annotations

import pathlib
import sys
import unittest
from functools import lru_cache

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1] / "src"))
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from support import slow

from game_coloring import graphs as G
from game_coloring.solver import build_solver, game_chromatic_number


@lru_cache(maxsize=None)
def _win_vector(graph: G.Graph, distance: int) -> tuple[bool, ...]:
    """Whether Alice wins, for every k from 0 to |V| inclusive.

    Deliberately evaluates every k rather than stopping at the first win, since
    the whole point is to look at what happens above the threshold.
    """
    reach = G.power_graph(graph, distance)
    size = len(reach)

    if G.is_complete(reach):
        return tuple(k >= size for k in range(size + 1))

    outcomes = [size == 0]
    for color_count in range(1, size + 1):
        solve = build_solver(reach, color_count)
        outcomes.append(solve((0,) * size, True))
    return tuple(outcomes)


def win_vector(graph: G.Graph, distance: int) -> list[bool]:
    # Several tests below examine the same vectors; a graph is hashable, so
    # compute each one once.
    return list(_win_vector(graph, distance))


QUICK_GRAPHS = (
    [(f"P_{n}", G.make_path_graph(n)) for n in range(2, 8)]
    + [(f"C_{n}", G.make_cycle_graph(n)) for n in range(3, 8)]
    + [(f"K_1,{n}", G.make_star_graph(n)) for n in range(1, 6)]
    + [(f"W_{n}", G.make_wheel_graph(n)) for n in range(3, 7)]
    + [("H_3", G.make_helm_graph(3)), ("S_3", G.make_sunlet_graph(3))]
)

EXPENSIVE_GRAPHS = [
    ("P_8", G.make_path_graph(8)),
    ("C_8", G.make_cycle_graph(8)),
    ("W_7", G.make_wheel_graph(7)),
    ("H_4", G.make_helm_graph(4)),
    ("S_4", G.make_sunlet_graph(4)),
]

DISTANCES = (1, 2, 3)


class MonotonicityInColoursTests(unittest.TestCase):
    """Alice must never lose a game she could win with fewer colours."""

    def check(self, graphs) -> None:
        for label, graph in graphs:
            for distance in DISTANCES:
                with self.subTest(graph=label, d=distance):
                    outcomes = win_vector(graph, distance)
                    for index in range(len(outcomes) - 1):
                        if outcomes[index]:
                            self.assertTrue(
                                outcomes[index + 1],
                                f"{label} d={distance}: Alice wins with k={index} "
                                f"but loses with k={index + 1}: {outcomes}",
                            )

    def test_quick_grid(self) -> None:
        self.check(QUICK_GRAPHS)

    @slow
    def test_large_grid(self) -> None:
        self.check(EXPENSIVE_GRAPHS)

    def test_win_vector_shape_is_a_single_step(self) -> None:
        # Equivalent phrasing of the same property, kept because it is the form
        # a reader recognises: a block of Bob wins then a block of Alice wins.
        for label, graph in QUICK_GRAPHS:
            for distance in DISTANCES:
                with self.subTest(graph=label, d=distance):
                    outcomes = win_vector(graph, distance)
                    threshold = outcomes.index(True) if True in outcomes else len(outcomes)
                    self.assertEqual(outcomes, [k >= threshold for k in range(len(outcomes))])

    def test_threshold_matches_the_reported_chromatic_number(self) -> None:
        for label, graph in QUICK_GRAPHS:
            for distance in DISTANCES:
                with self.subTest(graph=label, d=distance):
                    outcomes = win_vector(graph, distance)
                    self.assertEqual(
                        outcomes.index(True),
                        game_chromatic_number(graph, distance),
                    )


class MonotonicityInDistanceTests(unittest.TestCase):
    """Reaching further can only make Alice's job harder, never easier."""

    def check(self, graphs) -> None:
        for label, graph in graphs:
            with self.subTest(graph=label):
                values = [game_chromatic_number(graph, d) for d in DISTANCES]
                for index in range(len(values) - 1):
                    self.assertLessEqual(
                        values[index],
                        values[index + 1],
                        f"{label}: chi fell from d={DISTANCES[index]} to "
                        f"d={DISTANCES[index + 1]}: {values}",
                    )

    def test_quick_grid(self) -> None:
        self.check(QUICK_GRAPHS)

    @slow
    def test_large_grid(self) -> None:
        self.check(EXPENSIVE_GRAPHS)


class ScanDirectionTests(unittest.TestCase):
    """Why game_chromatic_number scans upward instead of binary searching.

    Cost grows steeply with k, so the upward scan stops at the threshold and
    never pays for the expensive large-k solves. A binary search over
    [0, |V|] would evaluate values above the threshold, which is exactly the
    costly half. Monotonicity holding does not make the swap worthwhile.
    """

    def test_scan_stops_at_the_threshold(self) -> None:
        for label, graph in [("C_7", G.make_cycle_graph(7)), ("P_7", G.make_path_graph(7))]:
            with self.subTest(graph=label):
                threshold = game_chromatic_number(graph, 2)
                outcomes = win_vector(graph, 2)
                # Every k below the threshold is a Bob win, so scanning upward
                # terminates exactly there and never touches larger k.
                self.assertTrue(all(not value for value in outcomes[:threshold]))
                self.assertTrue(outcomes[threshold])


if __name__ == "__main__":
    unittest.main()
