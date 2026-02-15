// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package grammar

import (
	"strings"
	"unicode"
)

// BuilderSink adapts Builder to the parser-facing Sink interface.
type BuilderSink struct {
	B *Builder

	// Rule state (the parser streams BeginRule -> Alternative* -> EndRule).
	curRule *RuleBuilder
	curLHS  *Symbol

	// Explicit declarations from directives.
	declTokens map[string]bool // %token TOKEN
	declTypes  map[string]string

	// Optional heuristic: treat ALLCAPS-ish names as terminals unless otherwise known.
	UseHeuristicCapsAsTerminal bool
}

// NewBuilderSink constructs a sink around a Builder.
func NewBuilderSink(b *Builder) *BuilderSink {
	return &BuilderSink{
		B:                          b,
		declTokens:                 map[string]bool{},
		declTypes:                  map[string]string{},
		UseHeuristicCapsAsTerminal: true,
	}
}

// --------------------
// Sink implementation
// --------------------

func (s *BuilderSink) ParserError(at *Span, msg string) {
	if s == nil || s.B == nil {
		return
	}
	s.B.error(at, "%s", msg)
}

func (s *BuilderSink) Directive(d Directive) {
	if s == nil || s.B == nil {
		return
	}
	switch d.Kind {
	case DirStartSymbol:
		name := strings.TrimSpace(d.Value)
		if name == "" {
			s.B.error(d.At, "%%start_symbol requires a symbol name")
			return
		}
		sym := s.B.EnsureNonterminal(name, d.At)
		s.B.SetStart(sym, d.At)

	case DirToken:
		// d.Value may hold one token, or d.List may hold many. Support both.
		if v := strings.TrimSpace(d.Value); v != "" {
			s.declTokens[v] = true
			s.B.EnsureTerminal(v, d.At)
		}
		for _, sr := range d.List {
			name := strings.TrimSpace(sr.Name)
			if name == "" {
				continue
			}
			s.declTokens[name] = true
			s.B.EnsureTerminal(name, sr.At)
		}

	case DirType:
		// %type NonTerminal {TYPE} or similar
		typeTag := strings.TrimSpace(d.Value)
		if typeTag == "" {
			s.B.error(d.At, "%%type requires a type tag")
			return
		}
		for _, sr := range d.List {
			name := strings.TrimSpace(sr.Name)
			if name == "" {
				continue
			}
			// In Lemon, %type applies to nonterminals.
			sym := s.B.EnsureNonterminal(name, sr.At)
			s.declTypes[name] = typeTag
			s.B.SetTypeTag(sym, typeTag, sr.At)
		}

	case DirTokenType:
		// This can mean different things in different Lemon-ish dialects.
		// We store it as a directive and leave deeper meaning for later lessons.
		s.B.SetDirective("token_type", d.Value, d.At)

	case DirLeft, DirRight, DirNonassoc:
		assoc := AssocNone
		switch d.Kind {
		case DirLeft:
			assoc = AssocLeft
		case DirRight:
			assoc = AssocRight
		case DirNonassoc:
			assoc = AssocNonassoc
		}

		var terms []*Symbol
		// Prefer d.List.
		for _, sr := range d.List {
			name := strings.TrimSpace(sr.Name)
			if name == "" {
				continue
			}
			s.declTokens[name] = true
			terms = append(terms, s.B.EnsureTerminal(name, sr.At))
		}
		// Also allow d.Value to contain a whitespace-separated list.
		if len(terms) == 0 {
			for _, name := range strings.Fields(d.Value) {
				s.declTokens[name] = true
				terms = append(terms, s.B.EnsureTerminal(name, d.At))
			}
		}
		if len(terms) == 0 {
			s.B.error(d.At, "precedence directive requires at least one terminal")
			return
		}
		s.B.DefinePrecedenceGroup(assoc, terms, d.At)

	case DirInclude, DirCode, DirFallback, DirUnknown:
		// For now: just store a generic key/value for later phases.
		// Your grammar parser can set Key/Value meaningfully.
		key := strings.TrimSpace(d.Key)
		if key == "" {
			// Provide a reasonable default for unknown directives.
			key = "directive"
		}
		s.B.SetDirective(key, d.Value, d.At)

	default:
		// Forward-compat: store anything we don't recognize.
		key := strings.TrimSpace(d.Key)
		if key == "" {
			key = "directive"
		}
		s.B.SetDirective(key, d.Value, d.At)
	}
}

