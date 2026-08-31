// Vertex positions for the SVG board.
//
// Computed here rather than baked into a data file, because the sandbox lets
// people build any size and there is no list of sizes to precompute. The
// viewBox is 520x340, fixed in index.html.
//
// Angles start at the top and go clockwise so that the drawing matches the
// numbering contract: rim vertex 1 (or 0, for a sunlet) sits at 12 o'clock and
// the numbers increase the way you would read a clock face.

const WIDTH = 520;
const HEIGHT = 340;
const CX = WIDTH / 2;
const CY = HEIGHT / 2;

function ring(count, radius, cx = CX, cy = CY) {
  const points = [];
  for (let i = 0; i < count; i += 1) {
    const angle = (-0.5 + (2 * i) / count) * Math.PI;
    points.push({
      x: round(cx + radius * Math.cos(angle)),
      y: round(cy + radius * Math.sin(angle)),
    });
  }
  return points;
}

function round(value) {
  return Math.round(value * 100) / 100;
}

// Rings get tighter as n grows so labels keep room to breathe.
function rimRadius(n, base) {
  if (n <= 6) return base;
  if (n <= 10) return base + 8;
  return base + 14;
}

export function positionsFor(family, n) {
  switch (family) {
    case "path": {
      if (n <= 0) return [];
      const y = CY;
      if (n === 1) return [{ x: CX, y }];
      const left = 60;
      const right = WIDTH - 60;
      const step = (right - left) / (n - 1);
      return Array.from({ length: n }, (_, i) => ({ x: round(left + step * i), y }));
    }

    case "cycle":
      return ring(n, rimRadius(n, 118));

    case "star": {
      // Centre is vertex 0, leaves are 1..n.
      return [{ x: CX, y: CY }, ...ring(n, rimRadius(n, 122))];
    }

    case "wheel": {
      // Hub is vertex 0, rim is 1..n.
      return [{ x: CX, y: CY }, ...ring(n, rimRadius(n, 120))];
    }

    case "helm": {
      // Hub 0, rim 1..n, pendant of rim i at n+i, drawn straight outward.
      const rim = ring(n, rimRadius(n, 92));
      const pendants = ring(n, rimRadius(n, 92) + 55);
      return [{ x: CX, y: CY }, ...rim, ...pendants];
    }

    case "sunlet": {
      // Rim 0..n-1, pendant of rim i at n+i, drawn straight outward.
      const rim = ring(n, rimRadius(n, 96));
      const pendants = ring(n, rimRadius(n, 96) + 55);
      return [...rim, ...pendants];
    }

    default:
      return [];
  }
}

// vertexLabel names a vertex by its role, the way the CLI's -explain does.
// "rim4 blocked by hub and its own stub" is readable; "v4 blocked by v0, v8"
// makes the reader decode indices.
export function vertexLabel(family, n, vertex) {
  switch (family) {
    case "star":
      return vertex === 0 ? "centre" : `leaf${vertex}`;
    case "wheel":
      return vertex === 0 ? "hub" : `rim${vertex}`;
    case "helm":
      if (vertex === 0) return "hub";
      return vertex <= n ? `rim${vertex}` : `stub${vertex - n}`;
    case "sunlet":
      return vertex < n ? `rim${vertex}` : `stub${vertex - n}`;
    default:
      return `v${vertex}`;
  }
}

export const FAMILIES = [
  { id: "path", label: "Path", sizeLabel: "vertices", min: 2, max: 14 },
  { id: "cycle", label: "Cycle", sizeLabel: "vertices", min: 3, max: 14 },
  { id: "star", label: "Star", sizeLabel: "leaves", min: 2, max: 12 },
  { id: "wheel", label: "Wheel", sizeLabel: "rim vertices", min: 3, max: 12 },
  { id: "helm", label: "Helm", sizeLabel: "rim vertices", min: 3, max: 7 },
  { id: "sunlet", label: "Sunlet", sizeLabel: "rim vertices", min: 3, max: 7 },
];

// orderOf mirrors the builders: the size parameter is not the vertex count for
// every family, and the page needs the real count to size a colour array.
export function orderOf(family, n) {
  switch (family) {
    case "path":
    case "cycle":
      return n;
    case "star":
    case "wheel":
      return n + 1;
    case "helm":
      return 2 * n + 1;
    case "sunlet":
      return 2 * n;
    default:
      return 0;
  }
}
