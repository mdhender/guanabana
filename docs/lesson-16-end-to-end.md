# Lesson 16: End-to-End Examples

**Goal**: Prove the parser generator works by building two complete examples
end-to-end: a calculator and a Wirth Syntax Notation (WSN) parser. Each
example goes from grammar file through code generation to parsing real input.

## What you will build in this lesson

- A calculator grammar (`examples/calc/calc.y`) with arithmetic expressions
- A hand-written calculator lexer (`examples/calc/lexer.go`)
- A generated calculator parser that evaluates expressions
- A WSN grammar (`examples/wsn/wsn.y`) for a small DSL
- A hand-written WSN lexer (`examples/wsn/lexer.go`)
- A generated WSN parser that builds an AST
- End-to-end tests for both

## Key concepts

**Full pipeline**: For each example, we exercise the complete pipeline:
1. Write a `.y` grammar file
2. Run the parser generator (tokenize → parse → validate → analyze → LALR → tables → codegen)
3. The generated parser is called by a hand-written lexer
4. Tests verify correct parsing and evaluation/AST construction

**Calculator example**: A classic parser generator demo. The grammar handles
`+`, `-`, `*`, `/`, parentheses, and integer literals. Action code performs
evaluation so the parser computes results directly.

**WSN example**: Wirth Syntax Notation describes grammars using `=`, `.`, `|`,
`(`, `)`, `[`, `]`, `{`, `}`, and quoted strings. The parser builds a simple
AST. This demonstrates the generator works for a non-trivial, non-arithmetic
grammar.

**Test strategy**: Tests in this lesson use `go test` build tags or `TestMain`
to run the full pipeline, then use the generated parser to parse sample inputs.
Since we generate Go code, we can either:
1. Pre-generate the parser and commit it (simpler)
2. Generate during test setup (more thorough)

We use approach 1 for simplicity: generate the parser as part of the lesson
implementation, then test it.

## Inputs / Outputs

### Files to create

| File | Purpose |
|------|---------|
| `examples/calc/calc.y` | Calculator grammar |
| `examples/calc/lexer.go` | Hand-written calculator token lexer |
| `examples/calc/parser.go` | Generated parser (output of generator) |
| `examples/calc/calc_test.go` | End-to-end calculator tests |
| `examples/wsn/wsn.y` | WSN grammar |
| `examples/wsn/lexer.go` | Hand-written WSN lexer |
| `examples/wsn/parser.go` | Generated parser (output of generator) |
| `examples/wsn/wsn_test.go` | End-to-end WSN tests |

### Calculator grammar (`calc.y`)

```
// Calculator grammar for Guanabana

%token_type { int }
%start_symbol program.

%left PLUS MINUS.
%left TIMES DIVIDE.

program ::= expr(A). { result = A }

expr(A) ::= expr(B) PLUS expr(C).  { A = B + C }
expr(A) ::= expr(B) MINUS expr(C). { A = B - C }
expr(A) ::= expr(B) TIMES expr(C). { A = B * C }
expr(A) ::= expr(B) DIVIDE expr(C). {
    if C == 0 { A = 0 } else { A = B / C }
}
expr(A) ::= LPAREN expr(B) RPAREN. { A = B }
expr(A) ::= INTEGER(B).            { A = B }
```

### Calculator lexer sketch

```go
// examples/calc/lexer.go
package calc

type CalcLexer struct {
    input  string
    pos    int
    parser *Parser
}

func NewCalcLexer(input string) *CalcLexer

// Run scans the input and feeds tokens to the parser.
func (l *CalcLexer) Run() (int, error) {
    p := NewParser()
    for {
        tokenType, value := l.nextToken()
        p.Parse(tokenType, value)
        if tokenType == 0 { // EOF
            break
        }
    }
    return p.ParseComplete()
}

func (l *CalcLexer) nextToken() (int, interface{}) {
    // Skip whitespace
    // If digit: scan integer, return (INTEGER, value)
    // If '+': return (PLUS, nil)
    // If '-': return (MINUS, nil)
    // If '*': return (TIMES, nil)
    // If '/': return (DIVIDE, nil)
    // If '(': return (LPAREN, nil)
    // If ')': return (RPAREN, nil)
    // At end: return (0, nil) // EOF
}
```

