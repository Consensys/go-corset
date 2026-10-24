# ZkC Architecture

ZkC follows a relatively standard compiler organisation with a
_front-end layer_, an _intermediate representation layer_ and,
finally, a _backend layer_:

- **[Front End](ZKC_ARCHITECTURE_AST.md)** — parses ZkC source files,
  links cross-file references, type-checks, validates control flow, and
  emits a fully resolved Abstract Syntax Tree (AST).

- **[Intermediate Representation](ZKC_ARCHITECTURE_IR.md)** — lowers the
  AST into a simple register-machine IR, vectorizes the flat instruction
  stream into VLIW-style bundles, and optionally splits wide registers to
  fit within the target field's bandwidth.

- **[Arithmetization Layer](ZKC_ARCHITECTURE_AIR.md)** — compiles the
  vectorized IR into vanishing constraints, range constraints, and lookup
  arguments that form the final arithmetized circuit.

- **[Execution and Tracing](ZKC_ARCHITECTURE_VM.md)** — runs the
  bytecode in an interpreter, records each step as a trace row, and
  optionally cuts long runs into shards.

```mermaid
flowchart TB
    SRC[".zkc source files"]

    subgraph FE["Front end: pkg/zkc/compiler"]
        PARSE["Lexer + parser<br/>parser/"]
        LINK["Linker: names -> indices<br/>linker.go"]
        VAL["Validate: types, control flow, ...<br/>validate/"]
        LOW["Lower: structured control flow to goto<br/>lower/"]
        CG["Codegen: AST -> bytecode<br/>codegen/"]
    end

    subgraph MID["Middle: pkg/zkc/vm"]
        PROG["vm.Program (raw bytecode,<br/>one module per function / memory)"]
        TEXEC["TransformForExecution<br/>inline, vectorize, split registers"]
        TTRACE["TransformForTracing<br/>+ lowering, timestamps, range checks"]
    end

    subgraph RUN["Execution: pkg/zkc/vm"]
        FAST["Interpreter, fast mode<br/>(outputs; checkpoints for sharding)"]
        TRACEMODE["Interpreter, tracing mode<br/>one row per executed vector"]
        TB["Trace builder<br/>pkg/zkc/constraints/trace"]
    end

    subgraph AR["Arithmetization: pkg/zkc/constraints + pkg/ir"]
        MIR["MIR constraints<br/>(vanish, lookup, bus)"]
        AIR["AIR constraints<br/>mir.LowerToAir"]
    end

    BIN[("binary file<br/>header + raw bytecode")]
    TRACE[("trace: one table per module,<br/>padded; possibly sharded")]
    PROVER(["Prover: AIR + trace"])

    SRC --> PARSE --> LINK --> VAL --> LOW --> CG --> PROG
    PROG --> BIN
    BIN --> TEXEC --> FAST
    BIN --> TTRACE --> TRACEMODE --> TB --> TRACE
    FAST -->|"checkpoints"| TRACEMODE
    TTRACE --> MIR --> AIR
    AIR --> PROVER
    TRACE --> PROVER
```
