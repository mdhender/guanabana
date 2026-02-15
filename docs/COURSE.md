# Guanabana Course: Building an LALR(1) Parser Generator

## Course Overview

This course builds a working LALR(1) parser generator in Go, step by step.
By the end, you will have a tool that reads Lemon-style grammar files and
generates Go parsers.

**Prerequisites**: You should be comfortable writing Go (structs, interfaces,
slices, maps, the `testing` package). No prior parsing knowledge is assumed.

**Time estimate**: Each lesson is designed for 30–45 minutes. The full course
is 16 lessons.

**How to use this course**: Work through the lessons in order. Each lesson
builds on the previous ones. Every lesson has unit tests that serve as the
pass/fail gate — your code is correct when the tests pass.

To start a lesson with an AI assistant, say:
> "Let's start the lesson as specified in docs/lesson-XX-slug.md."

---

## Lesson Map

### Phase 1: Lexer Contract and Grammar Input (Lessons 01–07)

These lessons establish the token stream contract, build the grammar model,
and parse Lemon grammar files into that model.

| # | Lesson | Package | Key Deliverable |
|---|--------|---------|-----------------|
| 01 | [Token Contract](lesson-01-token-contract.md) | `internal/lex` | Token types, Token struct, token stream interface |
| 02 | [Lexer Validation](lesson-02-lexer-validation.md) | `internal/lex` | Tests validating lexer output for representative grammars |
| 03 | [Grammar Model](lesson-03-grammar-model.md) | `internal/grammar` | Symbol table, Rule struct, Grammar struct |
| 04 | [Grammar Parsing](lesson-04-grammar-parsing.md) | `internal/grammar` | Parser that builds Grammar from token stream |
| 05 | [Precedence and Directives](lesson-05-precedence-directives.md) | `internal/grammar` | Precedence, associativity, %left/%right/%nonassoc |
| 06 | [Grammar Validation](lesson-06-grammar-validation.md) | `internal/grammar` | Start symbol, reachability, error diagnostics |
| 07 | [Self-Grammar](lesson-07-self-grammar.md) | `internal/grammar` | Lemon grammar of Lemon; validate lexer→parser integration |

### Phase 2: Analysis (Lessons 08–09)

Compute the sets needed for parser table construction.

| # | Lesson | Package | Key Deliverable |
|---|--------|---------|-----------------|
| 08 | [Nullable and FIRST Sets](lesson-08-nullable-first.md) | `internal/analysis` | Nullable set, FIRST sets for all symbols |
| 09 | [FOLLOW Sets](lesson-09-follow-sets.md) | `internal/analysis` | FOLLOW sets for all nonterminals |

### Phase 3: LALR(1) Construction (Lessons 10–14)

Build the LR(0) automaton, promote it to LALR(1), detect and resolve
conflicts, and produce parse tables.

| # | Lesson | Package | Key Deliverable |
|---|--------|---------|-----------------|
| 10 | [LR(0) Items and Closure](lesson-10-lr0-items-closure.md) | `internal/lalr` | Item type, Closure algorithm |
| 11 | [GOTO and Canonical Collection](lesson-11-goto-canonical.md) | `internal/lalr` | GOTO function, canonical LR(0) state set |
| 12 | [LALR(1) Lookaheads](lesson-12-lalr1-lookaheads.md) | `internal/lalr` | Lookahead propagation, state merging |
| 13 | [Conflict Detection and Resolution](lesson-13-conflict-resolution.md) | `internal/lalr` | Shift/reduce + reduce/reduce detection, precedence resolution |
| 14 | [Parse Table Construction](lesson-14-parse-tables.md) | `internal/lalr` | ACTION table, GOTO table, serialization |

### Phase 4: Code Generation and End-to-End (Lessons 15–16)

Generate parsers and prove they work on real grammars.

| # | Lesson | Package | Key Deliverable |
|---|--------|---------|-----------------|
| 15 | [Code Generation](lesson-15-codegen.md) | `internal/codegen` | Generate standalone .go parser file; package option |
| 16 | [End-to-End Examples](lesson-16-end-to-end.md) | `examples/` | Calculator + WSN/DSL grammars, full pipeline tests |

---

## Dependency Graph

```
Lesson 01 ──► Lesson 02 ──► Lesson 03 ──► Lesson 04 ──► Lesson 05 ──► Lesson 06 ──► Lesson 07
                                                                                        │
                                                                                        ▼
              Lesson 08 (Nullable + FIRST) ──► Lesson 09 (FOLLOW)
                                                     │
                                                     ▼
              Lesson 10 (LR(0) Items) ──► Lesson 11 (Canonical Collection)
                                                     │
                                                     ▼
              Lesson 12 (LALR(1)) ──► Lesson 13 (Conflicts) ──► Lesson 14 (Tables)
                                                                       │
                                                                       ▼
                                    Lesson 15 (Codegen) ──► Lesson 16 (End-to-End)
```

Lessons are strictly ordered. Each lesson's tests assume all prior lessons pass.

---

## Package Map

| Package | Purpose |
|---------|---------|
| `internal/lex` | Token types, lexer for Lemon grammar files |
| `internal/grammar` | Grammar model, Lemon grammar file parser, validation |
| `internal/analysis` | Nullable, FIRST, FOLLOW set computation |
| `internal/lalr` | LR(0) items, LALR(1) states, conflict resolution, tables |
| `internal/codegen` | Go code generation from parse tables |
| `internal/runtime` | Runtime engine used by generated parsers |
| `examples/calc` | Calculator end-to-end example |
| `examples/wsn` | Wirth Syntax Notation end-to-end example |

---

## Running Tests

```bash
# All tests
go test ./...

# Specific lesson's package
go test ./internal/lex/...        # Lessons 01-02
go test ./internal/grammar/...    # Lessons 03-07
go test ./internal/analysis/...   # Lessons 08-09
go test ./internal/lalr/...       # Lessons 10-14
go test ./internal/codegen/...    # Lesson 15
go test ./examples/...            # Lesson 16
```

---

## Conventions

- **Terminal symbols**: UPPER_CASE (e.g., `PLUS`, `INTEGER`, `LPAREN`)
- **Nonterminal symbols**: lower_case (e.g., `expr`, `term`, `program`)
- **Token type constants**: Go `int` constants in `internal/lex/token.go`
- **Error handling**: Functions return `error` or collect diagnostics in a slice; no `panic`
- **Test file naming**: `*_test.go` in the same package
