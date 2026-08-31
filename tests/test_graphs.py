from __future__ import annotations

import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1] / "src"))

from game_coloring.graphs import (
    build_square_graph,
    diameter,
    is_complete,
    make_helm_graph,
    make_sunlet_graph,
    make_wheel_graph,
    power_graph,
    make_binary_tree_graph,
    make_caterpillar_graph,
    make_cycle_graph,
    make_path_graph,
    make_star_graph,
    normalize_graph,
)


def edge_count(graph) -> int:
    return sum(len(neighbors) for neighbors in graph) // 2


def complete_graph(n: int):
    return normalize_graph([{u for u in range(n) if u != v} for v in range(n)])


def sample_graphs():
    for n in range(0, 7):
        yield f"P_{n}", make_path_graph(n)
        yield f"C_{n}", make_cycle_graph(n)
        yield f"K_1,{n}", make_star_graph(n)
    yield "caterpillar(3,(1,1,1))", make_caterpillar_graph(3, (1, 1, 1))
    yield "caterpillar(4,(2,0,1,0))", make_caterpillar_graph(4, (2, 0, 1, 0))
    for depth in range(0, 4):
        yield f"binary_tree({depth})", make_binary_tree_graph(depth)


class GraphInvariantTests(unittest.TestCase):
    """Properties every graph produced by this module must satisfy."""

    def test_adjacency_is_symmetric(self) -> None:
        for label, graph in sample_graphs():
            with self.subTest(graph=label):
                for vertex, neighbors in enumerate(graph):
                    for neighbor in neighbors:
                        self.assertIn(vertex, graph[neighbor])

    def test_no_self_loops(self) -> None:
        for label, graph in sample_graphs():
            with self.subTest(graph=label):
                for vertex, neighbors in enumerate(graph):
                    self.assertNotIn(vertex, neighbors)

    def test_neighbor_indices_are_in_range(self) -> None:
        for label, graph in sample_graphs():
            with self.subTest(graph=label):
                for neighbors in graph:
                    for neighbor in neighbors:
                        self.assertGreaterEqual(neighbor, 0)
                        self.assertLess(neighbor, len(graph))

    def test_squares_preserve_the_same_invariants(self) -> None:
        for label, graph in sample_graphs():
            with self.subTest(graph=label):
                square = build_square_graph(graph)
                self.assertEqual(len(square), len(graph))
                for vertex, neighbors in enumerate(square):
                    self.assertNotIn(vertex, neighbors)
                    for neighbor in neighbors:
                        self.assertIn(vertex, square[neighbor])


class PathGraphStructureTests(unittest.TestCase):
    def test_sizes(self) -> None:
        for n in range(0, 8):
            with self.subTest(n=n):
                graph = make_path_graph(n)
                self.assertEqual(len(graph), n)
                self.assertEqual(edge_count(graph), max(n - 1, 0))

    def test_endpoints_have_degree_one(self) -> None:
        graph = make_path_graph(5)
        self.assertEqual(sorted(len(neighbors) for neighbors in graph), [1, 1, 2, 2, 2])
        self.assertEqual(graph[0], frozenset({1}))
        self.assertEqual(graph[4], frozenset({3}))

    def test_rejects_negative(self) -> None:
        with self.assertRaises(ValueError):
            make_path_graph(-1)


class CycleGraphStructureTests(unittest.TestCase):
    def test_sizes(self) -> None:
        for n in range(3, 9):
            with self.subTest(n=n):
                graph = make_cycle_graph(n)
                self.assertEqual(len(graph), n)
                self.assertEqual(edge_count(graph), n)
                self.assertTrue(all(len(neighbors) == 2 for neighbors in graph))

    def test_degenerate_sizes(self) -> None:
        self.assertEqual(make_cycle_graph(0), ())
        self.assertEqual(make_cycle_graph(1), (frozenset(),))
        # C_2 collapses the two parallel edges into a single edge.
        two = make_cycle_graph(2)
        self.assertEqual(len(two), 2)
        self.assertEqual(edge_count(two), 1)

    def test_rejects_negative(self) -> None:
        with self.assertRaises(ValueError):
            make_cycle_graph(-1)


