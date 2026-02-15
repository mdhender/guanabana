# Lesson 13: Conflict Detection and Resolution

**Goal**: Detect shift/reduce and reduce/reduce conflicts in the LALR(1)
automaton, and resolve conflicts using precedence and associativity rules
(Lemon-style).

## What you will build in this lesson

- Conflict detection: identify states where shift and reduce (or two reduces)
  compete on the same lookahead terminal
- Shift/reduce resolution using precedence and associativity
- Reduce/reduce resolution: prefer the rule with lower index (earlier in
  grammar)
- A `Conflict` struct for reporting unresolved conflicts
- A `DetectConflicts` function and a `ResolveConflicts` function
- Conflict summary reporting

## Key concepts

**Shift/reduce conflict**: In a state, the parser can both shift a terminal
(continue reading) and reduce by a rule (apply the rule). This happens when:
- There's an item `A → α • t β` (shift on t)
- There's a reduce item `B → γ •` with t in its lookahead (reduce on t)

**Reduce/reduce conflict**: Two different reduce items in the same state have
overlapping lookahead sets.

**Resolution by precedence** (shift/reduce only):
- Compare the precedence of the terminal (shift side) with the precedence of
  the rule (reduce side).
- Higher precedence wins.
- Equal precedence: use associativity:
  - Left-associative → reduce
  - Right-associative → shift
  - Non-associative → error (neither shift nor reduce)

**Resolution by rule order** (reduce/reduce):
- Lemon resolves reduce/reduce conflicts by preferring the rule that appears
  earlier in the grammar file (lower index). A warning is emitted.

**Unresolved conflicts**: If neither precedence nor rule order can resolve
a conflict, it is reported as an unresolved conflict. The default is to shift
(for shift/reduce) or use the first rule (for reduce/reduce), with warnings.

## Inputs / Outputs

### Files to create

| File | Purpose |
|------|---------|
| `internal/lalr/conflict.go` | Conflict types, detection, resolution |
| `internal/lalr/conflict_test.go` | Conflict tests |

### Types defined

```go
type ConflictKind int

const (
    ShiftReduce  ConflictKind = iota
    ReduceReduce
)

type Conflict struct {
    Kind       ConflictKind
    StateID    int
    Terminal   *grammar.Symbol // the conflicting lookahead terminal
    ShiftItem  *LAItem         // for shift/reduce: the shift item
    ReduceItem *LAItem         // the reduce item
    ReduceItem2 *LAItem        // for reduce/reduce: the second reduce item
    Resolution string          // "shift", "reduce", "error", or ""
    Resolved   bool            // true if precedence/associativity resolved it
}

type ConflictReport struct {
    Conflicts  []Conflict
    Resolved   int
    Unresolved int
}

func DetectAndResolveConflicts(
    automaton *Automaton,
    g *grammar.Grammar,
) *ConflictReport
```

## Algorithm sketch

```
function DetectAndResolveConflicts(automaton, grammar):
    report = new ConflictReport

    for each state S in automaton:
        // Build a map: terminal → list of actions
        actions = map[terminalID][]Action

        for each item in S.LAItems:
            if item is not a reduce item:
                // It can shift on the symbol after the dot
                sym = item.SymbolAfterDot()
                if sym != nil and sym.Kind == Terminal:
                    actions[sym.ID].append(Shift{item})
            else:
                // It's a reduce item — add reduce action for each lookahead
                for each t in item.Lookahead:
                    actions[t].append(Reduce{item})

        // Check for conflicts
        for each terminal t, actionList in actions:
            if len(actionList) <= 1:
                continue  // no conflict

            shifts = filter(actionList, isShift)
            reduces = filter(actionList, isReduce)

            if len(shifts) > 0 and len(reduces) > 0:
                for each reduce in reduces:
                    conflict = resolveShiftReduce(t, shifts[0], reduce, grammar)
                    report.Conflicts.append(conflict)

            if len(reduces) > 1:
                for i, j pairs in reduces:
                    conflict = resolveReduceReduce(t, reduces[i], reduces[j], grammar)
                    report.Conflicts.append(conflict)

    return report

function resolveShiftReduce(terminal, shiftItem, reduceItem, grammar):
    rule = grammar.Rules[reduceItem.RuleIndex]
    termPrec = terminal.Precedence
    rulePrec = rule.Precedence

    if termPrec.Level == 0 or rulePrec.Level == 0:
        return Conflict{resolved: false}  // no precedence info

    if rulePrec.Level > termPrec.Level:
        return Conflict{resolved: true, resolution: "reduce"}
    if termPrec.Level > rulePrec.Level:
        return Conflict{resolved: true, resolution: "shift"}

    // Equal precedence — use associativity
    switch termPrec.Assoc:
        case Left:  return {resolved: true, resolution: "reduce"}
        case Right: return {resolved: true, resolution: "shift"}
        case Nonassoc: return {resolved: true, resolution: "error"}

function resolveReduceReduce(terminal, item1, item2, grammar):
    // Prefer earlier rule (lower index)
    winner = min(item1.RuleIndex, item2.RuleIndex)
    return Conflict{resolved: true, resolution: "rule " + winner}
    // Always emit a warning for reduce/reduce
```

