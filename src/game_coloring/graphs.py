from __future__ import annotations

from collections import deque


Graph = tuple[frozenset[int], ...]

# A relabelling of vertex indices: permutation[v] is the new name of vertex v.
Permutation = tuple[int, ...]


def normalize_graph(adjacency: list[set[int]]) -> Graph:
    return tuple(frozenset(neighbors) for neighbors in adjacency)


# --------------------------------------------------------------------------
# Builders.
#
# Each family's symmetries are declared immediately after its builder, because
# the builder is the only thing that knows the recipe: make_cycle_graph's
# "% n" *is* the rotation symmetry, written down. Keeping them adjacent means
# a change to the numbering is obviously a change to the symmetries too.
#
# Under-reporting symmetries costs speed only; over-reporting them corrupts
# answers. symmetry.is_symmetry exists to confirm every one of these.
# --------------------------------------------------------------------------


def make_path_graph(n: int) -> Graph:
    if n < 0:
        raise ValueError("n must be non-negative")

    adjacency = [set() for _ in range(n)]
    for vertex in range(n - 1):
        adjacency[vertex].add(vertex + 1)
        adjacency[vertex + 1].add(vertex)
    return normalize_graph(adjacency)


def path_symmetries(n: int) -> tuple[Permutation, ...]:
    """A path can only be flipped end for end."""
    if n < 0:
        raise ValueError("n must be non-negative")
    if n < 2:
        return ()
    return (tuple(n - 1 - vertex for vertex in range(n)),)


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


def cycle_symmetries(n: int) -> tuple[Permutation, ...]:
    """A cycle can be rotated one notch and flipped over."""
    if n < 0:
        raise ValueError("n must be non-negative")
    if n < 2:
        return ()
    if n == 2:
        # C_2 collapses to a single edge, whose only symmetry is the swap.
        return ((1, 0),)

    rotate = tuple((vertex + 1) % n for vertex in range(n))
    reflect = tuple((n - vertex) % n for vertex in range(n))
    return (rotate, reflect)


def make_star_graph(leaves: int) -> Graph:
    if leaves < 0:
        raise ValueError("leaves must be non-negative")

    adjacency = [set() for _ in range(leaves + 1)]
    for leaf in range(1, leaves + 1):
        adjacency[0].add(leaf)
        adjacency[leaf].add(0)
    return normalize_graph(adjacency)


def make_wheel_graph(rim: int) -> Graph:
    """W_n: a hub (vertex 0) joined to every vertex of a rim cycle 1..n."""
    if rim < 3:
        raise ValueError("a wheel needs at least 3 rim vertices")

    adjacency = [set() for _ in range(rim + 1)]
    for index in range(1, rim + 1):
        adjacency[0].add(index)
        adjacency[index].add(0)
        neighbor = index % rim + 1
        adjacency[index].add(neighbor)
        adjacency[neighbor].add(index)
    return normalize_graph(adjacency)


def wheel_symmetries(rim: int) -> tuple[Permutation, ...]:
    """The rim rotates and reflects; the hub cannot move anywhere.

    The hub is the only vertex of degree n while every rim vertex has degree 3,
    and relabelling cannot change a vertex's degree, so the hub is fixed.
    """
    if rim < 3:
        raise ValueError("a wheel needs at least 3 rim vertices")

    rotate = (0,) + tuple(index % rim + 1 for index in range(1, rim + 1))
    reflect = (0,) + tuple((rim - (index - 1)) % rim + 1 for index in range(1, rim + 1))
    return (rotate, reflect)


def make_helm_graph(rim: int) -> Graph:
    """H_n: a wheel with one pendant hung off each rim vertex.

    Hub 0, rim 1..n, and the pendant of rim vertex i sits at n + i.
    """
    if rim < 3:
        raise ValueError("a helm needs at least 3 rim vertices")

    wheel = make_wheel_graph(rim)
    adjacency = [set(neighbors) for neighbors in wheel] + [set() for _ in range(rim)]
    for index in range(1, rim + 1):
        pendant = rim + index
        adjacency[index].add(pendant)
        adjacency[pendant].add(index)
    return normalize_graph(adjacency)


def helm_symmetries(rim: int) -> tuple[Permutation, ...]:
    """The wheel's symmetries, with each pendant following its rim vertex."""
    if rim < 3:
        raise ValueError("a helm needs at least 3 rim vertices")

    def carry_pendants(permutation: Permutation) -> Permutation:
        return permutation + tuple(rim + permutation[index] for index in range(1, rim + 1))

    return tuple(carry_pendants(permutation) for permutation in wheel_symmetries(rim))


def make_sunlet_graph(rim: int) -> Graph:
    """S_n (also written C_n o K_1): a cycle with one pendant per vertex.

    Rim 0..n-1, and the pendant of rim vertex i sits at n + i.
    """
    if rim < 3:
        raise ValueError("a sunlet needs at least 3 rim vertices")

    cycle = make_cycle_graph(rim)
    adjacency = [set(neighbors) for neighbors in cycle] + [set() for _ in range(rim)]
    for index in range(rim):
        pendant = rim + index
        adjacency[index].add(pendant)
        adjacency[pendant].add(index)
    return normalize_graph(adjacency)


def sunlet_symmetries(rim: int) -> tuple[Permutation, ...]:
    """The cycle's symmetries, with each pendant following its rim vertex."""
    if rim < 3:
        raise ValueError("a sunlet needs at least 3 rim vertices")

    def carry_pendants(permutation: Permutation) -> Permutation:
        return permutation + tuple(rim + permutation[index] for index in range(rim))

    return tuple(carry_pendants(permutation) for permutation in cycle_symmetries(rim))


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


# --------------------------------------------------------------------------
# Derived structure.
# --------------------------------------------------------------------------


def power_graph(graph: Graph, distance: int = 2) -> Graph:
    """Join every pair of vertices lying within `distance` of each other.

    Solving distance-d colouring on `graph` is ordinary colouring on this, so
    the solver never has to reason about distance itself.
    """
    if distance < 0:
        raise ValueError("distance must be non-negative")

    adjacency: list[set[int]] = []
    for source in range(len(graph)):
        reached = {source}
        frontier = {source}
        for _ in range(distance):
            following: set[int] = set()
            for vertex in frontier:
                following |= set(graph[vertex])
            following -= reached
            if not following:
                break
            reached |= following
            frontier = following
        reached.discard(source)
        adjacency.append(reached)

    return normalize_graph(adjacency)


def build_square_graph(graph: Graph) -> Graph:
    """The distance-2 power graph. Kept as the original name."""
    return power_graph(graph, 2)


def is_complete(graph: Graph) -> bool:
    """True when every vertex is joined to every other one."""
    return all(len(neighbors) == len(graph) - 1 for neighbors in graph)


def diameter(graph: Graph) -> float:
    """Largest distance between any two vertices; inf if disconnected."""
    if not graph:
        return 0

    best = 0
    for source in range(len(graph)):
        distances = {source: 0}
        queue = deque([source])
        while queue:
            vertex = queue.popleft()
            for neighbor in graph[vertex]:
                if neighbor not in distances:
                    distances[neighbor] = distances[vertex] + 1
                    queue.append(neighbor)
        if len(distances) != len(graph):
            return float("inf")
        best = max(best, max(distances.values()))
    return best
