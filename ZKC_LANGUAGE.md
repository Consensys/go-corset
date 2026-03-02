# ZkC Language

ZkC is a simple imperative language for writing zero-knowledge programs. It provides
a C-like syntax with fixed-width unsigned integer types and explicit memory
declarations. ZkC programs are compiled and executed by the `zkc` toolchain, which
is part of the `go-corset` repository.

## Functions

Functions are the fundamental building blocks of ZkC programs:

```zkc
function max(u16 x, u16 y) -> (u16 res) {
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

In the above, parameters `x` and `y` are declared to be `u16`.  In
fact, unlike many other languages, ZkC supports unsigned integer types
of arbitrary bitwidth, such as: `u2`, `u3`, `u11`, `u15`, `u48`,
`u160`, etc.

There are some restrictions imposed upon functions.  For example,
parameters cannot be assigned and are immutable for the duration of a
function.  Likewise, return values and local variables must be defined
being used --- there are no default values in ZkC.

## Inputs

## Outputs

## Read / Write Memory

## Constants

Named constants can be declared at the top level and used in expressions:

```zkc
const MAX_ADDR = 255
const MASK     = 0xFF
const FLAGS    = 0b00001111
const SIZE     = 2^10
```

Numeric literals may be written in decimal, hexadecimal (`0x…`), or binary
(`0b…`), and may use `_` as a visual separator (e.g. `0xFF_FF`).

## Variable Declarations

Local variables are declared inside a function body with an optional initialiser:

```zkc
u8 x          // declared, initially undefined
u8 y = 42     // declared and initialised
```

Only a single variable may carry an initialiser per declaration statement.

### Assignments

```zkc
target = expression
```

A single expression result can be split across multiple targets listed on the
left-hand side.  The result is distributed big-endian across the targets, which
is useful for capturing carry bits:

```zkc
carry, lo = lo + 1   // lo gets the low bits; carry gets the overflow bit
```

### Loops

In addition to `if` conditions (as seen above), ZkC supports `while`
and `for` loops using a familiar syntax:

```zkc
u8 i = 0

while i < 10 {
  i = i + 1
  // body
}
```

The above can be written equivalently using the `for`-loop syntax:

```zkc
for i = 0; i < 10; i = i + 1 {
  // body
}
```

**NOTE**: loops in a function necessarily force it to be a so-called
_multi-line function_ (see below).  As such, in many cases, it can be
more efficient (in terms of generated constraints) to use recursion.

### Exceptions

ZkC supports a `fail` instruction which is similar to a `panic` in
other languages.  If executed, this immediately terminates the
machine.  More importantly, however, is that the generated constraints
cannot hold for any execution which reaches a `fail`.  The following
illustrates:

```zkc
function divide(u16 x, u16 y) -> (u16 r) {
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
|----------|----------------|
| `a + b`  | addition       |
| `a - b`  | subtraction    |
| `a * b`  | multiplication |

**Parentheses are required** when mixing operators.  `a + b * c` is a syntax
error; write `a + (b * c)` instead.

Comparison operators (used in conditions only):

| Operator | Meaning               |
|----------|-----------------------|
| `a == b` | equal                 |
| `a != b` | not equal             |
| `a < b`  | less than             |
| `a <= b` | less than or equal    |
| `a > b`  | greater than          |
| `a >= b` | greater than or equal |

## Memory Declarations

Memories are declared at the top level and accessed inside functions like
function calls.

### Read/Write RAM

```zkc
memory buf(u8 address) -> (u8 value)
```

RAM can be read and written freely from any function.

### Read-Only Input

```zkc
public  input rom(u8 address) -> (u8 value)
private input cfg(u16 address) -> (u32 value)
```

Input memories are provided externally before execution and cannot be written.

### Write-Once Output

```zkc
public  output result(u8 address) -> (u8 value)
private output log(u16 address) -> (u32 value)
```

Output memories can be written at most once per address.

## File Inclusion

```zkc
include "path/to/other.zkc"
```

Included files are processed as if their declarations appeared at the include
site.

## A Complete Example

```zkc
// Return the larger of two 4-bit values.
function max(u4 x, u4 y) -> (u4 z) {
  if x > y {
    z = x
  } else {
    z = y
  }
  return
}
```

### Constraint Compilation

The ultimate goal of ZkC is to compile programs down to the arithmetic
constraint systems used by ZK provers which consist of: **vanishing
constraints**, **lookup arguments**, and **permutation arguments**.

The underlying arithmetic constraint system is broken up into
_modules_, each of which is composed of one or more columns.
Polynomial constraints can be enforced on modules, whilst lookup
arguments can hold between columns either in the same module, or
between different modules.  Finally, permutation arguments hold
between columns in the same module.  A trace in this system is
likewise broken up into modules, where each module contains zero or
more rows of values (with one for each column).

In the ZkC system, each function body is implemented as a module
containing polynomial constraints generated from the instructions of
the function.  Likewise, function calls are implemented as lookups
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
