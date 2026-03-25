# ZkC Architecture: Arithmetization Layer

## Compilation

The arithmetization pass (`pkg/asm/compiler/`) converts each function's
vectorized instruction sequence into a MIR module of vanishing constraints,
range constraints, and lookup arguments.

### Modules and columns

Each function becomes one MIR module.  The module contains one column per
register (inputs, outputs, and locals), plus — for functions with more than
one vector bundle — a **program counter** (`PC`) column and a **return line**
(`RET`) column.  Each row of the module represents one execution step (one
vector bundle) of one invocation of the function.  Range constraints are
added automatically for every column to enforce its declared bit-width.

Functions that vectorize to a single bundle are called _atomic_.  They need
no `PC` or `RET` columns because every row is unconditionally one complete
invocation.

### One vanishing constraint per bundle

For each vector bundle at program counter `k`, the compiler emits a single
vanishing constraint guarded by `PC[i] == k` (i.e. the constraint need only
hold on rows where the current step is `k`).  The body of the constraint is
the conjunction of the constraints generated for each micro-instruction in
the bundle:

- **`Add` / `Mul`** — an equality `lhs = rhs` where the left-hand side is the
  target register on the current row, and the right-hand side is the
  polynomial over the source registers.  When a source register was already
  written earlier in the same bundle (forwarding), its current-row value
  `reg[i]` is used; otherwise the previous-row value `reg[i-1]` is used.
  For subtraction (`b, x := y - c`), the sign bit `b` is moved to the
  right-hand side and the constraint is rebalanced:
  `x + c = y + 2^N * b` (where `N` is the bit-width of `x`).

- **`SkipIf` / `Skip`** — these do not emit a constraint themselves, but they
  build a _branch table_ that records the path condition under which each
  subsequent micro-instruction executes.  Every constraint generated from a
  later micro-instruction in the bundle is wrapped in that condition, so only
  the constraints on the active path are enforced.

- **`Jmp`** — emits `PC[i+1] = target`, constraining the program counter
  transition.

- **`Return`** — emits `RET[i] = 1`, marking the row as a return step.  The
  return line is used as the enable signal for the lookup argument that
  enforces function-call correctness (see below).

- **`Fail`** — emits the constant `false`, making the constraint
  unsatisfiable on any row that reaches this instruction.

### Constancy constraints

Registers that are not written by a bundle must retain their value from the
previous row.  After building the per-micro-instruction constraints, the
compiler adds a **constancy constraint** `reg[i] = reg[i-1]` for every
register that the bundle does not definitely write.  For registers that are
only _sometimes_ written (e.g. on one branch of a `SkipIf`), the write
condition is derived from the branch table, negated, and used as a guard:
`(not written) => reg[i] = reg[i-1]`.

### Function calls as lookup arguments

Each call site (bus) in a function generates a **lookup argument**: the tuple
of argument and return columns at the call site must appear as a row in the
callee's module.  For multi-line callees the lookup is filtered to rows where
`RET == 1`, ensuring that only complete, returning invocations are looked up.

## Branch Table Optimisation

As described above, each micro-instruction within a vector bundle is guarded
by a _branch condition_ — a logical formula over register equalities and
inequalities that records the exact path through the bundle that must have
been taken for that instruction to execute.  These conditions can accumulate
redundant atoms as the path through a bundle grows, and translating them
naively into polynomial constraint terms produces unnecessarily large
constraints.

Branch table optimisation (`pkg/util/logical/`) simplifies these conditions
before they are translated.  Conditions are maintained in Disjunctive Normal
Form (DNF) and simplified using a set of rules including:

- **Subsumption** — `x==1 ∧ x!=0` simplifies to `x==1`, since `x==1`
  already implies `x!=0`.
- **Contradiction** — `x==1 ∧ x==2` simplifies to `⊥` (false), and any
  conjunction containing `⊥` is dropped.
- **Tautology** — `x==0 ∨ x!=0` simplifies to `⊤` (true), eliminating the
  guard entirely.
- **Unit propagation** — `x==0 ∧ y==x` simplifies to `x==0 ∧ y==0` by
  substituting the known value of `x` into subsequent atoms.

Simpler conditions translate into fewer and cheaper polynomial terms in the
generated constraints, directly reducing proof cost.

## Other optimisations (e.g. binary conditions, conjunctions)
