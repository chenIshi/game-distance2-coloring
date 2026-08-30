from __future__ import annotations

import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1] / "src"))

from game_coloring import graphs as G
from game_coloring.symmetry import (
    close_group,
    compose,
    identity,
    invert,
    is_permutation,
    is_symmetry,
)


def declared_families():
    """(label, graph, declared generators, expected group size)."""
    for n in range(2, 8):
        yield f"P_{n}", G.make_path_graph(n), G.path_symmetries(n), 2
    for n in range(3, 9):
        yield f"C_{n}", G.make_cycle_graph(n), G.cycle_symmetries(n), 2 * n
    for n in range(3, 8):
        yield f"W_{n}", G.make_wheel_graph(n), G.wheel_symmetries(n), 2 * n
    for n in range(3, 7):
        yield f"H_{n}", G.make_helm_graph(n), G.helm_symmetries(n), 2 * n
    for n in range(3, 7):
        yield f"S_{n}", G.make_sunlet_graph(n), G.sunlet_symmetries(n), 2 * n


class DeclaredSymmetryTests(unittest.TestCase):
    """Every symmetry a builder declares must survive the checker.

    This is the safety net that makes declaring them by hand acceptable: an
    over-reported symmetry would corrupt answers, and this catches it.
    """

    def test_generators_are_symmetries(self) -> None:
        for label, graph, generators, _ in declared_families():
            with self.subTest(graph=label):
                for generator in generators:
                    self.assertTrue(is_symmetry(graph, generator), f"{label}: {generator}")

    def test_whole_group_are_symmetries(self) -> None:
        for label, graph, generators, _ in declared_families():
            with self.subTest(graph=label):
                for element in close_group(len(graph), generators):
                    self.assertTrue(is_symmetry(graph, element), f"{label}: {element}")

    def test_group_sizes(self) -> None:
        for label, graph, generators, expected in declared_families():
            with self.subTest(graph=label):
                self.assertEqual(len(close_group(len(graph), generators)), expected)


class NumberingContractTests(unittest.TestCase):
    """Symmetries must respect the vertex numbering fixed in CLAUDE.md."""

    def test_wheel_and_helm_hubs_never_move(self) -> None:
        for n in range(3, 7):
            with self.subTest(n=n):
                for element in close_group(n + 1, G.wheel_symmetries(n)):
                    self.assertEqual(element[0], 0)
                for element in close_group(2 * n + 1, G.helm_symmetries(n)):
                    self.assertEqual(element[0], 0)

    def test_helm_pendants_stay_pendants(self) -> None:
        for n in range(3, 7):
            with self.subTest(n=n):
                pendants = set(range(n + 1, 2 * n + 1))
                for element in close_group(2 * n + 1, G.helm_symmetries(n)):
                    self.assertEqual({element[p] for p in pendants}, pendants)

    def test_sunlet_pendants_stay_pendants(self) -> None:
        for n in range(3, 7):
            with self.subTest(n=n):
                pendants = set(range(n, 2 * n))
                for element in close_group(2 * n, G.sunlet_symmetries(n)):
                    self.assertEqual({element[p] for p in pendants}, pendants)

    def test_pendants_follow_their_rim_vertex(self) -> None:
        for n in range(3, 7):
            with self.subTest(n=n):
                for element in close_group(2 * n, G.sunlet_symmetries(n)):
                    for rim in range(n):
                        self.assertEqual(element[n + rim], n + element[rim])


class RejectionTests(unittest.TestCase):
    def test_rejects_non_permutations(self) -> None:
        graph = G.make_cycle_graph(4)
        self.assertFalse(is_symmetry(graph, (0, 0, 2, 3)))  # repeated index
        self.assertFalse(is_symmetry(graph, (0, 1, 2)))  # wrong length
        self.assertFalse(is_symmetry(graph, (0, 1, 2, 4)))  # out of range

    def test_rejects_a_permutation_that_breaks_an_edge(self) -> None:
        # Swapping two adjacent vertices of a square is fine; swapping only
        # 0 and 1 while pinning 2 and 3 is not, because edge 1-2 would have to
        # become edge 0-2, and there is no such edge.
        self.assertFalse(is_symmetry(G.make_cycle_graph(4), (1, 0, 2, 3)))

    def test_rejects_moving_a_wheel_hub(self) -> None:
        graph = G.make_wheel_graph(4)
        swap_hub_with_rim = (1, 0, 2, 3, 4)
        self.assertFalse(is_symmetry(graph, swap_hub_with_rim))


class PowerGraphReuseTests(unittest.TestCase):
    """A symmetry of G is automatically a symmetry of every power of G.

    Relabelling cannot change distances, so symmetries are computed once per
    graph and reused for every distance.
    """

    def test_symmetries_survive_every_power(self) -> None:
        for label, graph, generators, _ in declared_families():
            group = close_group(len(graph), generators)
            for distance in range(1, 5):
                with self.subTest(graph=label, d=distance):
                    powered = G.power_graph(graph, distance)
                    for element in group:
                        self.assertTrue(is_symmetry(powered, element))


class GroupAlgebraTests(unittest.TestCase):
    def test_identity_is_present_and_neutral(self) -> None:
        group = close_group(6, G.cycle_symmetries(6))
        self.assertIn(identity(6), group)
        for element in group:
            self.assertEqual(compose(element, identity(6)), element)
            self.assertEqual(compose(identity(6), element), element)

    def test_group_is_closed_under_composition(self) -> None:
        group = close_group(6, G.cycle_symmetries(6))
        for left in group:
            for right in group:
                self.assertIn(compose(left, right), group)

    def test_group_is_closed_under_inversion(self) -> None:
        group = close_group(6, G.cycle_symmetries(6))
        for element in group:
            self.assertIn(invert(element), group)
            self.assertEqual(compose(element, invert(element)), identity(6))

    def test_compose_applies_inner_first(self) -> None:
        inner = (1, 2, 0)
        outer = (0, 2, 1)
        # vertex 0 -> 1 (inner) -> 2 (outer)
        self.assertEqual(compose(outer, inner)[0], 2)

    def test_is_permutation(self) -> None:
        self.assertTrue(is_permutation((2, 0, 1), 3))
        self.assertFalse(is_permutation((2, 0, 1), 4))
        self.assertFalse(is_permutation((1, 1, 0), 3))


class DegenerateSizeTests(unittest.TestCase):
    def test_tiny_paths_and_cycles_have_only_the_identity(self) -> None:
        for n in (0, 1):
            with self.subTest(n=n):
                self.assertEqual(G.path_symmetries(n), ())
                self.assertEqual(G.cycle_symmetries(n), ())
                self.assertEqual(len(close_group(n, ())), 1)

    def test_two_vertex_cycle_has_only_the_swap(self) -> None:
        graph = G.make_cycle_graph(2)
        generators = G.cycle_symmetries(2)
        self.assertEqual(generators, ((1, 0),))
        self.assertTrue(is_symmetry(graph, generators[0]))
        self.assertEqual(len(close_group(2, generators)), 2)

    def test_rim_families_reject_fewer_than_three(self) -> None:
        for builder in (G.make_wheel_graph, G.make_helm_graph, G.make_sunlet_graph):
            for symmetries in (G.wheel_symmetries, G.helm_symmetries, G.sunlet_symmetries):
                for n in (-1, 0, 2):
                    with self.subTest(builder=builder.__name__, n=n):
                        with self.assertRaises(ValueError):
                            builder(n)
                        with self.assertRaises(ValueError):
                            symmetries(n)


if __name__ == "__main__":
    unittest.main()
