# Lesson 05: Precedence, Associativity, and Directives

**Goal**: Fully process Lemon's precedence and associativity directives
(`%left`, `%right`, `%nonassoc`) and attach precedence information to
terminals and rules so that conflict resolution can use it later.

## What you will build in this lesson

- Full parsing of `%left`, `%right`, and `%nonassoc` directives
- A `Precedence` struct attached to terminals
- Automatic precedence assignment to rules (based on the rightmost terminal)
- Parsing of `%token_type`, `%type`, `%default_type`, `%start_symbol`,
  `%name`, `%extra_argument`, `%include`, `%token_prefix`, and `%code`
  directives into structured records
- Storage of directive data on the Grammar struct

## Key concepts

**Precedence and associativity**: When a parser faces a shift/reduce conflict,
it can use precedence rules to decide. Each terminal can have a precedence
level (higher = binds tighter) and an associativity (left, right, or
nonassoc). A rule's precedence is determined by its rightmost terminal, unless
overridden.

In Lemon, precedence is declared with directives:
```
%left PLUS MINUS.
%left TIMES DIVIDE.
%right EXP.
```
The first `%left` or `%right` line has the lowest precedence. Each subsequent
line has higher precedence. All tokens on the same line have the same level.

**Associativity types**:
- `left`: `a + b + c` parses as `(a + b) + c`
- `right`: `a ^ b ^ c` parses as `a ^ (b ^ c)`
- `nonassoc`: `a == b == c` is a syntax error

**Rule precedence**: By default, a rule's precedence comes from its rightmost
terminal. Lemon allows overriding with `[TOKEN]` syntax after the dot:
```
expr ::= MINUS expr. [UMINUS]
```
We will support this override syntax.

## Inputs / Outputs

### Files to create / modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/grammar/precedence.go` | Create | Precedence type and Assoc enum |
| `internal/grammar/directives.go` | Create | Directive record types |
| `internal/grammar/parser.go` | Modify | Full directive parsing |
| `internal/grammar/grammar.go` | Modify | Add directive fields to Grammar |
| `internal/grammar/rule.go` | Modify | Add Precedence field to Rule |
| `internal/grammar/symbol.go` | Modify | Add Precedence field to Symbol |
| `internal/grammar/precedence_test.go` | Create | Precedence tests |

### Types defined

```go
// internal/grammar/precedence.go

type Assoc int

const (
    AssocNone  Assoc = iota // no associativity assigned
    AssocLeft               // %left
    AssocRight              // %right
    AssocNonassoc           // %nonassoc
)

type Precedence struct {
    Level int   // 0 = unassigned, 1 = lowest, 2, 3, ...
    Assoc Assoc
}
```

```go
// internal/grammar/directives.go

type DirectiveKind int

const (
    DirTokenType DirectiveKind = iota
    DirType
    DirDefaultType
    DirStartSymbol
    DirName
    DirExtraArgument
    DirInclude
    DirTokenPrefix
    DirCode
    DirFallback
    DirWildcard
    DirDestructor
    DirSyntaxError
    DirParseAccept
    DirParseFailure
    DirStackOverflow
)

type Directive struct {
    Kind    DirectiveKind
    Pos     lex.Position
    Symbols []string // for directives that reference symbols
    Code    string   // for directives that have a code block
    Value   string   // for directives with a simple value
}
```

## Algorithm sketch

**Precedence assignment**:

```
function assignPrecedence(grammar):
    level = 0
    for each raw directive in declaration order:
        if directive is %left, %right, or %nonassoc:
            level++
            assoc = directive.assoc
            for each terminal name in directive.symbols:
                terminal = grammar.Symbols.Lookup(name)
                terminal.Precedence = Precedence{Level: level, Assoc: assoc}

    // Assign precedence to rules
    for each rule in grammar.Rules:
        if rule has explicit precedence override [TOKEN]:
            rule.Precedence = TOKEN.Precedence
        else:
            // Use rightmost terminal in RHS
            for i = len(rule.RHS)-1 downto 0:
                if rule.RHS[i].Kind == SymbolTerminal:
                    rule.Precedence = rule.RHS[i].Precedence
                    break
```