### WSN grammar (`wsn.y`)

```
// Wirth Syntax Notation grammar for Guanabana

%token_type { string }
%start_symbol syntax.

syntax ::= production_list.

production_list ::= production_list production.
production_list ::= production.

production ::= IDENT EQUALS expression DOT.

expression ::= term_list.

term_list ::= term_list PIPE term.
term_list ::= term.

term ::= factor_list.

factor_list ::= factor_list factor.
factor_list ::= factor.

factor ::= IDENT.
factor ::= STRING.
factor ::= LPAREN expression RPAREN.
factor ::= LBRACKET expression RBRACKET.
factor ::= LBRACE expression RBRACE.
```

## Algorithm sketch

### End-to-end pipeline (for test setup)

```
function generateAndTest(grammarFile, outputFile):
    // 1. Read grammar
    src = readFile(grammarFile)

    // 2. Tokenize
    tokens = lex.Tokenize(grammarFile, src)

    // 3. Parse grammar
    grammar, diags = grammar.ParseGrammar(tokens)

    // 4. Finalize
    grammar.Finalize()

    // 5. Compute sets
    nullable = analysis.ComputeNullable(grammar)
    first = analysis.ComputeFirst(grammar, nullable)

    // 6. Build automaton
    automaton = lalr.BuildCanonical(grammar)
    lalr.ComputeLookaheads(automaton, grammar, nullable, first)
    conflicts = lalr.DetectAndResolveConflicts(automaton, grammar)

    // 7. Build tables
    table = lalr.BuildTables(automaton, grammar, conflicts)

    // 8. Generate code
    code = codegen.Generate(config)
    writeFile(outputFile, code)
```

## Repository changes

| Action | File |
|--------|------|
| Create | `examples/calc/calc.y` |
| Create | `examples/calc/lexer.go` |
| Create | `examples/calc/parser.go` (generated) |
| Create | `examples/calc/calc_test.go` |
| Create | `examples/wsn/wsn.y` |
| Create | `examples/wsn/lexer.go` |
| Create | `examples/wsn/parser.go` (generated) |
| Create | `examples/wsn/wsn_test.go` |

## Unit tests (pass/fail gate)

### Calculator tests

```go
// examples/calc/calc_test.go
package calc

import "testing"

func TestCalcSimpleAddition(t *testing.T) {
    result, err := Evaluate("2 + 3")
    if err != nil {
        t.Fatal(err)
    }
    if result != 5 {
        t.Errorf("2 + 3 = %d, want 5", result)
    }
}

func TestCalcPrecedence(t *testing.T) {
    result, err := Evaluate("2 + 3 * 4")
    if err != nil {
        t.Fatal(err)
    }
    if result != 14 {
        t.Errorf("2 + 3 * 4 = %d, want 14", result)
    }
}

func TestCalcParentheses(t *testing.T) {
    result, err := Evaluate("(2 + 3) * 4")
    if err != nil {
        t.Fatal(err)
    }
    if result != 20 {
        t.Errorf("(2 + 3) * 4 = %d, want 20", result)
    }
}

func TestCalcLeftAssociativity(t *testing.T) {
    result, err := Evaluate("10 - 3 - 2")
    if err != nil {
        t.Fatal(err)
    }
    if result != 5 {
        t.Errorf("10 - 3 - 2 = %d, want 5", result)
    }
}

func TestCalcDivision(t *testing.T) {
    result, err := Evaluate("12 / 3")
    if err != nil {
        t.Fatal(err)
    }
    if result != 4 {
        t.Errorf("12 / 3 = %d, want 4", result)
    }
}

func TestCalcComplex(t *testing.T) {
    result, err := Evaluate("(1 + 2) * (3 + 4)")
    if err != nil {
        t.Fatal(err)
    }
    if result != 21 {
        t.Errorf("(1 + 2) * (3 + 4) = %d, want 21", result)
    }
}

func TestCalcSyntaxError(t *testing.T) {
    _, err := Evaluate("2 + + 3")
    if err == nil {
        t.Error("expected syntax error for '2 + + 3'")
    }
}
```

