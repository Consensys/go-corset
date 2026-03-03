# ZkC Architecture: Intermediate Representation

The front end's AST is designed around the source language — it preserves
structure that is useful for type-checking and linking but is awkward to
translate directly into polynomial constraints.  The Intermediate
Representation (IR) bridges this gap by recasting the program as a simple
**register machine** whose semantics map cleanly onto the row-per-step
structure of an arithmetized trace.

The IR describes a machine with the following components:

- **Registers** — typed storage slots, each with a declared bit-width.
  Within a single function frame the registers are identified by integer
  indices rather than names.  Each function therefore has a fixed-size
  register file containing its parameters, return values, and temporaries.

- **Machine words** (`pkg/zkc/vm/word/`) — an abstraction over the concrete
  numeric type used to hold register values.  During simulation an unbounded
  `Uint` (arbitrary-precision integer) is used so that overflow is detected
  rather than silently truncated.  When targeting a specific prime field a
  field-element type replaces `Uint`, and any register whose bit-width exceeds
  the field's bandwidth must first be split across multiple narrower registers
  (see _Register splitting_ below).

- **Instructions** (`pkg/zkc/vm/instruction/`) — a small, flat instruction
  set that replaces the richer AST statement nodes:

  - `Add` and `Mul` — arithmetic operations with one or more sources and a
    constant; the result is distributed big-endian across one or more target
    registers, which is how carry bits are captured naturally.
  - `SkipIf` / `Skip` — conditional and unconditional forward jumps measured
    in instruction slots; these replace the `IfGoto`/`Goto` from the AST.
  - `Return` and `Fail` — terminate the current frame normally or with an
    error.
  - `Vector` — a VLIW-style bundle of the above micro-instructions that are
    executed "in parallel" within a single program-counter step.  The key
    constraint is that no register may be written twice on the same execution
    path within a bundle.  Bundling is important because each vector maps to
    exactly one row of polynomial constraints: the columns of that row are the
    register values, and the constraint enforces the relationship between them.

- **Memories** (`pkg/zkc/vm/memory/`) — three abstract kinds reflecting the
  ZkC source-level memory declarations: read/write `Memory` (RAM), `ReadOnlyMemory`
  (ROM / input), and `WriteOnceMemory` (WOM / output).

- **Machine state** (`pkg/zkc/vm/machine/`) — a call stack of _frames_, where
  each frame holds the current function's program counter and register file.
  All memory banks are shared across the entire call stack.

The IR is therefore deliberately low-level: there are no expressions, no
structured control flow, and no symbolic names.  This makes the subsequent
register-splitting, vectorization, and constraint-generation passes
straightforward transformations over a small, well-defined instruction set.

## Register splitting

## Vectorization

Vectorization (`pkg/asm/lower.go`) merges the flat sequence of
micro-instructions produced for each function into the fewest possible
`Vector` bundles.  Each bundle corresponds to one row of polynomial
constraints, so reducing the number of bundles directly reduces the size of
the generated trace and the cost of proving.

Two concepts govern what may be placed in the same bundle: **conflicts** and
**forwarding**.

### Conflicts

A _conflicting write_ occurs when two micro-instructions in the same bundle
both assign to the same register on the same execution path.  This is
forbidden because each register corresponds to exactly one column in the row,
and a column can only hold one value.  For example:

```
x = 0 ; x = 1          // INVALID: x written twice on the same path
```

Writes on _different_ execution paths (separated by a `SkipIf`) do not
conflict, because at most one path executes for any given row.  For example:

```
skip_if x != y 2 ; r = 0 ; ret ; r = 1 ; ret
```

Here `r` is assigned on both the taken branch (`r = 0`) and the fall-through
branch (`r = 1`), but since the branches are mutually exclusive this is
valid.  The vectorizer uses a DFA that tracks two write sets per program
point within a bundle — _definitely written_ (written on every path so far)
and _maybe written_ (written on at least one path) — and rejects any bundle
where a write targets a register already in the _maybe written_ set.

### Forwarding

_Forwarding_ allows a register written by an earlier micro-instruction in a
bundle to be read by a later one within the same bundle.  This is analogous
to register forwarding in CPU pipelines.  For example:

```
x = 0 ; y = x + 1 ; ret
```

Here `x` is written by the first micro-instruction and immediately read by
the second.  Because `x` is _definitely written_ before the read, the
vectorizer considers this valid and the written value is forwarded.

Forwarding is not permitted if the prior write is only on _some_ paths.  For
example:

```
skip_if a != b 1 ; x = 0 ; y = x + 1
```

Here `x` is only _maybe written_ (the `SkipIf` may bypass the assignment),
so reading it in `y = x + 1` would be ambiguous and is rejected.