**Invariants**:
- Precedence levels are strictly increasing (1, 2, 3, …).
- A terminal with Level=0 has no precedence assigned.
- A rule with Level=0 has no precedence (conflicts involving it cannot be
  resolved by precedence).

## Repository changes

| Action | File |
|--------|------|
| Create | `internal/grammar/precedence.go` |
| Create | `internal/grammar/directives.go` |
| Create | `internal/grammar/precedence_test.go` |
| Modify | `internal/grammar/parser.go` — implement directive parsing |
| Modify | `internal/grammar/grammar.go` — add `Directives []Directive`, `PrecLevel int` |
| Modify | `internal/grammar/rule.go` — add `Precedence Precedence`, `PrecOverride string` |
| Modify | `internal/grammar/symbol.go` — add `Precedence Precedence` to Symbol |

## Unit tests (pass/fail gate)

### Test 1: Precedence levels assigned correctly

```go
func TestPrecedenceLevels(t *testing.T) {
    src := []byte(`
%left PLUS MINUS.
%left TIMES DIVIDE.
%right EXP.
expr ::= expr PLUS expr.
expr ::= expr TIMES expr.
expr ::= NUM.
`)
    tokens, _ := lex.Tokenize("test.y", src)
    g, diags, err := ParseGrammar(tokens)
    if err != nil {
        t.Fatal(err)
    }
    if len(diags) > 0 {
        t.Errorf("diagnostics: %v", diags)
    }

    plus, _ := g.Symbols.Lookup("PLUS")
    times, _ := g.Symbols.Lookup("TIMES")
    exp, _ := g.Symbols.Lookup("EXP")

    if plus.Precedence.Level != 1 {
        t.Errorf("PLUS level = %d, want 1", plus.Precedence.Level)
    }
    if plus.Precedence.Assoc != AssocLeft {
        t.Errorf("PLUS assoc = %v, want Left", plus.Precedence.Assoc)
    }
    if times.Precedence.Level != 2 {
        t.Errorf("TIMES level = %d, want 2", times.Precedence.Level)
    }
    if exp.Precedence.Level != 3 {
        t.Errorf("EXP level = %d, want 3", exp.Precedence.Level)
    }
    if exp.Precedence.Assoc != AssocRight {
        t.Errorf("EXP assoc = %v, want Right", exp.Precedence.Assoc)
    }
}
```

### Test 2: Rule precedence from rightmost terminal

```go
func TestRulePrecedenceRightmostTerminal(t *testing.T) {
    src := []byte(`
%left PLUS.
%left TIMES.
expr ::= expr PLUS expr.
expr ::= expr TIMES expr.
expr ::= NUM.
`)
    tokens, _ := lex.Tokenize("test.y", src)
    g, _, _ := ParseGrammar(tokens)

    // Rule "expr ::= expr PLUS expr" should have PLUS precedence
    if g.Rules[0].Precedence.Level != 1 {
        t.Errorf("rule 0 prec level = %d, want 1", g.Rules[0].Precedence.Level)
    }
    // Rule "expr ::= expr TIMES expr" should have TIMES precedence
    if g.Rules[1].Precedence.Level != 2 {
        t.Errorf("rule 1 prec level = %d, want 2", g.Rules[1].Precedence.Level)
    }
    // Rule "expr ::= NUM" — NUM has no precedence
    if g.Rules[2].Precedence.Level != 0 {
        t.Errorf("rule 2 prec level = %d, want 0", g.Rules[2].Precedence.Level)
    }
}
```

### Test 3: Precedence override with bracket syntax

```go
func TestPrecedenceOverride(t *testing.T) {
    src := []byte(`
