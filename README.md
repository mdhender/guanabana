# Guanabana

Guanabana is a learning project focused on building a parser generator
in Go, inspired by the Lemon grammar file format.

This is not a direct port of Lemon. Instead, it is an educational
exploration of how parser generators work internally, with an emphasis
on practical implementation over theory.

The goal is to:

-   Learn how grammars are represented internally.
-   Implement FIRST/FOLLOW and nullable analysis.
-   Construct LR(0) and LALR(1) item sets.
-   Generate ACTION and GOTO tables.
-   Build a runtime shift/reduce parser engine.
-   Parse a simple calculator language.
-   Parse Wirth Syntax Notation (WSN).
-   Eventually parse the Lemon grammar format itself.

This project assumes the reader is a competent Go developer but new to
parsing.

------------------------------------------------------------------------

## Repository Layout

Planned structure:

    cmd/guanabana/         CLI tool
    internal/grammar/      Grammar AST structures
    internal/analysis/     FIRST/FOLLOW, nullable, etc.
    internal/lalr/         Item sets and table construction
    internal/runtime/      Table-driven parser engine
    examples/calc/         Calculator grammar and evaluator
    examples/wsn/          WSN grammar and parser
    docs/                  Course materials

------------------------------------------------------------------------

## Philosophy

-   Practical over academic.
-   One concept at a time.
-   Every major step results in runnable code or visible output.
-   Deterministic, testable components.
-   Clean, idiomatic Go.

This repository accompanies the "Guanabana 100" course outline in
`docs/COURSE-100.md`.
