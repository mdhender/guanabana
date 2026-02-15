# Lesson 14: Parse Table Construction

**Goal**: Build the ACTION and GOTO tables from the LALR(1) automaton with
resolved conflicts. These tables drive the generated parser.

## What you will build in this lesson

- An `ActionTable` mapping (state, terminal) → action (shift/reduce/accept/error)
- A `GotoTable` mapping (state, nonterminal) → state
- An `Action` type representing shift, reduce, accept, and error actions
- A `BuildTables` function
- Table validation: determinism and completeness checks
- Unit tests verifying table entries for the arithmetic grammar

## Key concepts

**ACTION table**: For each parser state and each terminal, the ACTION table
says what to do:
- **Shift s**: Push the terminal and go to state s
- **Reduce r**: Apply rule r (pop RHS symbols, push LHS, then consult GOTO)
- **Accept**: Input is valid, parsing complete (only for `$accept → S •` on `$`)
- **Error**: No valid action — syntax error

**GOTO table**: For each parser state and each nonterminal, the GOTO table
says which state to go to after a reduce pushes the nonterminal. This comes
directly from the automaton's transitions on nonterminals.

**Table construction from the automaton**:
- For each transition `(state, terminal) → target_state`: ACTION[state, terminal] = Shift(target_state)
- For each reduce item in a state with lookahead set: ACTION[state, lookahead_terminal] = Reduce(rule)
- For `$accept → S •` in state with `$` lookahead: ACTION[state, $] = Accept
- For each transition `(state, nonterminal) → target_state`: GOTO[state, nonterminal] = target_state

After conflict resolution, each (state, terminal) pair should have at most
one action.

**Determinism**: The table must be deterministic — no cell has two different
actions. Conflict resolution (Lesson 13) should have eliminated all conflicts,
or flagged them.

## Inputs / Outputs

### Files to create

| File | Purpose |
|------|---------|
| `internal/lalr/table.go` | Table types, BuildTables function |
| `internal/lalr/table_test.go` | Table construction tests |

### Types defined

```go
type ActionKind int

const (
    ActionError  ActionKind = iota
    ActionShift
    ActionReduce
    ActionAccept
)

type Action struct {
    Kind      ActionKind
    State     int // for Shift: target state ID
    RuleIndex int // for Reduce: rule index
}

func (a Action) String() string

type ParseTable struct {
    NumStates    int
    NumTerminals int
    NumNonterminals int
    Action       [][]Action         // [stateID][terminalID] → Action
    Goto         [][]int            // [stateID][nonterminalID] → stateID (-1 = error)
    Rules        []*grammar.Rule    // rule table for reduce actions
}

func BuildTables(
    automaton *Automaton,
    g *grammar.Grammar,
    conflicts *ConflictReport,
) *ParseTable
```

## Algorithm sketch

```
function BuildTables(automaton, grammar, conflicts):
    numStates = len(automaton.States)
    numTerminals = count of terminal symbols
    numNonterminals = count of nonterminal symbols

    // Initialize tables with default error/no-goto
    action = 2D array [numStates][numTerminals], default ActionError
    goto = 2D array [numStates][numNonterminals], default -1

    // Fill from transitions (shift actions and goto entries)
    for each transition (from, symbol, to) in automaton:
        if symbol.Kind == Terminal:
            action[from][symbol.ID] = Action{Shift, State: to}
        else:
            goto[from][symbol.nonterminalIndex] = to

    // Fill from reduce items (applying conflict resolutions)
    for each state S in automaton:
        for each LAItem in S:
            if not item.IsReduce(grammar):
                continue

            rule = grammar.Rules[item.RuleIndex]

            if rule.LHS.Name == "$accept":
                // Accept action on $
                action[S.ID][grammar.EOF.ID] = Action{Accept}
                continue

            for each t in item.Lookahead:
                existing = action[S.ID][t]
                newAction = Action{Reduce, RuleIndex: item.RuleIndex}

                if existing.Kind == ActionError:
                    action[S.ID][t] = newAction
                else:
                    // Conflict — check resolution from ConflictReport
                    resolved = lookupResolution(conflicts, S.ID, t)
                    if resolved == "shift":
                        // keep existing shift
                    else if resolved == "reduce":
                        action[S.ID][t] = newAction
                    else if resolved == "error":
                        action[S.ID][t] = Action{ActionError}
                    else:
                        // Default: shift wins for s/r, first rule for r/r
                        // (already handled by resolution)

    return ParseTable{action, goto, grammar.Rules, ...}
```

**Invariants**:
- Every cell in the ACTION table has exactly one action after construction.
- GOTO entries for unreachable (state, nonterminal) pairs are -1.
- The accept action appears exactly once in the table.
- Table dimensions are deterministic and reproducible.

## Repository changes

| Action | File |
|--------|------|
| Create | `internal/lalr/table.go` |
| Create | `internal/lalr/table_test.go` |

## Unit tests (pass/fail gate)

### Test grammar (augmented arithmetic):
```
Rule 0: $accept → expr
Rule 1: expr → expr PLUS term
Rule 2: expr → term
Rule 3: term → NUM
```

### Test 1: Table dimensions

```go
func TestTableDimensions(t *testing.T) {
    g := buildAugmentedArith()
    table := buildFullTable(g) // helper that runs the full pipeline

    if table.NumStates != 6 {
        t.Errorf("NumStates = %d, want 6", table.NumStates)
    }
    // Terminals: $, PLUS, NUM = 3
    if table.NumTerminals < 3 {
        t.Errorf("NumTerminals = %d, want >= 3", table.NumTerminals)
    }
}
```

