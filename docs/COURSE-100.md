# Guanabana 100 --- Build a Parser Generator in Go

This is a 100-level, practical course on building a parser generator
inspired by the Lemon grammar file format.

The focus is on implementation and understanding through building.

------------------------------------------------------------------------

## Course Outcomes

By the end of this course, you will have:

1.  A working LALR(1) parser generator in Go.
2.  A table-driven shift/reduce runtime parser.
3.  A parser for a simple calculator language.
4.  A parser for Wirth Syntax Notation (WSN).
5.  A grammar (written in Lemon-style format) for the Lemon grammar
    subset.

------------------------------------------------------------------------

## Lesson Plan

### Lesson 0 --- Orientation

-   What a parser generator is.
-   Overview of grammar → tables → runtime.
-   Project structure setup.

Deliverable: Repository skeleton.

------------------------------------------------------------------------

### Lesson 1 --- Grammars as Data Structures

-   Terminals vs nonterminals.
-   Productions and start symbols.
-   Symbol indexing and normalization.

Deliverable: `internal/grammar` package.

------------------------------------------------------------------------

### Lesson 2 --- FIRST, FOLLOW, and Nullable

-   Nullable computation.
-   FIRST sets.
-   FOLLOW sets.
-   Unit tests for validation.

Deliverable: `internal/analysis` package.

------------------------------------------------------------------------

### Lesson 3 --- LR(0) Items

-   Item definition.
-   closure() and goto().
-   Building the LR(0) automaton.

Deliverable: State machine dump tool.

------------------------------------------------------------------------

### Lesson 4 --- LALR(1) Lookaheads

-   Kernel items.
-   Lookahead propagation.
-   State merging.
-   Observing conflicts.

Deliverable: LALR state construction.

------------------------------------------------------------------------

### Lesson 5 --- Parse Table Construction

-   ACTION and GOTO tables.
-   Shift, reduce, accept, error.
-   Conflict reporting.

Deliverable: Table builder with diagnostics.

------------------------------------------------------------------------

### Lesson 6 --- Precedence and Associativity

-   %left, %right, %nonassoc.
-   Conflict resolution rules.
-   Expression grammar validation.

Deliverable: Precedence-aware parser tables.

------------------------------------------------------------------------

### Lesson 7 --- Runtime Parser Engine

-   Stack machine model.
-   Shift/reduce execution loop.
-   Semantic value handling.

Deliverable: Table-driven runtime parser.

------------------------------------------------------------------------

### Lesson 8 --- Calculator Example

-   Arithmetic grammar.
-   Precedence handling.
-   Expression evaluation.

Deliverable: Working calculator CLI example.

------------------------------------------------------------------------

### Lesson 9 --- Lemon Grammar Parsing

-   Subset definition.
-   Grammar parsing into AST.
-   Table generation from grammar files.

Deliverable: Lemon-style grammar parser.

------------------------------------------------------------------------

### Lesson 10 --- Wirth Syntax Notation (WSN)

-   WSN grammar parsing.
-   AST construction.
-   Optional transpilation to Lemon-style format.

Deliverable: WSN parser.

------------------------------------------------------------------------

## Non-Goals (100-Level)

-   GLR parsing.
-   Scannerless parsing.
-   Deep formal proofs.
-   Full Lemon compatibility immediately.

------------------------------------------------------------------------

This course is designed to be iterative, practical, and enjoyable. Each
lesson builds directly on the previous one.
