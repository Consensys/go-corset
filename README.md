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

This program ensures all values in the column `NIBBLE` are in the
range `0..15` (inclusive).  This is done by decomposing the value of
`NIBBLE` on any given row into exactly four bits.  In turn, the values
of each `BIT_X` column are enforced using a range constraint
(specified via the `@prove` modifier).

## Command-Line Interface

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
