# Lesson 10: LR(0) Items and Closure

**Goal**: Define the LR(0) item representation and implement the closure
algorithm, which expands a set of items by adding all items reachable through
nonterminal transitions.

## What you will build in this lesson

- An `Item` struct representing an LR(0) item: a rule with a dot position
- An `ItemSet` type (a set of items with deduplication)
- A `Closure(items ItemSet, grammar) ItemSet` function
- Unit tests verifying closure on small item sets

## Key concepts

**LR(0) item**: A production rule with a dot (•) somewhere in the RHS,
indicating how much of the rule has been recognized so far.
- `A → • X Y Z` — nothing recognized yet (initial item)
- `A → X • Y Z` — X has been recognized, expecting Y next
- `A → X Y Z •` — complete item, ready to reduce

An item is uniquely identified by (rule_index, dot_position). The dot position
ranges from 0 (before the first symbol) to len(RHS) (after the last symbol).

**Closure**: Given a set of items, closure adds new items that the parser must
also track. The rule is: if `A → α • B β` is in the set (the dot is before
nonterminal B), then for every rule `B → γ`, add `B → • γ` to the set.

Intuitively: if the parser is about to see a B, it needs to know all the ways
B can start.

**Why closure matters**: Each state in the LR automaton is a closed item set.
The closure ensures the state captures all possible parse positions at that
point.

## Inputs / Outputs

### Files to create

| File | Purpose |
|------|---------|
| `internal/lalr/item.go` | Item struct and ItemSet type |
| `internal/lalr/closure.go` | Closure algorithm |
| `internal/lalr/lalr_test.go` | Tests for items and closure |

### Types defined

```go
package lalr

import "github.com/mdhender/guanabana/internal/grammar"

// Item is an LR(0) item: a rule with a dot position.
type Item struct {
    RuleIndex int // index into Grammar.Rules
    Dot       int // position of the dot (0 = before first symbol)
}

// SymbolAfterDot returns the symbol immediately after the dot, or nil if
// the dot is at the end (reduce item).
func (item Item) SymbolAfterDot(g *grammar.Grammar) *grammar.Symbol

// IsReduce returns true if the dot is at the end of the RHS.
func (item Item) IsReduce(g *grammar.Grammar) bool

// ItemSet is an ordered, deduplicated set of items.
type ItemSet struct {
    Items []Item
}

func NewItemSet(items ...Item) ItemSet
func (s *ItemSet) Add(item Item) bool // returns true if item was new
func (s ItemSet) Contains(item Item) bool
func (s ItemSet) Len() int
func (s ItemSet) Equal(other ItemSet) bool
```

### Functions

```go
// Closure expands the item set by adding initial items for all
// nonterminals that appear after a dot.
func Closure(items ItemSet, g *grammar.Grammar) ItemSet
```

## Algorithm sketch

```
function Closure(items, grammar):
    result = copy of items
    worklist = list(items.Items)

    while worklist not empty:
        item = worklist.pop()
        B = item.SymbolAfterDot(grammar)
        if B == nil or B.Kind == Terminal:
            continue

        // B is a nonterminal — add all B → • γ items
        for each rule r where r.LHS == B:
            newItem = Item{RuleIndex: r.Index, Dot: 0}
            if result.Add(newItem):
                worklist.append(newItem)

    return result
```

**Termination**: Each item (rule_index, dot=0) is added at most once. The
number of possible initial items equals the number of rules. The worklist
shrinks every iteration (pop) and grows by at most the number of new items
added. Since items are never re-added, the loop terminates.

**Invariant**: The returned ItemSet is closed — for every item `A → α • B β`
in the set, all `B → • γ` items are also in the set.

## Repository changes

| Action | File |
|--------|------|
| Create | `internal/lalr/item.go` |
| Create | `internal/lalr/closure.go` |
| Create | `internal/lalr/lalr_test.go` |

## Unit tests (pass/fail gate)

### Grammar for testing (arithmetic):
```
Rule 0: $accept → expr     (augmented)
Rule 1: expr → expr PLUS term
Rule 2: expr → term
Rule 3: term → NUM
```

### Test 1: Single item, no closure needed

```go
func TestClosureSingleTerminal(t *testing.T) {
    g := buildAugmentedArith()
    // Item: term → • NUM (dot before terminal — no closure expansion)
    items := NewItemSet(Item{RuleIndex: 3, Dot: 0})
    closed := Closure(items, g)
    if closed.Len() != 1 {
        t.Errorf("closure size = %d, want 1", closed.Len())
    }
}
```

### Test 2: Closure expands nonterminal