class StarGraphStructureTests(unittest.TestCase):
    def test_sizes_and_degrees(self) -> None:
        for leaves in range(0, 6):
            with self.subTest(leaves=leaves):
                graph = make_star_graph(leaves)
                self.assertEqual(len(graph), leaves + 1)
                self.assertEqual(edge_count(graph), leaves)
                self.assertEqual(len(graph[0]), leaves)
                for leaf in range(1, leaves + 1):
                    self.assertEqual(graph[leaf], frozenset({0}))

    def test_rejects_negative(self) -> None:
        with self.assertRaises(ValueError):
            make_star_graph(-1)


class CaterpillarStructureTests(unittest.TestCase):
    def test_sizes(self) -> None:
        graph = make_caterpillar_graph(3, (1, 1, 1))
        self.assertEqual(len(graph), 6)
        self.assertEqual(edge_count(graph), 5)

    def test_leaves_attach_to_their_spine_vertex(self) -> None:
        graph = make_caterpillar_graph(3, (1, 0, 2))
        self.assertEqual(len(graph), 6)
        self.assertEqual(graph[3], frozenset({0}))
        self.assertEqual(graph[4], frozenset({2}))
        self.assertEqual(graph[5], frozenset({2}))

    def test_rejects_bad_arguments(self) -> None:
        with self.assertRaises(ValueError):
            make_caterpillar_graph(-1, ())
        with self.assertRaises(ValueError):
            make_caterpillar_graph(3, (1, 1))
        with self.assertRaises(ValueError):
            make_caterpillar_graph(2, (1, -1))


class BinaryTreeStructureTests(unittest.TestCase):
    def test_sizes(self) -> None:
        for depth in range(0, 5):
            with self.subTest(depth=depth):
                graph = make_binary_tree_graph(depth)
                self.assertEqual(len(graph), 2 ** (depth + 1) - 1)
                self.assertEqual(edge_count(graph), 2 ** (depth + 1) - 2)

    def test_children_indices(self) -> None:
        graph = make_binary_tree_graph(2)
        self.assertEqual(graph[0], frozenset({1, 2}))
        self.assertEqual(graph[1], frozenset({0, 3, 4}))
        self.assertEqual(graph[2], frozenset({0, 5, 6}))

    def test_rejects_negative(self) -> None:
        with self.assertRaises(ValueError):
            make_binary_tree_graph(-1)


class SquareGraphTests(unittest.TestCase):
    def test_path_square_is_exact(self) -> None:
        square = build_square_graph(make_path_graph(4))
        self.assertEqual(square[0], frozenset({1, 2}))
        self.assertEqual(square[1], frozenset({0, 2, 3}))
        self.assertEqual(square[2], frozenset({0, 1, 3}))
        self.assertEqual(square[3], frozenset({1, 2}))

    def test_square_contains_every_original_edge(self) -> None:
        for label, graph in sample_graphs():
            with self.subTest(graph=label):
                square = build_square_graph(graph)
                for vertex, neighbors in enumerate(graph):
                    self.assertTrue(neighbors <= square[vertex])

    def test_diameter_two_graphs_square_to_complete(self) -> None:
        # A star and a short cycle both have diameter <= 2, so every pair of
        # vertices lands within distance 2 and the square is complete.
        for label, graph in [
            ("K_1,4", make_star_graph(4)),
            ("C_3", make_cycle_graph(3)),
            ("C_4", make_cycle_graph(4)),
            ("C_5", make_cycle_graph(5)),
        ]:
            with self.subTest(graph=label):
                self.assertEqual(build_square_graph(graph), complete_graph(len(graph)))

    def test_complete_graph_is_its_own_square(self) -> None:
        for n in range(1, 6):
            with self.subTest(n=n):
                graph = complete_graph(n)
                self.assertEqual(build_square_graph(graph), graph)

    def test_empty_graph_squares_to_empty(self) -> None:
        self.assertEqual(build_square_graph(()), ())


