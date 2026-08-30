"""Shared test helpers.

Not named test_*.py, so unittest discovery ignores it.
"""

from __future__ import annotations

import os
import unittest


RUN_SLOW = os.environ.get("GAME_COLORING_SLOW") == "1"

# The everyday suite has to stay fast enough to run on every edit. A handful of
# genuinely verified values (11-vertex helms, distance-3 grids) cost a minute or
# more each, so they live behind this flag and run before results are trusted.
slow = unittest.skipUnless(RUN_SLOW, "slow; set GAME_COLORING_SLOW=1 to run")
