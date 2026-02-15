# Guanabana

An LALR(1) parser generator in Go, inspired by [SQLite Lemon](https://www.sqlite.org/lemon.html).

Guanabana is a learning project and working tool. It reads Lemon-style grammar
files and generates Go parsers that are driven by the lexer (the lexer calls the
parser, not the other way around).

## Status

Under active development, following the structured curriculum in `docs/COURSE.md`.

## Quick Start

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/grammar/...
go test ./internal/analysis/...

# Build the CLI
go build -o guanabana ./cmd/guanabana

# Generate a parser from a grammar file
./guanabana build -g examples/calc/calc.y -o examples/calc/parser.go
```

## Project Layout

```
guanabana/
├── README.md                  # This file
├── GUANABANA.md               # Project philosophy
├── docs/
│   ├── COURSE.md              # Course overview and lesson map
│   ├── lemon.md               # Lemon reference documentation
│   └── lesson-XX-*.md         # Individual lesson files
├── cmd/
│   └── guanabana/             # CLI entry point
├── internal/
│   ├── lex/                   # Lexer for Lemon grammar files
│   │   ├── token.go           # Token types and Token struct
│   │   ├── lexer.go           # Lexer implementation
│   │   └── lexer_test.go      # Lexer validation tests
│   ├── grammar/               # Grammar model and parser
│   │   ├── symbol.go          # Symbol table (terminals + nonterminals)
│   │   ├── rule.go            # Production rules
│   │   ├── grammar.go         # Grammar struct and builder
│   │   ├── parser.go          # Lemon grammar file parser
│   │   └── *_test.go          # Grammar tests
│   ├── analysis/              # FIRST, FOLLOW, nullable
│   │   ├── nullable.go        # Nullable set computation
│   │   ├── first.go           # FIRST set computation
│   │   ├── follow.go          # FOLLOW set computation
│   │   └── *_test.go          # Analysis tests
│   ├── lalr/                  # LR(0) items, LALR(1) states, tables
│   │   ├── item.go            # LR(0) item representation
│   │   ├── closure.go         # Closure algorithm
│   │   ├── goto.go            # GOTO function
│   │   ├── collection.go      # Canonical LR(0) collection
│   │   ├── lookahead.go       # LALR(1) lookahead propagation
│   │   ├── conflict.go        # Conflict detection and resolution
│   │   ├── table.go           # ACTION and GOTO tables
│   │   └── *_test.go          # LALR tests
│   ├── codegen/               # Code generation
│   │   ├── generate.go        # Parser code generator
│   │   ├── template.go        # Go code templates
│   │   └── *_test.go          # Codegen tests
│   └── runtime/               # Runtime support for generated parsers
│       ├── parser.go          # Parser state and shift/reduce engine
│       └── parser_test.go     # Runtime tests
├── examples/
│   ├── calc/                  # Calculator grammar + lexer
│   │   ├── calc.y             # Calculator grammar file
│   │   ├── lexer.go           # Hand-written calculator lexer
│   │   ├── parser.go          # Generated parser (output)
│   │   └── calc_test.go       # End-to-end tests
│   └── wsn/                   # Wirth Syntax Notation example
│       ├── wsn.y              # WSN grammar file
│       ├── lexer.go           # Hand-written WSN lexer
│       ├── parser.go          # Generated parser (output)
│       └── wsn_test.go        # End-to-end tests
└── go.mod
```

## Curriculum

The complete course is documented in `docs/COURSE.md` with 16 lessons that
incrementally build the parser generator from token definitions through
end-to-end examples.

To work through a lesson:

```
"Let's start the lesson as specified in docs/lesson-08-nullable-first.md."
```

Each lesson includes unit tests that serve as the pass/fail gate.

## Design Decisions

- **Lexer drives the parser**: Following Lemon's architecture, the lexer calls
  `Parse(token, value)` for each token. The parser does not pull tokens.
- **Standard library only**: The parser generator packages use only Go's
  standard library. The CLI uses cobra for convenience.
- **Clarity over speed**: The implementation favors readability and
  correctness. No micro-optimizations.
- **Structured diagnostics**: Errors are collected, not panicked. Every
  diagnostic includes location information.

## License

See [LICENSE](LICENSE).