class WheelStructureTests(unittest.TestCase):
    def test_sizes_and_degrees(self) -> None:
        for rim in range(3, 8):
            with self.subTest(rim=rim):
                graph = make_wheel_graph(rim)
                self.assertEqual(len(graph), rim + 1)
                self.assertEqual(edge_count(graph), 2 * rim)
                self.assertEqual(len(graph[0]), rim, "hub joins every rim vertex")
                for index in range(1, rim + 1):
                    self.assertEqual(len(graph[index]), 3, "rim: hub plus two neighbours")
                    self.assertIn(0, graph[index])

    def test_rim_is_a_cycle(self) -> None:
        graph = make_wheel_graph(5)
        for index in range(1, 6):
            self.assertIn(index % 5 + 1, graph[index])

    def test_rejects_fewer_than_three_rim_vertices(self) -> None:
        for rim in (-1, 0, 1, 2):
            with self.subTest(rim=rim), self.assertRaises(ValueError):
                make_wheel_graph(rim)


class HelmStructureTests(unittest.TestCase):
    def test_sizes_and_degrees(self) -> None:
        for rim in range(3, 7):
            with self.subTest(rim=rim):
                graph = make_helm_graph(rim)
                self.assertEqual(len(graph), 2 * rim + 1)
                self.assertEqual(edge_count(graph), 3 * rim)
                self.assertEqual(len(graph[0]), rim)
                for index in range(1, rim + 1):
                    self.assertEqual(len(graph[index]), 4, "hub, two rim, one pendant")
                for pendant in range(rim + 1, 2 * rim + 1):
                    self.assertEqual(len(graph[pendant]), 1)

    def test_pendant_indices_follow_the_numbering_contract(self) -> None:
        rim = 4
        graph = make_helm_graph(rim)
        for index in range(1, rim + 1):
            self.assertEqual(graph[rim + index], frozenset({index}))

    def test_rejects_fewer_than_three_rim_vertices(self) -> None:
        for rim in (-1, 0, 2):
            with self.subTest(rim=rim), self.assertRaises(ValueError):
                make_helm_graph(rim)


class SunletStructureTests(unittest.TestCase):
    def test_sizes_and_degrees(self) -> None:
        for rim in range(3, 7):
            with self.subTest(rim=rim):
                graph = make_sunlet_graph(rim)
                self.assertEqual(len(graph), 2 * rim)
                self.assertEqual(edge_count(graph), 2 * rim)
                for index in range(rim):
                    self.assertEqual(len(graph[index]), 3, "two rim, one pendant")
                for pendant in range(rim, 2 * rim):
                    self.assertEqual(len(graph[pendant]), 1)

    def test_pendant_indices_follow_the_numbering_contract(self) -> None:
        rim = 5
        graph = make_sunlet_graph(rim)
        for index in range(rim):
            self.assertEqual(graph[rim + index], frozenset({index}))

    def test_rejects_fewer_than_three_rim_vertices(self) -> None:
        for rim in (-1, 0, 2):
            with self.subTest(rim=rim), self.assertRaises(ValueError):
                make_sunlet_graph(rim)


