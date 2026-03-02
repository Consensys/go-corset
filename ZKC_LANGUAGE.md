# ZkC Language

ZkC is a simple imperative language for writing zero-knowledge programs. It provides
a C-like syntax with fixed-width unsigned integer types and explicit memory
declarations. ZkC programs are compiled and executed by the `zkc` toolchain, which
is part of the `go-corset` repository.

## Program Structure

A ZkC source file is a sequence of top-level declarations:

- **`function`** — a callable subroutine
- **`memory`** — a private read/write RAM bank
- **`public input` / `private input`** — a read-only (write-once from the outside) memory
- **`public output` / `private output`** — a write-once output memory
- **`const`** — a named compile-time constant
- **`include`** — include another source file

The entry point for execution is a function named `main`.

## Types

ZkC currently supports unsigned integer types of arbitrary bit-width:

```
u1   u2   u4   u8   u16   u32   u64   ...
```

The syntax is `u<N>` where `N` is the number of bits.  There are no signed
integer, floating-point, struct, or array types at this time.

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

## Functions

```zkc
function name(u8 param1, u16 param2) -> (u32 result) {
  // body
  result = param1 + param2
  return
}
```

- Parameters are immutable; they cannot be assigned to inside the body.
- Return values are mutable local variables that must be set before `return`.
- Every execution path must end with `return` (or `fail`).

## Variables

Local variables are declared inside a function body with an optional initialiser:

```zkc
u8 x          // declared, initially undefined
u8 y = 42     // declared and initialised
```

Only a single variable may carry an initialiser per declaration statement.

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

## Statements

### Assignment

```zkc
target = expression
```

A single expression result can be split across multiple targets listed on the
left-hand side.  The result is distributed big-endian across the targets, which
is useful for capturing carry bits:

```zkc
carry, lo = lo + 1   // lo gets the low bits; carry gets the overflow bit
```

### Conditionals

```zkc
if x > y {
  z = x
} else {
  z = y
}
```

The `else` branch is optional.

### While Loop

```zkc
while i < 10 {
  i = i + 1
}
```

### For Loop

```zkc
for i = 0; i < 10; i = i + 1 {
  // body
}
```

### Return and Fail

```zkc
return   // normal return from the current function
fail     // signal an exceptional (error) termination
```

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
