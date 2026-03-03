# Intermediate Representation (IR)

The front end's AST is designed around the source language — it preserves
structure that is useful for type-checking and linking but is awkward to
translate directly into polynomial constraints. The Intermediate
Representation (IR) bridges this gap by recasting the program as a simple
**register machine** whose semantics map cleanly onto the row-per-step
structure of an arithmetized trace.

The IR describes a machine with the following components:

- **Registers** — typed storage slots, each with a declared bit-width.
  Within a single function frame the registers are identified by integer
  indices rather than names. Each function therefore has a fixed-size
  register file containing its parameters, return values, and temporaries.

- **Machine words** (`pkg/zkc/vm/word/`) — an abstraction over the concrete
  numeric type used to hold register values. During simulation an unbounded
  `Uint` (arbitrary-precision integer) is used so that overflow is detected
  rather than silently truncated. When targeting a specific prime field a
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
    executed "in parallel" within a single program-counter step. The key
    constraint is that no register may be written twice on the same execution
    path within a bundle. Bundling is important because each vector maps to
    exactly one row of polynomial constraints: the columns of that row are the
    register values, and the constraint enforces the relationship between them.

- **Memories** (`pkg/zkc/vm/memory/`) — three abstract kinds reflecting the
  ZkC source-level memory declarations: read/write `Memory` (RAM), `ReadOnlyMemory`
  (ROM / input), and `WriteOnceMemory` (WOM / output).

- **Machine state** (`pkg/zkc/vm/machine/`) — a call stack of _frames_, where
  each frame holds the current function's program counter and register file.
  All memory banks are shared across the entire call stack.

The IR is therefore deliberately low-level: there are no expressions, no
structured control flow, and no symbolic names. This makes the subsequent
register-splitting, vectorization, and constraint-generation passes
straightforward transformations over a small, well-defined instruction set.

## Example

The IR for a given `zkc` source file can be generated using the
command `zkc compile --ir test.zkc`. For example, the following IR
code might be generated for the `pow()` example from the previous
section:

```
fn pow(u4 n, u4 m) -> (u4 r) {
        u8 i
[0]     i = 0x0 ; r = 0x1
[1]     skip_if i < m 1 ; ret ; r = r * n ; i = i + 0x1 ; jmp 1
}
```

Here, we can see the original program has been aggressively
_vectorized_ (see below for more on this). Doing this reduces the
number of trace rows required to represent an instance of the
function.

## Register splitting

(to be completed)

## Vectorization

Vector instructions are instructions composed of some number of micro
instructions which, with restrictions, can be executed by the
underlying machine "in parallel". The approach is analoguous to the
concept of [Very-Long Instruction Words
(VLIW)](https://en.wikipedia.org/wiki/Very_long_instruction_word) but
taken to more of an extreme --- there is no limit on the number of
micro-instructions.

To better understand vector instructions, consider two instructions executed
in sequence:

```
[0] x = y + 1
[1] z = 0
```

When executing these instructions, an intermediate state exists after
the first instruction is executed but before the second has been where
x has been written but z has not. Alternatively, the two instructions
can be composed together to form a _vector instruction_
(`pkg/zkc/vm/instruction/vector.go`), written like so:

```
[0] x = y + 1 ; z = 0
```

In this case, both instructions are executed _in parallel_ and there
is no intermediate state where `x` is written but `z` is not.

### Vector Control-Flow

The `skip` and `skip_if` micro-instructions enable control flow
_within_ a single bundle by "skipping over" some number of the
following micro-instructions. Here, `skip n` always skips over the
following `n` micro-instructions, whilst `skip_if cond n` does when
`cond` holds. For example:

```
skip_if x != y 2 ; r = 0 ; ret ; r = 1 ; ret
```

Here the vector instruction has two execution paths: (1) when `x == y`
the skip is not taken and the machine executes `r = 0 ; ret`; (2) when
`x != y` the skipt is taken and the machine executes `r = 1 ; ret`.

### Register Conflicts

To ensure easy translation into polynomial constraints, there are
restrictions on how vector instructions can be composed. A
_conflicting write_ occurs when two micro-instructions assign to the
same register on the same execution path[^1]. For example:

```
x = 0 ; x = 1          // INVALID: x written twice on the same path
```

Writes on _different_ execution paths do not conflict, because at most
one path executes for any given row. For example:

```
skip_if x != y 2 ; r = 0 ; ret ; r = 1 ; ret
```

Here `r` is assigned on both the taken branch (`r = 0`) and the
fall-through branch (`r = 1`), but since the branches are mutually
exclusive this is valid. The vectorizer uses a data-flow analysis
that tracks two write sets per position within a vector bundle —
_definitely written_ (written on every path so far) and _maybe
written_ (written on at least one path) — and rejects any bundle where
a write targets a register already in the _maybe written_ set.

### Register Forwarding

_Forwarding_ allows a register written by an earlier micro-instruction
in a vector instruction to be read by a later one within the same
instruction. This is analogous to [register
forwarding](https://en.wikipedia.org/wiki/Operand_forwarding) as used
in CPU pipelines. For example:

```
x = 0 ; y = x + 1 ; ret
```

Here `x` is written by the first micro-instruction and immediately
read by the second. Because `x` is _definitely written_ before the
read, the vectorizer considers this valid and the written value is
said to be "forwarded"[^2].

Forwarding is not permitted if the prior write is only on _some_
paths. For example:

```
skip_if a != b 1 ; x = 0 ; y = x + 1
```

Here `x` is said to be _maybe written_ so reading it in `y = x + 1`
would be ambiguous and is rejected.

\[^1\]: This is forbidden because each register corresponds to exactly
one column in the corresponding table's row, and a column can only
hold one value.

\[^2\]: In real terms, this means constraint generated for this
instruction refers to `x` on the _current row_ of the trace
(i.e. rather than the previous row).