class PowerGraphTests(unittest.TestCase):
    def test_distance_one_is_the_graph_itself(self) -> None:
        for label, graph in sample_graphs():
            with self.subTest(graph=label):
                self.assertEqual(power_graph(graph, 1), graph)

    def test_distance_two_matches_the_original_square(self) -> None:
        for label, graph in sample_graphs():
            with self.subTest(graph=label):
                self.assertEqual(power_graph(graph, 2), build_square_graph(graph))

    def test_distance_zero_has_no_edges(self) -> None:
        graph = power_graph(make_path_graph(5), 0)
        self.assertEqual(len(graph), 5)
        self.assertEqual(edge_count(graph), 0)

    def test_reach_only_grows_with_distance(self) -> None:
        for label, graph in sample_graphs():
            with self.subTest(graph=label):
                for distance in range(0, 5):
                    smaller = power_graph(graph, distance)
                    larger = power_graph(graph, distance + 1)
                    for vertex in range(len(graph)):
                        self.assertTrue(smaller[vertex] <= larger[vertex])

    def test_exact_on_a_path(self) -> None:
        cubed = power_graph(make_path_graph(5), 3)
        self.assertEqual(cubed[0], frozenset({1, 2, 3}))
        self.assertEqual(cubed[1], frozenset({0, 2, 3, 4}))
        self.assertEqual(cubed[2], frozenset({0, 1, 3, 4}))
        self.assertEqual(cubed[3], frozenset({0, 1, 2, 4}))
        self.assertEqual(cubed[4], frozenset({1, 2, 3}))

    def test_reaching_the_diameter_completes_the_graph(self) -> None:
        graphs = [
            ("P_5", make_path_graph(5)),
            ("C_7", make_cycle_graph(7)),
            ("W_5", make_wheel_graph(5)),
            ("H_4", make_helm_graph(4)),
            ("S_4", make_sunlet_graph(4)),
        ]
        for label, graph in graphs:
            with self.subTest(graph=label):
                span = diameter(graph)
                self.assertFalse(is_complete(power_graph(graph, int(span) - 1)))
                self.assertTrue(is_complete(power_graph(graph, int(span))))

    def test_rejects_negative_distance(self) -> None:
        with self.assertRaises(ValueError):
            power_graph(make_path_graph(3), -1)

    def test_empty_graph(self) -> None:
        for distance in range(0, 4):
            self.assertEqual(power_graph((), distance), ())


class CompletenessAndDiameterTests(unittest.TestCase):
    def test_is_complete(self) -> None:
        self.assertTrue(is_complete(complete_graph(4)))
        self.assertTrue(is_complete(complete_graph(1)))
        self.assertTrue(is_complete(()), "the empty graph is vacuously complete")
        self.assertFalse(is_complete(make_path_graph(3)))
        self.assertFalse(is_complete(make_cycle_graph(5)))

    def test_known_diameters(self) -> None:
        expected = [
            ("P_5", make_path_graph(5), 4),
            ("P_7", make_path_graph(7), 6),
            ("C_5", make_cycle_graph(5), 2),
            ("C_6", make_cycle_graph(6), 3),
            ("C_7", make_cycle_graph(7), 3),
            ("K_1,4", make_star_graph(4), 2),
            ("W_3", make_wheel_graph(3), 1),
            ("W_6", make_wheel_graph(6), 2),
            ("H_3", make_helm_graph(3), 3),
            ("H_4", make_helm_graph(4), 4),
            ("H_5", make_helm_graph(5), 4),
            ("S_3", make_sunlet_graph(3), 3),
            ("S_4", make_sunlet_graph(4), 4),
            ("S_5", make_sunlet_graph(5), 4),
        ]
        for label, graph, span in expected:
            with self.subTest(graph=label):
                self.assertEqual(diameter(graph), span)

    def test_wheels_and_stars_have_diameter_two(self) -> None:
        # This is why both families are trivial at every distance >= 2.
        for rim in range(4, 9):
            with self.subTest(rim=rim):
                self.assertEqual(diameter(make_wheel_graph(rim)), 2)
                self.assertEqual(diameter(make_star_graph(rim)), 2)

    def test_disconnected_graph_has_infinite_diameter(self) -> None:
        disconnected = normalize_graph([{1}, {0}, set()])
        self.assertEqual(diameter(disconnected), float("inf"))

    def test_empty_graph_diameter(self) -> None:
        self.assertEqual(diameter(()), 0)


if __name__ == "__main__":
    unittest.main()
