# Draft PR: Speed up computed register expansion

## Title

`perf: parallelize non-recursive computed register expansion`

## Summary

This PR speeds up trace expansion for computed registers by introducing row-level parallelism for non-recursive computations, compiled PolyFil evaluation, and direct field element writes. Based on PR #1631, rebased onto main with all review comments addressed.

**Key optimizations:**

- **Parallel row evaluation** -- For non-recursive computed registers, each row's evaluation is independent. Rows are split into chunks (one per available CPU core) and processed in parallel goroutines. Falls back to sequential for small heights (< 4096 rows).

- **Compiled PolyFil fast path** -- Pre-compile `PolyFil` expressions once (resolve column references upfront), then evaluate the cached representation per row, avoiding repeated map lookups.

- **Direct field writes** -- Write computed results directly to field elements via `fieldWordSplitter` instead of going through the `word.BigEndian -> SetBytes -> field` round-trip used by `concretizeColumns`.

- **Uint64 fast path in `columnAdapter.Get`** -- Avoid expensive `SetBytes` for small values (common case for most trace columns) by checking if the value fits in a uint64.

## Changes

| File | Change |
|------|--------|
| `pkg/ir/assignment/computed_register.go` | Parallel dispatch for non-recursive computations, `compiledPolyFil` cached evaluator, `fieldWordSplitter` for direct field writes, configurable inner worker count |
| `pkg/schema/agnostic/filler.go` | Export `Polynomial()` and `RightShift()` on `PolyFil` for compiled evaluation |
| `pkg/trace/util.go` | `Uint64` fast path in `columnAdapter.Get` |
| `pkg/cmd/corset/root.go` | New `--workers` CLI flag for controlling inner parallelism |

## Review comments addressed (vs PR #1631)

### 1. Removed dead code (`fwdComputationParallel`)
The original PR contained an unused `fwdComputationParallel` function that operated on `[][]word.BigEndian`. This was an intermediate implementation superseded by `fwdComputationParallelDirect` and `fwdCompiledPolyFilParallel`. It has been removed entirely.
*(Flagged by: Cursor Bugbot, GitHub Copilot)*

### 2. Fixed `from.SetUint64(u)` mutation in `columnAdapter.Get`
The original code used `from.Equals(from.SetUint64(u))` which could mutate `from` depending on the field element implementation. Now uses a fresh temporary variable `check` for the comparison:
```go
var check F1
if from.Equals(check.SetUint64(u)) {
```
*(Flagged by: GitHub Copilot)*

### 3. Configurable inner parallelism instead of hardcoded `GOMAXPROCS`
The original PR used `runtime.GOMAXPROCS(0)` directly and hardcoded `1024` as the minimum rows per worker. This conflicts with the outer parallelism in `ParallelTraceExpansion`.

Now:
- **`INNER_WORKERS`** -- exported package-level variable (same pattern as `mir.EXPLODING_MULTIPLIER`) that defaults to 0 (auto-detect from `GOMAXPROCS`), configurable via the new `--workers` CLI flag.
- **`minRowsPerWorker`** -- named constant (1024) replacing the magic number.
- **`minParallelHeight`** -- named constant (4096) for the sequential fallback threshold.
- **`innerWorkers(height)`** -- helper that computes the effective worker count, capping based on height to avoid goroutine overhead.

Users can now control inner parallelism via `--workers N` to coordinate with the outer batch-level parallelism controlled by `--batch`.
*(Flagged by: DavePearce, repo maintainer)*

## Performance

Original PR benchmark (r8a.24xlarge, 96c AMD EPYC, Linea mainnet 22 txs / 7.3M gas):

| Configuration | Trace Expansion Time |
|---|---|
| go-corset v1.2.7 (upstream) | 2m 51s |
| This branch | 24.6s |
| **Speedup** | **6.9x** |

Local benchmark (Test_AsmBench_Bin, 3 iterations, Apple Silicon):

| Configuration | Wall time | User CPU |
|---|---|---|
| main | 84.6s | 287s |
| This branch | 84.5s | 267s |

Small test traces (< 4096 rows) don't trigger the parallel path, so no speedup is expected. The real gains appear on production-sized traces with millions of rows. No regression observed.

## Test plan

- [x] `go build ./...` passes
- [x] `go vet` passes on all modified packages
- [x] `Test_Valid_Comp`, `Test_Valid_ByteDecomp`, `Test_Valid_Word`, `Test_Valid_Shift` pass
- [x] `Test_Invalid_Comp`, `Test_Invalid_ByteDecomp`, `Test_Invalid_Word` pass
- [x] `Test_AsmUnit`, `Test_AsmInvalid` pass
- [x] `Test_Valid_Basic`, `Test_Valid_Interleave` pass
- [x] Race detector passes on computed register tests
- [x] No regression on benchmark tests
- [ ] Trace expansion produces identical outputs on large production traces (to be verified by maintainer)
- [ ] Recursive computed registers still use sequential path (by design)