%left PLUS.
%left TIMES.
%left UMINUS.
expr ::= MINUS expr. [UMINUS]
expr ::= expr PLUS expr.
expr ::= NUM.
`)
    tokens, _ := lex.Tokenize("test.y", src)
    g, _, _ := ParseGrammar(tokens)

    // Rule "expr ::= MINUS expr. [UMINUS]" should have UMINUS precedence (level 3)
    if g.Rules[0].Precedence.Level != 3 {
        t.Errorf("override rule prec = %d, want 3", g.Rules[0].Precedence.Level)
    }
}
```

### Test 4: Token type directive parsed

```go
func TestTokenTypeDirective(t *testing.T) {
    src := []byte(`
%token_type { Value }
expr ::= NUM.
`)
    tokens, _ := lex.Tokenize("test.y", src)
    g, _, _ := ParseGrammar(tokens)

    found := false
    for _, d := range g.Directives {
        if d.Kind == DirTokenType {
            found = true
            if d.Code == "" {
                t.Error("token_type directive should have code")
            }
        }
    }
    if !found {
        t.Error("token_type directive not found")
    }
}
```

### Test 5: Multiple tokens on same precedence line

```go
func TestSamePrecedenceLine(t *testing.T) {
    src := []byte(`
%left PLUS MINUS.
expr ::= NUM.
`)
    tokens, _ := lex.Tokenize("test.y", src)
    g, _, _ := ParseGrammar(tokens)

    plus, _ := g.Symbols.Lookup("PLUS")
    minus, _ := g.Symbols.Lookup("MINUS")

    if plus.Precedence.Level != minus.Precedence.Level {
        t.Error("PLUS and MINUS should have same precedence level")
    }
    if plus.Precedence.Assoc != minus.Precedence.Assoc {
        t.Error("PLUS and MINUS should have same associativity")
    }
}
```

## Common mistakes

1. **Off-by-one in precedence levels**: The first `%left` line should be
   level 1, not 0. Level 0 means "no precedence assigned."

2. **Not handling the `[TOKEN]` override syntax**: After the dot in a rule,
   Lemon allows `[TOKEN]` to override rule precedence. The lexer emits
   `TOKEN_LBRACKET`, `TOKEN_TERMINAL`, `TOKEN_RBRACKET` for this.

3. **Applying precedence before all directives are parsed**: Parse all
   directives first, then assign precedence. Otherwise, forward-referenced
   tokens won't have their levels set.

4. **Forgetting that rule precedence defaults to the rightmost terminal**:
   Not the leftmost, not the highest — the rightmost terminal in the RHS.

5. **Not storing directives in declaration order**: The order matters for
   precedence levels.

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-05-precedence-directives.md and
implement exactly what it describes.

Scope:
- Create internal/grammar/precedence.go with Assoc enum and Precedence struct.
- Create internal/grammar/directives.go with DirectiveKind and Directive types.
- Modify internal/grammar/parser.go to fully parse precedence directives
  (%left, %right, %nonassoc) and other directives (%token_type, %type,
  %default_type, %start_symbol, %name, %extra_argument, %include,
  %token_prefix, %code, %fallback, %wildcard, %destructor, %syntax_error,
  %parse_accept, %parse_failure, %stack_overflow).
- Support the [TOKEN] precedence override syntax after rule dots.
- Modify Grammar to store Directives slice and assign precedence to symbols
  and rules.
- Modify Rule to include Precedence field and PrecOverride.
- Modify Symbol to include Precedence field.
- Create internal/grammar/precedence_test.go with all five tests.
- Use only the Go standard library.
- Keep code readable.

Run `go test ./internal/grammar/...` and ensure all tests pass.
```

## Checkpoint questions

1. Why is precedence level 0 reserved for "no precedence"?
2. What does `%nonassoc` mean in practice? Give an example.
3. Why does rule precedence default to the *rightmost* terminal, not the
   leftmost?
4. When would you need the `[TOKEN]` precedence override syntax?
5. If two terminals on the same `%left` line have the same precedence, how
   is a conflict between them resolved?