func (s *BuilderSink) BeginRule(lhs SymRef) {
	if s == nil || s.B == nil {
		return
	}
	// End any open rule defensively (parser bug guard).
	if s.curRule != nil {
		s.B.warn(lhs.At, "begin rule while previous rule still open; closing previous rule")
		s.curRule.End()
		s.curRule = nil
		s.curLHS = nil
	}

	name := strings.TrimSpace(lhs.Name)
	if name == "" {
		s.B.error(lhs.At, "rule LHS is empty")
		return
	}

	s.curLHS = s.B.EnsureNonterminal(name, lhs.At)
	// Apply any pending %type for this symbol (if not already set).
	if tt, ok := s.declTypes[name]; ok && tt != "" {
		s.B.SetTypeTag(s.curLHS, tt, lhs.At)
	}

	s.curRule = s.B.BeginRule(s.curLHS, lhs.At)
}

func (s *BuilderSink) Alternative(alt Alt) {
	if s == nil || s.B == nil {
		return
	}
	if s.curRule == nil || s.curLHS == nil {
		s.B.error(alt.At, "alternative encountered without an open rule")
		return
	}

	// Resolve RHS symbols.
	rhs := make([]*SymbolRef, 0, len(alt.RHS))
	for _, sr := range alt.RHS {
		sym := s.resolveSymbolInRHS(sr)
		// Apply per-occurrence TypeTag if present (rare, but harmless).
		if sr.TypeTag != "" {
			s.B.SetTypeTag(sym, sr.TypeTag, sr.At)
		}
		rhs = append(rhs, s.B.NewRef(sym, sr.Label, sr.At))
	}

	// Resolve precedence override.
	var prec *Symbol
	if alt.Prec != nil {
		ps := *alt.Prec
		name := strings.TrimSpace(ps.Name)
		if name == "" {
			s.B.error(ps.At, "precedence override symbol is empty")
		} else {
			// Prec symbol must be terminal.
			s.declTokens[name] = true
			prec = s.B.EnsureTerminal(name, ps.At)
		}
	}

	s.curRule.Alt(rhs, alt.Action, prec, alt.At)
}

func (s *BuilderSink) EndRule(at *Span) {
	if s == nil || s.B == nil {
		return
	}
	if s.curRule == nil {
		// Allow benign extra EndRule calls.
		return
	}
	s.curRule.End()
	s.curRule = nil
	s.curLHS = nil
}

// --------------------
// Symbol resolution
// --------------------

// resolveSymbolInRHS decides whether an RHS symbol is terminal or nonterminal.
// Precedence rules (explicit > inferred):
//  1. If explicitly declared by %token -> terminal
//  2. If already interned, use its existing kind
//  3. Heuristic: ALLCAPS-ish (or contains non-letters) => terminal
//  4. Otherwise => nonterminal
func (s *BuilderSink) resolveSymbolInRHS(sr SymRef) *Symbol {
	name := strings.TrimSpace(sr.Name)
	if name == "" {
		s.B.error(sr.At, "rhs symbol name is empty")
		return s.B.internDummy(sr.At)
	}

	// 1) Explicit %token declaration.
	if s.declTokens[name] {
		return s.B.EnsureTerminal(name, sr.At)
	}

	// 2) Already known.
	if existing, ok := s.B.Lookup(name); ok {
		return existing
	}

	// 3) Heuristic.
	if s.UseHeuristicCapsAsTerminal && looksLikeTerminal(name) {
		return s.B.EnsureTerminal(name, sr.At)
	}

	// 4) Default: nonterminal.
	return s.B.EnsureNonterminal(name, sr.At)
}

// looksLikeTerminal returns true for names that appear token-like:
// - contains any non-letter (e.g. "+", "==", "TK_ID", "NUM1")
// - OR is all-uppercase letters (ASCII) (e.g. "PLUS", "MINUS")
func looksLikeTerminal(name string) bool {
	if name == "" {
		return false
	}
	hasLetter := false
	allUpperLetters := true

	for _, r := range name {
		if unicode.IsLetter(r) {
			hasLetter = true
			// Only treat ASCII-ish upper as "upper" for this heuristic.
			if unicode.ToUpper(r) != r {
				allUpperLetters = false
			}
			continue
		}
		// Any non-letter character makes it token-ish.
		return true
	}
	return hasLetter && allUpperLetters
}
