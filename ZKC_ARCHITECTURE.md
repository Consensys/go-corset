# ZkC Architecture

ZkC follows a relatively standard compiler organisation with a
_front-end layer_, an _intermediate representation layer_ and,
finally, a _backend layer_.

## Front End

The front end parses ZkC source files and produces an _Abstract Syntax Tree_
(AST) — a structured, in-memory representation of the program that subsequent
compiler phases can analyse and transform.

An AST is a tree where each node represents a syntactic construct.  The root
of the tree is a **program**, which is a list of top-level **declarations**
(functions, memories, constants).  A function declaration carries a list of
typed **variables** (parameters, return values, and locals) together with a
flat sequence of **statements** that form its body.  Each statement in turn
may contain one or more **expressions**, which describe how values are
computed.

The concrete node types are:

- **Declarations** (`pkg/zkc/compiler/ast/decl/`) — `Function`, `Memory`
  (read-only input, write-once output, or read/write RAM), and `Constant`.
- **Statements** (`pkg/zkc/compiler/ast/stmt/`) — `Assign`, `IfGoto`
  (conditional branch), `Goto` (unconditional branch), `Return`, and `Fail`.
  Note that structured control flow (`if`/`else`, `while`, `for`) is lowered
  to `IfGoto`/`Goto` pairs during parsing so that later phases work with a
  simple, uniform instruction set.
- **Expressions** (`pkg/zkc/compiler/ast/expr/`) — `Add`, `Sub`, `Mul`,
  `Const` (numeric literal), `LocalAccess` (read a local variable), and
  `NonLocalAccess` (reference to a constant or memory declared elsewhere).
- **Types** (`pkg/zkc/compiler/ast/data/`) — currently `UnsignedInt` (`uN`)
  and `Tuple` (a composite of multiple `uN` fields used for multi-word memory
  buses).

Every expression node exposes a `BitWidth()` method that returns the minimum
number of bits required to hold any possible result.  The compiler uses this
throughout type-checking and code generation to detect overflow and to size
the registers that hold intermediate values.

References between declarations (a function calling another function, an
expression reading a constant defined in a different file, etc.) are initially
represented as **unresolved symbols** — plain string names.  The _linker_
phase replaces these with **resolved symbols** that point directly to the
target declaration, completing the AST before the later phases run.

### Parsing

The parser (`pkg/zkc/compiler/parser/`) converts a ZkC source file into an
`UnlinkedSourceFile` — a list of declarations whose cross-file references are
still unresolved string names.  It proceeds in two sub-phases:

1. **Lexing** — the source text is tokenised into a flat token stream.
   Whitespace and comments are discarded.  The lexer recognises keywords
   (`function`, `memory`, `input`, `output`, `if`, `while`, `for`, `return`,
   `fail`, `const`, `include`, …), identifiers, numeric literals (decimal,
   hex, binary, and `2^N` exponent form), and operator symbols.

1. **Recursive-descent parsing** — the token stream is consumed top-down by a
   hand-written recursive-descent parser.  At the top level it dispatches on
   the leading keyword to parse each declaration kind (`parseFunction`,
   `parseConstant`, `parseInputOutputMemory`, `parseReadWriteMemory`,
   `parseInclude`).  Inside a function body, statements are parsed one at a
   time; structured control flow (`if`/`else`, `while`, `for`) is immediately
   lowered to sequences of `IfGoto` and `Goto` instructions so that all later
   phases work with a uniform, flat instruction list.  A source map is built
   in parallel, recording the span of source text that corresponds to each
   instruction, for use in later error messages.

The parser halts and returns the accumulated errors on the first declaration
that cannot be parsed, rather than attempting error recovery.

### Linking

The linker (`pkg/zkc/compiler/linker.go`) merges the `UnlinkedSourceFile`
produced by each parsed file into a single `ast.Program` and replaces every
unresolved symbol with a `symbol.Resolved` — an integer index into the
program's flat declaration table.

It works in two passes:

1. **Registration** — every declaration from every source file is entered into
   a name-to-index map (`busmap`).  Duplicate names are rejected immediately
   with an error.

1. **Resolution** — each declaration is rewritten by walking its instructions
   and expressions.  Most nodes are copied unchanged; a `NonLocalAccess` node
   (a reference to a constant or memory by name) is looked up in the map and
   replaced with a `NonLocalAccess[symbol.Resolved]` carrying the target's
   index.  Arity (number of inputs and outputs) is checked at the same time:
   if the name exists but the declared arity does not match the call site, an
   error is reported.