### Test 2: Accept action exists

```go
func TestAcceptAction(t *testing.T) {
    g := buildAugmentedArith()
    table := buildFullTable(g)

    // Find the state with $accept → expr • and check ACTION[state, $] = Accept
    found := false
    for s := 0; s < table.NumStates; s++ {
        a := table.Action[s][g.EOF.ID]
        if a.Kind == ActionAccept {
            found = true
        }
    }
    if !found {
        t.Error("no Accept action found in table")
    }
}
```

### Test 3: Shift on NUM in state 0

```go
func TestShiftNUMState0(t *testing.T) {
    g := buildAugmentedArith()
    table := buildFullTable(g)

    num, _ := g.Symbols.Lookup("NUM")
    a := table.Action[0][num.ID]
    if a.Kind != ActionShift {
        t.Errorf("ACTION[0, NUM] = %v, want Shift", a)
    }
}
```

### Test 4: Reduce in correct state

```go
func TestReduceInState(t *testing.T) {
    g := buildAugmentedArith()
    table := buildFullTable(g)

    // Find state where term → NUM • — should have reduce action on PLUS and $
    num, _ := g.Symbols.Lookup("NUM")
    plus, _ := g.Symbols.Lookup("PLUS")

    // The state reached by shifting NUM from state 0
    shiftAction := table.Action[0][num.ID]
    targetState := shiftAction.State

    // In the target state, ACTION on PLUS and $ should be Reduce
    a := table.Action[targetState][plus.ID]
    if a.Kind != ActionReduce {
        t.Errorf("ACTION[%d, PLUS] = %v, want Reduce", targetState, a)
    }
    a = table.Action[targetState][g.EOF.ID]
    if a.Kind != ActionReduce {
        t.Errorf("ACTION[%d, $] = %v, want Reduce", targetState, a)
    }
}
```

### Test 5: GOTO table entries

```go
func TestGotoEntries(t *testing.T) {
    g := buildAugmentedArith()
    table := buildFullTable(g)

    expr, _ := g.Symbols.Lookup("expr")
    term, _ := g.Symbols.Lookup("term")

    // From state 0, GOTO on expr and term should be valid states
    exprGoto := table.Goto[0][expr.ID] // adjust indexing as needed
    termGoto := table.Goto[0][term.ID]

    if exprGoto < 0 {
        t.Error("GOTO[0, expr] should be a valid state")
    }
    if termGoto < 0 {
        t.Error("GOTO[0, term] should be a valid state")
    }
}
```

### Test 6: Table is deterministic

```go
func TestTableDeterministic(t *testing.T) {
    g := buildAugmentedArith()
    table := buildFullTable(g)

    // Every ACTION cell should have at most one non-error action
    // (This is guaranteed by conflict resolution, but verify)
    for s := 0; s < table.NumStates; s++ {
        for term := 0; term < table.NumTerminals; term++ {
            a := table.Action[s][term]
            if a.Kind < ActionError || a.Kind > ActionAccept {
                t.Errorf("invalid action kind %d at [%d][%d]", a.Kind, s, term)
            }
        }
    }
}
```

### Helper

```go
func buildFullTable(g *grammar.Grammar) *ParseTable {
    nullable := analysis.ComputeNullable(g)
    first := analysis.ComputeFirst(g, nullable)
    automaton := BuildCanonical(g)
    ComputeLookaheads(automaton, g, nullable, first)
    conflicts := DetectAndResolveConflicts(automaton, g)
    return BuildTables(automaton, g, conflicts)
}
```

## Common mistakes

1. **Double-filling cells without checking resolution**: When a reduce action
   overlaps with a shift action, you must consult the conflict report. Don't
   just overwrite.

2. **Wrong GOTO indexing**: GOTO is indexed by nonterminal, but nonterminal
   IDs may not start at 0. Either offset the index or use a mapping.

3. **Forgetting the Accept action**: The accept is a special reduce for the
   augmented start rule on `$`. Don't treat it as a regular reduce.

4. **Tables not matching state count**: If table dimensions don't match the
   automaton's state count, something went wrong in BuildCanonical.

5. **Non-deterministic table filling order**: If you fill tables in hash-map
   iteration order, results may vary between runs. Always iterate in a
   deterministic order (by state ID, then symbol ID).

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-14-parse-tables.md and implement
exactly what it describes.

Scope:
- Create internal/lalr/table.go with ActionKind, Action, ParseTable types,
  and BuildTables function.
- ACTION table: shift from transitions, reduce from lookaheads, accept for
  augmented start rule on $.
- GOTO table: from nonterminal transitions.
- Apply conflict resolutions from the ConflictReport.
- Create internal/lalr/table_test.go with all six tests and the buildFullTable
  helper.
- Use only the Go standard library.
- Keep the table representation simple (2D slices indexed by int).
- Leave TODO: "// TODO(lesson-15): code generation from tables"

Run `go test ./internal/lalr/...` and ensure all tests pass.
```

## Checkpoint questions

1. What are the four possible actions in the ACTION table?
2. Why is the GOTO table only indexed by nonterminals?
3. How does the Accept action differ from a regular Reduce?
4. What does it mean if an ACTION cell contains ActionError?
5. Why must table construction be deterministic?