```go
func TestClosureExpandsNonterminal(t *testing.T) {
    g := buildAugmentedArith()
    // Item: $accept → • expr
    // Should add: expr → • expr PLUS term, expr → • term
    // Then: term → • NUM (because expr → • term has dot before term)
    items := NewItemSet(Item{RuleIndex: 0, Dot: 0})
    closed := Closure(items, g)

    // Expected items:
    // $accept → • expr        (original)
    // expr → • expr PLUS term (from closure on expr)
    // expr → • term           (from closure on expr)
    // term → • NUM            (from closure on term)
    if closed.Len() != 4 {
        t.Errorf("closure size = %d, want 4", closed.Len())
    }
    // Verify specific items exist
    if !closed.Contains(Item{RuleIndex: 1, Dot: 0}) {
        t.Error("missing: expr → • expr PLUS term")
    }
    if !closed.Contains(Item{RuleIndex: 2, Dot: 0}) {
        t.Error("missing: expr → • term")
    }
    if !closed.Contains(Item{RuleIndex: 3, Dot: 0}) {
        t.Error("missing: term → • NUM")
    }
}
```

### Test 3: Reduce item (dot at end) does not expand

```go
func TestClosureReduceItem(t *testing.T) {
    g := buildAugmentedArith()
    // term → NUM • (dot at end — reduce item)
    items := NewItemSet(Item{RuleIndex: 3, Dot: 1})
    closed := Closure(items, g)
    if closed.Len() != 1 {
        t.Errorf("closure of reduce item should stay at 1, got %d", closed.Len())
    }
}
```

### Test 4: ItemSet deduplication

```go
func TestItemSetDedup(t *testing.T) {
    s := NewItemSet()
    s.Add(Item{RuleIndex: 1, Dot: 0})
    s.Add(Item{RuleIndex: 1, Dot: 0}) // duplicate
    s.Add(Item{RuleIndex: 2, Dot: 0})
    if s.Len() != 2 {
        t.Errorf("ItemSet.Len() = %d, want 2", s.Len())
    }
}
```

### Test 5: SymbolAfterDot

```go
func TestSymbolAfterDot(t *testing.T) {
    g := buildAugmentedArith()
    // expr → expr • PLUS term (dot before PLUS)
    item := Item{RuleIndex: 1, Dot: 1}
    sym := item.SymbolAfterDot(g)
    if sym == nil || sym.Name != "PLUS" {
        t.Errorf("SymbolAfterDot = %v, want PLUS", sym)
    }
    // term → NUM • (dot at end)
    item2 := Item{RuleIndex: 3, Dot: 1}
    if item2.SymbolAfterDot(g) != nil {
        t.Error("reduce item should return nil SymbolAfterDot")
    }
}
```

### Helper

```go
func buildAugmentedArith() *grammar.Grammar {
    g := grammar.NewGrammar()
    g.AddTerminal("PLUS")
    g.AddTerminal("NUM")
    g.AddNonterminal("expr")
    g.AddNonterminal("term")
    g.AddRule("expr", []string{"expr", "PLUS", "term"}, "")
    g.AddRule("expr", []string{"term"}, "")
    g.AddRule("term", []string{"NUM"}, "")
    g.Finalize() // adds $accept → expr
    return g
}
```

## Common mistakes

1. **Not deduplicating items**: Adding the same (rule, dot) pair twice
   creates infinite loops or bloated state sets. Always check before adding.

2. **Only expanding one level**: Closure is transitive. If `expr → • term`
   causes `term → • NUM` to be added, and `NUM` were a nonterminal, its
   rules would be added too. Use a worklist or queue, not just one pass.

3. **Forgetting the augmented rule**: The initial state always starts with
   `$accept → • S`. Make sure `Finalize()` has been called before building
   items.

4. **Wrong dot position range**: Dot ranges from 0 to `len(RHS)` inclusive.
   Position `len(RHS)` means "dot at the end" (reduce item). Off-by-one
   errors here cause subtle bugs.

5. **Treating ItemSet as unordered**: While the set semantics don't depend
   on order, deterministic ordering (sorted by RuleIndex, then Dot) makes
   debugging and state comparison reliable.

## Codex prompt for this lesson

```
Read the lesson specification in docs/lesson-10-lr0-items-closure.md and
implement exactly what it describes.

Scope:
- Create internal/lalr/item.go with Item struct (RuleIndex, Dot),
  SymbolAfterDot, IsReduce methods, and ItemSet type with Add, Contains,
  Len, Equal, and NewItemSet.
- Create internal/lalr/closure.go with the Closure function using a
  worklist algorithm.
- Create internal/lalr/lalr_test.go with all five tests and the
  buildAugmentedArith helper.
- Items should be sorted deterministically within ItemSet (by RuleIndex,
  then Dot).
- Use only the Go standard library.
- Keep code simple and readable.
- Leave TODO: "// TODO(lesson-11): GOTO function and canonical collection"

Run `go test ./internal/lalr/...` and ensure all tests pass.
```

## Checkpoint questions

1. What does the dot in an LR(0) item represent?
2. Why does closure only need to add items where dot=0?
3. If a nonterminal `B` has 5 rules, how many items does closure add when it
   encounters `A → α • B β`?
4. What is a "reduce item" and how do you recognize one?
5. Why must ItemSet use deduplication?