Local variable references (`LocalAccess`) and control-flow targets (`Goto`,
`IfGoto`) are already integer indices at this point — they are set during
parsing — so the linker leaves them untouched.  Source-map entries are copied
from the unresolved nodes to their resolved counterparts so that subsequent
error messages still point to the original source locations.

### Typing

ZkC uses a **value-range type system** (`pkg/zkc/compiler/validate/typing/`).
Rather than tracking only the declared type of a variable, the type checker
propagates the _maximum possible value_ of every sub-expression as a
`big.Int`.  The bit-width of a type is then simply the number of bits required
to represent that maximum value (`BitLen()`).  This means types are inferred
bottom-up through expressions rather than checked top-down against a declared
type.

The rules are straightforward:

- A **constant literal** has a maximum value equal to its own value.
- A **variable access** has a maximum value of `2^N - 1`, where `N` is the
  register's declared bit-width.
- **Addition** sums the maximum values of its operands (so `u8 + u8` produces
  a type with maximum value 510, i.e. effectively `u9`).
- **Multiplication** multiplies the maximum values.
- **Subtraction** keeps the maximum value of the minuend (the left-hand side),
  but separately records the maximum value of the subtrahend so that the
  correct signed bit-width can be allocated later.

After the type of the right-hand side of an assignment is inferred, the checker
verifies that its bit-width does not exceed the total bit-width of the
left-hand side targets.  If it does, a _bit overflow_ error is reported.  The
checker also rejects assignments to parameter registers (which are immutable)
and duplicate writes to the same target within one assignment.  The inferred
bit-widths are written back into the expression nodes and used by later passes.

### Control-Flow Analysis

Control-flow analysis (`pkg/zkc/compiler/validate/control_flow.go`) checks two
properties of every function using a standard **worklist-based dataflow
analysis** over the flat `IfGoto`/`Goto` instruction list:

1. **Definite assignment** — every register must be definitely assigned before
   it is read.  The dataflow state at each program point is a bit-set of
   registers that are _not yet assigned_.  Parameters are removed from the set
   on entry (they arrive pre-assigned); all other registers start in the set.
   As instructions are processed, registers written by an instruction are
   removed from the set, and any register read while still in the set triggers
   an "possibly undefined" error.

1. **Termination** — every control-flow path must end at a `Return` or `Fail`
   instruction.  A `Return` additionally checks that all return-value registers
   have been removed from the undefined set (i.e. are definitely assigned).
   Falling off the end of the instruction list without a `Return` or `Fail`
   is an error.  Any instruction that was never reached by the worklist is
   flagged as unreachable.

The worklist drives the analysis: each entry is a `(pc, state)` pair.  When a
branch instruction is processed its successor program points are joined into
the worklist — the join is a set-union of undefined registers, conservatively
assuming either path may be taken.  A successor is only re-queued when the
join changes its state, ensuring termination.

### Lowering

### Code Generation

Code generation (`pkg/zkc/compiler/codegen/`) translates the validated
`ast.Program` into a `machine.Boot` — a fully initialised IR machine ready
for execution or further transformation.  It is a single-pass, structurally
recursive translation with no optimisation.

The top-level `Compile` function iterates over every declaration:

- **Constants** are skipped; their values were already inlined by the type
  checker wherever they appear in expressions.
- **Memory declarations** are turned into the appropriate `memory.Boot`
  instance (static ROM, input ROM, output WOM, or RAM) and collected into
  the corresponding slot of the machine.
- **Functions** are handed to `compileFunction`, which first flattens each
  AST `variable.Descriptor` into one or more `register.Register` entries
  (a tuple variable expands into one register per field), then compiles
  every statement in the function body one-to-one into an IR `Instruction`:
  - `Assign` → a `Vector` of `Add` / `Mul` micro-instructions.  Each
    sub-expression is compiled recursively; non-atomic operands (anything
    that is not a plain variable access) are spilled into a fresh temporary
    register allocated on the fly.  Constant and `NonLocalAccess` operands
    are evaluated eagerly to a `word.Uint` and folded into the instruction's
    constant field.
  - `IfGoto` → a `Vector` containing the comparison operand setup, a
    `SkipIf`, and two `Jmp` micro-instructions (one for the taken branch,
    one for fall-through).
  - `Goto` → a bare `Jmp`.
  - `Return` / `Fail` → the corresponding terminal instructions.

The result is a `machine.Boot` value that bundles the compiled functions
with their memory banks, ready to be booted with a concrete set of inputs.

## Intermediate Representation (IR)

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

### Register splitting

### Vectorization

## Arithmetization Layer

### Compilation

- Forwarding.

### Branch Table Optimisation

### Other optimisations (e.g. binary conditions, conjunctions)
