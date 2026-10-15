#!/usr/bin/env bash
# Count AIR lookups named range_u1_* .. range_u16_*.
# Usage: ./count_range_lookups.sh [file.zkc]   (default: lineth riscv main.zkc)
set -euo pipefail

SRC="${1:-../lineth-monorepo/arithmetization/src/main/riscv/main.zkc}"

./bin/zkc compile --air "$SRC" \
  | grep -oE '\(lookup "range_u[0-9]+_' \
  | awk '
      { sub(/.*range_u/, ""); sub(/_$/, ""); count[$0]++ }
      END {
        printf "%-10s %7s\n", "table", "lookups"
        printf "%-10s %7s\n", "----------", "-------"
        for (w = 1; w <= 16; w++) {
          printf "%-10s %7d\n", "range_u" w, count[w]
          total += count[w]
        }
        printf "%-10s %7s\n", "----------", "-------"
        printf "%-10s %7d\n", "total", total
      }'