## Repository changes

| Action | File |
|--------|------|
| Create | `internal/lalr/conflict.go` |
| Create | `internal/lalr/conflict_test.go` |

## Unit tests (pass/fail gate)

### Test 1: No conflicts in unambiguous grammar

```go
func TestNoConflicts(t *testing.T) {
    g := buildAugmentedArith() // expr → expr PLUS term | term; term → NUM
    // This grammar is unambiguous, no conflicts
    nullable := analysis.ComputeNullable(g)
    first := analysis.ComputeFirst(g, nullable)
    automaton := BuildCanonical(g)
    ComputeLookaheads(automaton, g, nullable, first)
    report := DetectAndResolveConflicts(automaton, g)

    if report.Unresolved > 0 {
        t.Errorf("expected no unresolved conflicts, got %d", report.Unresolved)
    }
}
```

### Test 2: Shift/reduce conflict with precedence resolution

```go
func TestShiftReduceWithPrecedence(t *testing.T) {
    // Ambiguous grammar resolved by precedence:
    // %left PLUS.
    // %left TIMES.
    // expr → expr PLUS expr | expr TIMES expr | NUM
    g := grammar.NewGrammar()
    g.AddTerminal("PLUS")
    g.AddTerminal("TIMES")
    g.AddTerminal("NUM")
    g.AddNonterminal("expr")
    g.AddRule("expr", []string{"expr", "PLUS", "expr"}, "")
    g.AddRule("expr", []string{"expr", "TIMES", "expr"}, "")
    g.AddRule("expr", []string{"NUM"}, "")

    // Set precedence
    plus, _ := g.Symbols.Lookup("PLUS")
    plus.Precedence = grammar.Precedence{Level: 1, Assoc: grammar.AssocLeft}
    times, _ := g.Symbols.Lookup("TIMES")
    times.Precedence = grammar.Precedence{Level: 2, Assoc: grammar.AssocLeft}
    // Assign rule precedence from rightmost terminal
    g.Rules[0].Precedence = plus.Precedence   // expr → expr PLUS expr
    g.Rules[1].Precedence = times.Precedence   // expr → expr TIMES expr

    g.Finalize()

    nullable := analysis.ComputeNullable(g)
    first := analysis.ComputeFirst(g, nullable)
    automaton := BuildCanonical(g)
    ComputeLookaheads(automaton, g, nullable, first)
    report := DetectAndResolveConflicts(automaton, g)

    if report.Unresolved > 0 {
        t.Errorf("expected all conflicts resolved, got %d unresolved", report.Unresolved)
    }
    if len(report.Conflicts) == 0 {
        t.Error("expected conflicts to be detected (and resolved)")
    }
    // All should be resolved
    for _, c := range report.Conflicts {
        if !c.Resolved {
            t.Errorf("unresolved conflict in state %d on %v", c.StateID, c.Terminal)
        }
    }
}
```

### Test 3: Reduce/reduce conflict

```go
func TestReduceReduceConflict(t *testing.T) {
    // Grammar with reduce/reduce conflict:
    // S → A | B
    // A → x
    // B → x
    g := grammar.NewGrammar()
    g.AddTerminal("x")
    g.AddNonterminal("S")
    g.AddNonterminal("A")
    g.AddNonterminal("B")
    g.AddRule("S", []string{"A"}, "")
    g.AddRule("S", []string{"B"}, "")
    g.AddRule("A", []string{"x"}, "")
    g.AddRule("B", []string{"x"}, "")
    g.Finalize()

    nullable := analysis.ComputeNullable(g)
    first := analysis.ComputeFirst(g, nullable)
    automaton := BuildCanonical(g)
    ComputeLookaheads(automaton, g, nullable, first)
    report := DetectAndResolveConflicts(automaton, g)

    // Should have a reduce/reduce conflict
    hasRR := false
    for _, c := range report.Conflicts {
        if c.Kind == ReduceReduce {
            hasRR = true
        }
    }
    if !hasRR {
        t.Error("expected reduce/reduce conflict")
    }
}
```

