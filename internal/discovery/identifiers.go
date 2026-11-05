package discovery

import (
	"sort"
	"strings"
)

// identifierMatcher searches all names in a single pass over a test file.
// Transitions include failure links, so matching takes constant work per byte
// plus the number of candidate occurrences.
type identifierMatcher struct {
	names  []string
	states []identifierState
}
type identifierState struct {
	next    [256]int
	fail    int
	outputs []int
}

func newIdentifierMatcher(names []string) *identifierMatcher {
	m := &identifierMatcher{names: names, states: make([]identifierState, 1)}
	for id, name := range names {
		if name == "" || strings.Contains(name, "\n") {
			continue
		}
		state := 0
		for i := 0; i < len(name); i++ {
			b := name[i]
			next := m.states[state].next[b]
			if next == 0 {
				next = len(m.states)
				m.states = append(m.states, identifierState{})
				m.states[state].next[b] = next
			}
			state = next
		}
		m.states[state].outputs = append(m.states[state].outputs, id)
	}
	var queue []int
	for _, next := range m.states[0].next {
		if next != 0 {
			queue = append(queue, next)
		}
	}
	for head := 0; head < len(queue); head++ {
		state := queue[head]
		for b := 0; b < 256; b++ {
			next := m.states[state].next[b]
			fallback := m.states[m.states[state].fail].next[b]
			if next == 0 {
				m.states[state].next[b] = fallback
				continue
			}
			m.states[next].fail = fallback
			m.states[next].outputs = append(m.states[next].outputs, m.states[fallback].outputs...)
			queue = append(queue, next)
		}
	}
	return m
}

func (m *identifierMatcher) find(content []byte) []identifierMatch {
	if len(m.states) == 1 {
		return nil
	}
	var matches []identifierMatch
	lastLine := make([]int, len(m.names))
	nextOffset := make([]int, len(m.names))
	state, line := 0, 1
	for i, b := range content {
		state = m.states[state].next[b]
		for _, id := range m.states[state].outputs {
			name := m.names[id]
			start, end := i+1-len(name), i+1
			// The previous implementation advanced by the full name even after an
			// invalid boundary, and restarted its search on each line.
			if start < nextOffset[id] {
				continue
			}
			nextOffset[id] = end
			if (start > 0 && isIdentifierByte(content[start-1])) || (end < len(content) && isIdentifierByte(content[end])) {
				continue
			}
			if lastLine[id] != line {
				matches = append(matches, identifierMatch{line: line, name: name})
				lastLine[id] = line
			}
		}
		if b == '\n' {
			line++
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].line != matches[j].line {
			return matches[i].line < matches[j].line
		}
		return matches[i].name < matches[j].name
	})
	return matches
}