### WSN tests

```go
// examples/wsn/wsn_test.go
package wsn

import "testing"

func TestWSNSimpleProduction(t *testing.T) {
    input := `digit = "0" | "1" | "2" .`
    ast, err := Parse(input)
    if err != nil {
        t.Fatal(err)
    }
    if ast == nil {
        t.Error("expected non-nil AST")
    }
}

func TestWSNMultipleProductions(t *testing.T) {
    input := `
number = digit { digit } .
digit = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" .
`
    ast, err := Parse(input)
    if err != nil {
        t.Fatal(err)
    }
    if ast == nil {
        t.Error("expected non-nil AST")
    }
}

func TestWSNOptionalGroup(t *testing.T) {
    input := `sign = [ "-" | "+" ] .`
    ast, err := Parse(input)
    if err != nil {
        t.Fatal(err)
    }
    if ast == nil {
        t.Error("expected non-nil AST")
    }
}

func TestWSNNestedGroups(t *testing.T) {
    input := `expr = term { ( "+" | "-" ) term } .`
    _, err := Parse(input)
    if err != nil {
        t.Fatal(err)
    }
}

func TestWSNSyntaxError(t *testing.T) {
    input := `bad = .`  // missing expression
    _, err := Parse(input)
    if err == nil {
        t.Error("expected syntax error")
    }
}
```

## Common mistakes

1. **Lexer token types not matching generated constants**: The hand-written
   lexer must use the same token type integers as the generated parser's
   constants. Import them from the generated file or define them consistently.

2. **Forgetting to send EOF**: After all tokens are sent, the lexer must call
   `Parse(0, nil)` to signal end-of-input. Without this, the parser never
   reaches the Accept action.

3. **Action code not matching Go syntax**: The action code in the grammar
   file must be valid Go. Watch for variable scoping — aliases like `A`, `B`,
   `C` need to be mapped to the value stack properly.

4. **Not handling whitespace in the lexer**: The generated parser only sees
   tokens. The lexer must skip whitespace between tokens.

5. **WSN lexer complexity**: The WSN lexer needs to handle quoted strings,
   identifiers, and single-character punctuation (`=`, `.`, `|`, `(`, `)`,
   `[`, `]`, `{`, `}`). Don't over-complicate it.

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-16-end-to-end.md and implement
exactly what it describes.

Scope:
- Create examples/calc/calc.y with the calculator grammar.
- Create examples/calc/lexer.go with a hand-written lexer for calculator
  expressions (integers, +, -, *, /, parentheses).
- Generate examples/calc/parser.go using the Guanabana pipeline (or write
  a test/generate script that does this).
- Create examples/calc/calc_test.go with all seven calculator tests.
- Create examples/wsn/wsn.y with the WSN grammar.
- Create examples/wsn/lexer.go with a hand-written WSN lexer.
- Generate examples/wsn/parser.go.
- Create examples/wsn/wsn_test.go with all five WSN tests.
- Provide an Evaluate(input string) (int, error) function for calc and a
  Parse(input string) (interface{}, error) function for WSN.
- Use only the Go standard library.
- Keep lexers simple and readable.

Run `go test ./examples/...` and ensure all tests pass.
```

## Checkpoint questions

1. Why does the lexer drive the parser instead of the parser pulling tokens?
2. How does precedence affect the calculator's handling of `2 + 3 * 4`?
3. What would happen if the calculator grammar didn't have `%left` directives?
4. In the WSN grammar, why are `LBRACKET expression RBRACKET` and
   `LBRACE expression RBRACE` separate productions?
5. How would you add support for floating-point numbers to the calculator
   example?
