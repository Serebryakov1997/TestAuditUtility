package audit

import (
	"strconv"
	"strings"
)

// Segment save source key or index. Key "a.b" not equal path a -> b.
type Segment struct {
	Key string
	Index int
	IsIndex bool
}

type Path []Segment

// String use dot notation for simple keys and bracket notation for others.
// Control characters in keys are escaped
func (p Path) String() string {
	var b strings.Builder
	b.WriteByte('$')

	for _, s := range p {
		if s.IsIndex {
			b.WriteByte('[')
			b.WriteString(strconv.Itoa(s.Index))
			b.WriteByte(']')
		} else if simpleKey(s.Key) {
			b.WriteByte('.')
			b.WriteString(s.Key)
		} else {
			b.WriteByte('[')
			b.WriteString(strconv.Quote(s.Key))
			b.WriteByte(']')
		}
	}

	return b.String()
}

func simpleKey(key string) bool {
	if key == "" {
		return false
	}

	for i, r := range key {
		letter := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_'
		digit := r >= '0' && r <= '9'
		if !letter && !(i > 0 && digit) {
			return false
		}
	}

	return true
}