### Test 4: Left associativity resolves to reduce

```go
func TestLeftAssocReduces(t *testing.T) {
    // expr → expr PLUS expr | NUM, %left PLUS
    g := grammar.NewGrammar()
    g.AddTerminal("PLUS")
    g.AddTerminal("NUM")
    g.AddNonterminal("expr")
    g.AddRule("expr", []string{"expr", "PLUS", "expr"}, "")
    g.AddRule("expr", []string{"NUM"}, "")

    plus, _ := g.Symbols.Lookup("PLUS")
    plus.Precedence = grammar.Precedence{Level: 1, Assoc: grammar.AssocLeft}
    g.Rules[0].Precedence = plus.Precedence

    g.Finalize()

    nullable := analysis.ComputeNullable(g)
    first := analysis.ComputeFirst(g, nullable)
    automaton := BuildCanonical(g)
    ComputeLookaheads(automaton, g, nullable, first)
    report := DetectAndResolveConflicts(automaton, g)

    // Find the shift/reduce conflict on PLUS — should resolve to reduce
    for _, c := range report.Conflicts {
        if c.Kind == ShiftReduce && c.Terminal == plus {
            if c.Resolution != "reduce" {
                t.Errorf("expected 'reduce' for left-assoc PLUS, got %q", c.Resolution)
            }
        }
    }
}
```

### Test 5: Right associativity resolves to shift

```go
func TestRightAssocShifts(t *testing.T) {
    // expr → expr EXP expr | NUM, %right EXP
    g := grammar.NewGrammar()
    g.AddTerminal("EXP")
    g.AddTerminal("NUM")
    g.AddNonterminal("expr")
    g.AddRule("expr", []string{"expr", "EXP", "expr"}, "")
    g.AddRule("expr", []string{"NUM"}, "")

    exp, _ := g.Symbols.Lookup("EXP")
    exp.Precedence = grammar.Precedence{Level: 1, Assoc: grammar.AssocRight}
    g.Rules[0].Precedence = exp.Precedence

    g.Finalize()

    nullable := analysis.ComputeNullable(g)
    first := analysis.ComputeFirst(g, nullable)
    automaton := BuildCanonical(g)
    ComputeLookaheads(automaton, g, nullable, first)
    report := DetectAndResolveConflicts(automaton, g)

    for _, c := range report.Conflicts {
        if c.Kind == ShiftReduce && c.Terminal == exp {
            if c.Resolution != "shift" {
                t.Errorf("expected 'shift' for right-assoc EXP, got %q", c.Resolution)
            }
        }
    }
}
```

## Common mistakes

1. **Not checking all pairs of actions**: In a single state on a single
   terminal, there can be one shift and multiple reduces. Check all
   combinations.

2. **Applying precedence to reduce/reduce**: Precedence only applies to
   shift/reduce conflicts. Reduce/reduce conflicts use rule ordering.

3. **Wrong direction**: Left-associative means *reduce* (not shift).
   `a + b + c` should parse as `(a + b) + c`, which requires reducing
   `a + b` before shifting the second `+`.

4. **Forgetting nonassoc**: `%nonassoc` means the conflict is an error (the
   syntax is invalid), not that one side wins.

5. **Not reporting resolved conflicts**: Even resolved conflicts should be
   logged/reported. The user needs to see what decisions the generator made.

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-13-conflict-resolution.md and
implement exactly what it describes.

Scope:
- Create internal/lalr/conflict.go with Conflict, ConflictKind, ConflictReport
  types, and DetectAndResolveConflicts function.
- Implement shift/reduce resolution via precedence and associativity.
- Implement reduce/reduce resolution via rule index ordering.
- Create internal/lalr/conflict_test.go with all five tests.
- Report all conflicts (resolved and unresolved) in the ConflictReport.
- Use only the Go standard library.
- Keep the code straightforward.
- Leave TODO: "// TODO(lesson-14): generate parse tables from resolved actions"

Run `go test ./internal/lalr/...` and ensure all tests pass.
```

## Checkpoint questions

1. What is the difference between a shift/reduce and reduce/reduce conflict?
2. When left-associativity is declared, which action wins in a shift/reduce
   conflict: shift or reduce? Why?
3. Why does Lemon resolve reduce/reduce conflicts by rule order?
4. What happens when `%nonassoc` is used and a conflict arises on that token?
5. Can a state have both a shift/reduce and a reduce/reduce conflict
   simultaneously?
