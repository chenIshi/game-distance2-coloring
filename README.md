# game-distance2-coloring

A small solver and experimental framework for the distance-2 coloring game, used to compute and analyze the game distance-2 chromatic number under recursive game semantics.

## Quick Start

Create a virtual environment:

```bash
python3 -m venv .venv
source .venv/bin/activate
```

Run the solver directly from the repo root:

```bash
python run_cli.py path 6
```

Check a fixed number of colors:

```bash
python run_cli.py path 6 --k 4
```

Show an explanation line:

```bash
python run_cli.py path 2 --k 1 --explain
```

## CLI Usage

Built-in graph families:

```bash
python run_cli.py path 6
python run_cli.py cycle 5
python run_cli.py star 4
```

More examples:

```bash
python run_cli.py path 6 --k 4
python run_cli.py path 4 --k 3 --explain
```

## Optional Install

You do not need to install the package to use the solver.

If you want the `game-coloring` command anyway, install it in editable mode:

```bash
pip install -r requirements.txt
```

After that, you can also use:

```bash
game-coloring path 6
python -m game_coloring path 6
```

## Example Output

```text
P_6: chi_g,2 = 4
  k=1: Bob
  k=2: Bob
  k=3: Bob
  k=4: Alice
```

For a Bob-winning case:

```text
P_2 with k=1: Bob
one failed line:
  Alice: v0 -> color 1
dead vertex set: v1
```

For an Alice-winning case, the CLI prints a winning opening move and one sample continuation. That sample line is illustrative, not a full proof of Alice's strategy tree.

## Tests

Run the test suite with:

```bash
python3 -m unittest discover -s tests
```

## Web Demo

The repo also includes a static teaching demo under `web/`.

Generate the preset case data with:

```bash
python3 scripts/export_web_cases.py
```

Then serve the repo locally with any static file server, for example:

```bash
python3 -m http.server 8000
```

After that, open `http://localhost:8000/web/`.

The intended deployment target for this demo is GitHub Pages, so the frontend is plain static HTML, CSS, JavaScript, and JSON.

## Current Scope

The current implementation includes:

- recursive game-state search
- memoization
- square-graph conversion
- path, cycle, and star graph constructors

The first target is path graphs, with extension to other small graph families after the core solver is validated.
