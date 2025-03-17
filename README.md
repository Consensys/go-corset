# Go Corset

The `go-corset` tool is a compiler for arithmetic constraints written
in a lisp-like [domain-specific
language](https://en.wikipedia.org/wiki/Domain-specific_language)
(called _Corset_).  The tool is specifically designed for used with
the [Linea constraint
system](https://github.com/Consensys/linea-constraints/) but could, in
principle, be used elsewhere.  The tool is based upon (but now
supercedes) the original Rust-based [`corset`
compiler](https://github.com/Consensys/corset) and
[language](https://github.com/Consensys/corset/wiki/The-Corset-Language).

### Table of Contents:

- [Overview](#overview)
- [Command-Line Interface](#command-line-interface)
- [Contributing](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Developers](#developer-setup)

## Overview

The purpose of the Corset language is to provide a more human-friendly
interface for writing arithmetic constraints.  Corset constraints are
compiled down into an Arithmetic Intermediate Representation (AIR)
consisting of vanishing constraints, range constraints, lookup
arguments and permutation arguments.

An example program written in Corset is:

```lisp
(defcolumns
  (NIBBLE :i4)
  (BIT_0 :i1@prove)
  (BIT_1 :i1@prove)
  (BIT_2 :i1@prove)
  (BIT_3 :i1@prove))

;; NIBBLE = 8*BIT_3 + 4*BIT_2 + 2*BIT_1 + BIT_0
(defconstraint decomp () (eq! NIBBLE (+ BIT_0 (* 2 BIT_1) (* 4 BIT_2) (* 8 BIT_3))))
```

This program ensures all rows of the column `NIBBLE` are in the range
`0..15` (inclusive) by decomposing its value into exactly four bits.
In turn, the rows of each `BIT_X` column are enforced using a range
constraint (specified via the `@prove` modifier).

## Command-Line Interface

The `go-corset` tool provides a toolbox of commands for working with
constraints and traces.  For example, we can compile source files into
binary format; or, check traces against constraints; or, inspect a
trace or binary constraint file, etc.  We now examine a selection of
the most useful top-level commands:

- [`go-corset check`](#check) allows one to check whether a given
  trace (or batch of traces) satisfies a given set of constraints.  If
  a failure arises, a useful error report can be provided.
- [Compile](#check).  This command allows one to compile a given set
  of Corset source files into a single binary file.  This is useful
  for packaging up constraints for use with other tools, etc.
- [Debug](#debug).  This command provides various ways of examining a
  given set of constraints.  For example, one can print out the low
  level arithmetic intermediate representation (AIR) which is
  generated; or, one can generate summary statistics (e.g. number of
  columns, number of constraints, etc); or, look at any metadata
  embedded within a binary constraint file, etc.
- [Inspect](#inspect).  This command provides an interactive trace
  visualisation tool to assist debugging.  This is not graphical, but
  runs in a terminal and supports general queries over the trace
  (e.g. find a row where column `CT > 0`, etc).
- [Trace](#trace).  This command allows ones to inspect and/or
  manipulate a given trace file in various ways.  For example, one can
  obtain statistical information such as the total number of cells
  contained within; or, the number of unique elements in a given
  column; or, to print the values of certain columns on specific rows;
  or, to convert the trace file into a different format (e.g. JSON) or
  trim the trace file in some way (e.g. keeping only the first `n`
  rows, etc); or, to view any metadata embedded within the trace file.

### Check

### Compile

### Debug

### Inspect

### Trace

## Developer Setup

**Step 0.** Install [pre-commit](https://pre-commit.com/):

```shell
pip install pre-commit

# For macOS users.
brew install pre-commit
```

Then run `pre-commit install` to setup git hook scripts.
Used hooks can be found [here](.pre-commit-config.yaml).

______________________________________________________________________

NOTE

> `pre-commit` aids in running checks (end of file fixing,
> markdown linting, go linting, runs go tests, json validation, etc.)
> before you perform your git commits.

______________________________________________________________________

**Step 1.** Install external tooling (golangci-lint etc.):

```shell script
make install
```

**Step 2.** Setup project for local testing (code lint, runs tests, builds all needed binaries):

```shell script
make all
```

______________________________________________________________________

NOTE

> All binaries can be found in `<project_root>/bin` directory.
> Use `make clean` to delete old binaries.
>
> Check [Makefile](Makefile) for other useful commands.

______________________________________________________________________
