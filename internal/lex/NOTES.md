# Lexer Notes

Coding agents should ignore the contents of this file.

## Historical
The `lemon` tool combined lexing and parsing.

```go
        // Read the file line by line
        scanner := bufio.NewScanner(file)
        lineNo := 0
        var currentRule *Rule

        for scanner.Scan() {
                lineNo++
                line := scanner.Text()

                // Skip empty lines and comments
                line = strings.TrimSpace(line)
                if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
                        continue
                }

                // Handle directives
                if strings.HasPrefix(line, "%") {
                        err := p.parseDirective(line, lineNo)
                        if err != nil {
                                return fmt.Errorf("line %d: %v", lineNo, err)
                        }
                        continue
                }

                // Handle grammar rules
                if strings.Contains(line, "::=") {
                        currentRule, err = p.parseRule(line, lineNo)
                        if err != nil {
                                return fmt.Errorf("line %d: %v", lineNo, err)
                        }
                        p.Rule = append(p.Rule, currentRule)
                        p.Nrule++
                        continue
                }

                // Handle code blocks or continuations of rules
                // TODO: Implement code block parsing
        }
```

## Lossless Source Notes

Trivia can be used to rebuild the source from the tokens.

```go
type obsoleteToken_t struct {
	Kind           obsoleteKind_t
	LeadingTrivia  []*Span
	Value          []*Span
	TrailingTrivia []*Span
}

// Position returns the line and column number of the token.
func (t *obsoleteToken_t) Position() (int, int) {
	if t == nil {
		return 0, 0
	} else if len(t.Value) != 0 {
		return t.Value[0].Line, t.Value[0].Col
	} else if len(t.LeadingTrivia) != 0 {
		return t.LeadingTrivia[0].Line, t.LeadingTrivia[0].Col
	} else if len(t.TrailingTrivia) != 0 {
		return t.TrailingTrivia[0].Line, t.TrailingTrivia[0].Col
	}
	return 0, 0
}

// Length returns the length of the token.
func (t *obsoleteToken_t) Length() int {
	if t == nil {
		return 0
	}
	length := 0
	for _, span := range t.Value {
		length += span.Length()
	}
	return length
}

// Bytes is a helper for diagnostics / debugging.
// Returns an empty slice if there is no value.
func (t *obsoleteToken_t) Bytes() []byte {
	b := &bytes.Buffer{}
	if t != nil {
		for _, span := range t.Value {
			b.Write(span.Bytes())
		}
	}
	return b.Bytes()
}

// Source is a helper function to rebuild the input from the token stream.
func (t *obsoleteToken_t) Source() []byte {
	b := &bytes.Buffer{}
	for _, span := range t.LeadingTrivia {
		b.Write(span.Bytes())
	}
	for _, span := range t.Value {
		b.Write(span.Bytes())
	}
	for _, span := range t.TrailingTrivia {
		b.Write(span.Bytes())
	}
	return b.Bytes()
}

// Merge is a helper function to merge tokens
func (t *obsoleteToken_t) Merge(tokens ...*obsoleteToken_t) {
	for _, tok := range tokens {
		if tok.LeadingTrivia != nil {
			t.Value = append(t.Value, tok.LeadingTrivia...)
		}
		if tok.Value != nil {
			t.Value = append(t.Value, tok.Value...)
		}
		if tok.TrailingTrivia != nil {
			t.Value = append(t.Value, tok.TrailingTrivia...)
		}
	}
}

// Merge is a helper function to merge tokens
func Merge(kind obsoleteKind_t, tokens ...*obsoleteToken_t) *obsoleteToken_t {
	t := &obsoleteToken_t{Kind: kind}
	for _, tok := range tokens {
		t.Merge(tok)
	}
	return t
}

// ToSource is a helper function to rebuild the source from the token stream.
func ToSource(tokens ...*obsoleteToken_t) []byte {
	b := &bytes.Buffer{}
	for _, t := range tokens {
		for _, span := range t.LeadingTrivia {
			b.Write(span.Bytes())
		}
		for _, span := range t.Value {
			b.Write(span.Bytes())
		}
		for _, span := range t.TrailingTrivia {
			b.Write(span.Bytes())
		}
	}
	return b.Bytes()
}

type obsoleteKind_t int
```