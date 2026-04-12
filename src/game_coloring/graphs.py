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


def make_caterpillar_graph(spine_length: int, leaf_counts: tuple[int, ...]) -> Graph:
    if spine_length < 0:
        raise ValueError("spine_length must be non-negative")
    if len(leaf_counts) != spine_length:
        raise ValueError("leaf_counts must have one entry per spine vertex")
    if any(count < 0 for count in leaf_counts):
        raise ValueError("leaf counts must be non-negative")

    total_vertices = spine_length + sum(leaf_counts)
    adjacency = [set() for _ in range(total_vertices)]

    for vertex in range(spine_length - 1):
        adjacency[vertex].add(vertex + 1)
        adjacency[vertex + 1].add(vertex)

    next_leaf = spine_length
    for spine_vertex, leaf_count in enumerate(leaf_counts):
        for _ in range(leaf_count):
            adjacency[spine_vertex].add(next_leaf)
            adjacency[next_leaf].add(spine_vertex)
            next_leaf += 1

    return normalize_graph(adjacency)


def make_binary_tree_graph(depth: int) -> Graph:
    if depth < 0:
        raise ValueError("depth must be non-negative")

    vertex_count = 2 ** (depth + 1) - 1
    adjacency = [set() for _ in range(vertex_count)]

    for parent in range(vertex_count):
        left_child = 2 * parent + 1
        right_child = 2 * parent + 2
        if left_child < vertex_count:
            adjacency[parent].add(left_child)
            adjacency[left_child].add(parent)
        if right_child < vertex_count:
            adjacency[parent].add(right_child)
            adjacency[right_child].add(parent)

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
