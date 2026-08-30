"""Graph symmetries: relabellings of the vertices that leave the edges alone.

A symmetry is a permutation `p` where `p[v]` is the new name of vertex `v`, and
every edge is still an edge afterwards. Two positions of the colouring game that
differ by a symmetry are the same position, so recognising them collapses work.

The safety property that shapes this module: missing a symmetry only costs
speed, while claiming a false one silently corrupts answers. So builders declare
the obvious rotations and reflections and `is_symmetry` confirms them, rather
than anyone trying to prove a list is complete.
"""

from __future__ import annotations

from collections.abc import Iterable

from .graphs import Graph, Permutation


def identity(size: int) -> Permutation:
    return tuple(range(size))


def compose(outer: Permutation, inner: Permutation) -> Permutation:
    """Apply `inner` first, then `outer`."""
    return tuple(outer[inner[vertex]] for vertex in range(len(inner)))


def invert(permutation: Permutation) -> Permutation:
    inverse = [0] * len(permutation)
    for vertex, image in enumerate(permutation):
        inverse[image] = vertex
    return tuple(inverse)


def is_permutation(candidate: Permutation, size: int) -> bool:
    return len(candidate) == size and sorted(candidate) == list(range(size))


def is_symmetry(graph: Graph, candidate: Permutation) -> bool:
    """True when `candidate` renames vertices without disturbing any edge."""
    if not is_permutation(candidate, len(graph)):
        return False

    for vertex, neighbors in enumerate(graph):
        for neighbor in neighbors:
            if candidate[neighbor] not in graph[candidate[vertex]]:
                return False
    return True


def close_group(size: int, generators: Iterable[Permutation]) -> tuple[Permutation, ...]:
    """Every relabelling reachable by combining `generators` any number of times.

    The generators are a small starter set; this expands them into the whole
    collection, the way six face turns generate every position of a cube.
    """
    generators = tuple(generators)
    start = identity(size)
    found = {start}
    frontier = [start]

    while frontier:
        following: list[Permutation] = []
        for element in frontier:
            for generator in generators:
                combined = compose(generator, element)
                if combined not in found:
                    found.add(combined)
                    following.append(combined)
        frontier = following

    return tuple(sorted(found))
