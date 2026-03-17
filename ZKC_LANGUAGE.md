# ZkC Language

ZkC is a simple imperative language for writing zero-knowledge programs. It provides
a C-like syntax with fixed-width unsigned integer types and explicit memory
declarations. ZkC programs are compiled and executed by the `zkc` toolchain, which
is part of the `go-corset` repository.

## Functions

Functions are the fundamental building blocks of ZkC programs:

```zkc
fn max(x:u16, y:u16) -> (res:u16) {
  // body
  if x > y {
    res = x
  } else {
    res = y
  }
  //
  return
}
```

In the above, parameters `x` and `y` are declared to be `u16`. In
fact, unlike many other languages, ZkC supports unsigned integer types
of arbitrary bitwidth, such as: `u2`, `u3`, `u11`, `u15`, `u48`,
`u160`, etc.

There are some restrictions imposed upon functions. For example,
parameters cannot be assigned and are immutable for the duration of a
function. Likewise, return values and local variables must be defined
being used --- there are no default values in ZkC.

## Inputs

Input memories provide the means for the communicating input data from
the external environment to be used within an executing ZkC program.

```zkc
pub input rom(address:u8) -> (value:u8)
    input cfg(address:u16) -> (value:u32)
```

Input memories are a form of _read-only memory_ and cannot be written
during execution. Inputs can optionally be marked `pub`; unmarked
inputs are private by default. The distinction is determined by the
proving system in use: `pub` inputs are committed to whilst private
inputs form part of the witness.

## Outputs

Output memories provide the means for the communicating data generated
whilst executing a ZkC program back to the external environment.

```zkc
pub output result(address:u8) -> (value:u8)
    output log(address:u16) -> (value:u32)
```

Output memories are a form of _write-once memory_ meaning that each
location must be written exactly once. As with input memories, the
`pub` modifier makes an output public; outputs are private by default.

## Read / Write Memory

Read/Write Memory (a.k.a Random-Access Memory) provides a form of
unbound internal storage. Multiple read/write memories can be defined
with different types as part of a ZkC program.

```zkc
memory buf(address:u8) -> (value:u8)
```

Observe that any data written into a read/write memory is not
available to the external environment once execution has completed.
As such, any important data needing to be communicated back must be
written into an output memory before execution terminates.

## Constants

Named constants can be declared at the top level and used in expressions:

```zkc
const MAX_ADDR:u8  = 255
const MASK:u8      = 0xFF
const FLAGS:u8     = 0b00001111
const SIZE:u16     = 2^10
```

Numeric literals may be written in decimal, hexadecimal (`0x…`), or binary
(`0b…`), and may use `_` as a visual separator (e.g. `0xFF_FF`).

## Variable Declarations

Local variables are declared inside a function body with an optional initialiser:

```zkc
var x:u8          // declared, initially undefined
var y:u8 = 42     // declared and initialised
```

Only a single variable may carry an initialiser per declaration statement.

### Assignments

```zkc
target = expression
```

A single expression result can be split across multiple targets listed on the
left-hand side. The result is distributed big-endian across the targets, which
is useful for capturing carry bits:

```zkc
carry, lo = lo + 1   // lo gets the low bits; carry gets the overflow bit
```

### Loops

In addition to `if` conditions (as seen above), ZkC supports `while`
and `for` loops using a familiar syntax:

```zkc
var i:u8 = 0

while i < 10 {
  i = i + 1
  // body
}
```

The above can be written equivalently using the `for`-loop syntax:

```zkc
for i:u8 = 0; i < 10; i = i + 1 {
  // body
}
```

**NOTE**: loops in a function necessarily force it to be a so-called
_multi-line function_ (see below). As such, in many cases, it can be
more efficient (in terms of generated constraints) to use recursion.

### Exceptions

ZkC supports a `fail` instruction which is similar to a `panic` in
other languages. If executed, this immediately terminates the
machine. More importantly, however, is that the generated constraints
cannot hold for any execution which reaches a `fail`. The following
illustrates:

```zkc
fn divide(x:u16, y:u16) -> (r:u16) {
  // sanity check that y != 0
  if y == 0 { fail }
  ...
}
```

In this case, the `fail` instruction is being used to enforce an
expected precondition to the function.

## Expressions

ZkC supports the following arithmetic operators:

| Operator | Meaning        |
| -------- | -------------- |
| `a + b`  | addition       |
| `a - b`  | subtraction    |
| `a * b`  | multiplication |

Bitwise operators:

| Operator | Meaning                               |
| -------- | ------------------------------------- |
| `a & b`  | bitwise AND                           |
| `a \| b` | bitwise OR                            |
| `a ^ b`  | bitwise XOR                           |
| `~a`     | bitwise NOT (complement within width) |
| `a << b` | left shift (result masked to width)   |
| `a >> b` | right shift                           |

All operands of a binary bitwise or shift expression must have the same
type. The result type equals the operand type. For shifts, the shift
amount must be the same type as the value being shifted; left-shift
results are masked to the declared bit width of the target.

**Parentheses are required** when mixing operators of different kinds.
Chains of the _same_ operator are permitted without extra parentheses:

```zkc
// OK — same operator chained
var r:u8 = x & y & z
var s:u8 = x << 1 << 2

// OK — different operators, disambiguated with braces
var t:u8 = (x & y) | z
var u:u8 = (x << 2) >> 1

// ERROR — mixing operators without braces
var bad:u8 = x & y | z
var bad2:u8 = x << y >> z
```

Comparison operators (used in conditions only):

| Operator | Meaning               |
| -------- | --------------------- |
| `a == b` | equal                 |
| `a != b` | not equal             |
| `a < b`  | less than             |
| `a <= b` | less than or equal    |
| `a > b`  | greater than          |
| `a >= b` | greater than or equal |

## Type Aliases

Type aliases introduce a new name for an existing type. They are useful
for improving readability and for defining domain-specific names (e.g.
`address` for `u160`, `bool` for `u1`):

```zkc
type address = u160
type bool    = u1
```

Circular alias definitions are rejected.

Aliases can be used anywhere a type is expected: in function parameters
and returns, variable declarations, constants, casts, and expressions.
An alias and its underlying type are interchangeable for type-checking
purposes (e.g. a `word` and a `u8` of the same bitwidth are compatible
in arithmetic, shifts, and comparisons).

```zkc
type word = u8
fn f(x:word) -> (r:u8) {
  var r = x << 1
  return
}
```

## File Inclusion

```zkc
include "path/to/other.zkc"
```

Included files are processed as if their declarations appeared at the include
site.

### Constraint Compilation

The ultimate goal of ZkC is to compile programs down to the arithmetic
constraint systems used by ZK provers which consist of: **vanishing
constraints**, **lookup arguments**, and **permutation arguments**.

The underlying arithmetic constraint system is broken up into
_modules_, each of which is composed of one or more columns.
Polynomial constraints can be enforced on modules, whilst lookup
arguments can hold between columns either in the same module, or
between different modules. Finally, permutation arguments hold
between columns in the same module. A trace in this system is
likewise broken up into modules, where each module contains zero or
more rows of values (with one for each column).

In the ZkC system, each function body is implemented as a module
containing polynomial constraints generated from the instructions of
the function. Likewise, function calls are implemented as lookups
between modules and map parameters / returns at the call site to the
corresponding inputs / outputs of the function being called.

## CLI Usage

```shell
# Parse and type-check one or more source files; --ir prints the AST
zkc compile [--field BLS12_377] [--ir] file.zkc ...

# Execute a program; input.json provides the entry-point inputs
zkc execute [--field BLS12_377] input.json file.zkc ...
```

Available fields: `BLS12_377` (default), `KOALABEAR_16`, `GF_8209`, `GF_251`.

## Further Reading

- Source code lives under `pkg/zkc/` (compiler, AST, VM).
- Test fixtures are in `testdata/zkc/` (valid programs under `unit/`, expected
  errors under `invalid/`).
- The broader `go-corset` compilation pipeline is described in `CLAUDE.md`.
