from __future__ import annotations


Graph = tuple[frozenset[int], ...]


def normalize_graph(adjacency: list[set[int]]) -> Graph:
    return tuple(frozenset(neighbors) for neighbors in adjacency)


def make_path_graph(n: int) -> Graph:
    if n < 0:
        raise ValueError("n must be non-negative")

    adjacency = [set() for _ in range(n)]
    for vertex in range(n - 1):
        adjacency[vertex].add(vertex + 1)
        adjacency[vertex + 1].add(vertex)
    return normalize_graph(adjacency)


def make_cycle_graph(n: int) -> Graph:
    if n < 0:
        raise ValueError("n must be non-negative")
    if n == 0:
        return ()
    if n == 1:
        return (frozenset(),)

    adjacency = [set() for _ in range(n)]
    for vertex in range(n):
        neighbor = (vertex + 1) % n
        adjacency[vertex].add(neighbor)
        adjacency[neighbor].add(vertex)
    return normalize_graph(adjacency)


def make_star_graph(leaves: int) -> Graph:
    if leaves < 0:
        raise ValueError("leaves must be non-negative")

    adjacency = [set() for _ in range(leaves + 1)]
    for leaf in range(1, leaves + 1):
        adjacency[0].add(leaf)
        adjacency[leaf].add(0)
    return normalize_graph(adjacency)


def build_square_graph(graph: Graph) -> Graph:
    square_adjacency = [set() for _ in range(len(graph))]

    for vertex, neighbors in enumerate(graph):
        distance_two_neighbors = set(neighbors)
        for neighbor in neighbors:
            distance_two_neighbors.update(graph[neighbor])

        distance_two_neighbors.discard(vertex)
        square_adjacency[vertex] = distance_two_neighbors

    return normalize_graph(square_adjacency)
