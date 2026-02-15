// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package grammar

// Finalize performs semantic validation and emits diagnostics.
// It does not panic; it records errors/warnings and returns the grammar anyway.
// Call this after the grammar-file parser has finished emitting events.
func (b *Builder) Finalize() *Grammar {
	if b == nil || b.g == nil {
		return nil
	}

	g := b.g

	// ---- 1) Basic existence checks ----

	if len(g.Rules) == 0 {
		b.error(nil, "grammar has no rules")
		return g
	}

	// Start symbol: inferred in BeginRule if nil; still validate.
	if g.Start == nil {
		b.error(nil, "start symbol is not set and could not be inferred")
	} else if g.Start.Kind != SymNonterminal {
		b.error(g.Start.DeclaredAt, "start symbol %q must be a nonterminal", g.Start.Name)
	}

	// ---- 2) Build indices used by validation ----

	// lhsHasRule: which nonterminals appear as LHS of any rule?
	lhsHasRule := make(map[*Symbol]bool, len(g.Rules))
	for _, r := range g.Rules {
		if r == nil || r.LHS == nil {
			b.error(nil, "rule has nil LHS")
			continue
		}
		if r.LHS.Kind != SymNonterminal {
			b.error(r.At, "rule LHS %q must be a nonterminal", r.LHS.Name)
		}
		lhsHasRule[r.LHS] = true
	}

	// usage counts for warnings
	used := make(map[*Symbol]int, len(g.Symbols))

	// Track edges among nonterminals for reachability.
	// NT A -> NT B if A has an alternative referencing B on RHS.
	edges := make(map[*Symbol]map[*Symbol]bool)

	// ---- 3) Validate productions ----

	for _, r := range g.Rules {
		if r == nil || r.LHS == nil {
			continue
		}

		// Make sure map entry exists for reachability.
		if _, ok := edges[r.LHS]; !ok {
			edges[r.LHS] = map[*Symbol]bool{}
		}

		if len(r.Alternatives) == 0 {
			b.warn(r.At, "rule %q has no alternatives", r.LHS.Name)
			continue
		}

		for _, alt := range r.Alternatives {
			if alt == nil {
				b.error(r.At, "rule %q has a nil alternative", r.LHS.Name)
				continue
			}

			// Validate RHS refs.
			for pos, sr := range alt.RHS {
				if sr == nil || sr.Sym == nil {
					b.error(alt.At, "rule %q has nil rhs symbol at position %d", r.LHS.Name, pos)
					continue
				}

				used[sr.Sym]++

				// Record reachability edge if RHS contains a nonterminal.
				if sr.Sym.Kind == SymNonterminal {
					edges[r.LHS][sr.Sym] = true
				}
			}

			// Validate precedence override.
			if alt.PrecSym != nil {
				if alt.PrecSym.Kind != SymTerminal {
					b.error(alt.At, "precedence override symbol %q must be a terminal", alt.PrecSym.Name)
				} else if alt.PrecSym.Precedence == 0 {
					// This is not necessarily fatal; many tools allow it,
					// but it usually indicates a missed %left/%right entry.
					b.warn(alt.At, "precedence override uses %q, but it has no precedence (missing %%left/%%right/%%nonassoc?)", alt.PrecSym.Name)
				}
				used[alt.PrecSym]++
			}
		}
	}

	// ---- 4) Undefined nonterminals (referenced but no rule) ----

	for sym, n := range used {
		if n == 0 || sym == nil {
			continue
		}
		if sym.Kind == SymNonterminal && !lhsHasRule[sym] {
			// Mention where we first saw it if available.
			at := sym.DeclaredAt
			b.error(at, "nonterminal %q is used but has no rule", sym.Name)
		}
	}

	// ---- 5) Reachability from start symbol ----

	reachable := map[*Symbol]bool{}
	if g.Start != nil && g.Start.Kind == SymNonterminal {
		// BFS/DFS from start over nonterminal edges.
		stack := []*Symbol{g.Start}
		reachable[g.Start] = true

		for len(stack) > 0 {
			nt := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			for next := range edges[nt] {
				if next == nil || next.Kind != SymNonterminal {
					continue
				}
				if !reachable[next] {
					reachable[next] = true
					stack = append(stack, next)
				}
			}
		}

		// Warn for nonterminals that have rules but are unreachable.
		for nt := range lhsHasRule {
			if nt == nil {
				continue
			}
			if !reachable[nt] {
				b.warn(nt.DeclaredAt, "nonterminal %q has rules but is unreachable from start symbol %q", nt.Name, g.Start.Name)
			}
		}
	}

	// ---- 6) Unused symbol warnings ----

	for _, sym := range g.Symbols {
		if sym == nil {
			continue
		}
		// Ignore dummy placeholder.
		if sym.Name == "<invalid>" {
			continue
		}
		if used[sym] == 0 {
			switch sym.Kind {
			case SymTerminal:
				b.warn(sym.DeclaredAt, "terminal %q is declared but never used", sym.Name)
			case SymNonterminal:
				// If it has rules, it'll be caught by reachability warnings.
				// If it has no rules, it might also be caught as "used but has no rule".
				if lhsHasRule[sym] {
					b.warn(sym.DeclaredAt, "nonterminal %q has rules but is never referenced (except as LHS)", sym.Name)
				} else {
					b.warn(sym.DeclaredAt, "nonterminal %q is declared but never used and has no rules", sym.Name)
				}
			}
		}
	}

	return g
}
