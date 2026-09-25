package rules

import (
	"strings"

	"github.com/serebryakov1997/utility/audit"
)

func normalize(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '_' || r == '-' || r == '.' {
			return -1
		}
		return r
	}, strings.ToLower(strings.TrimSpace(s)))
}

func oneOf(needValue string, values ...string) bool {
	for _, value := range values {
		if needValue == value {
			return true
		}
	}
	return false
}

func boolean(value any) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "yes", "on", "1":
			return true, true
		case "false", "no", "off", "0":
			return false, true
		}
	}
	return false, false
}

// owner return nearest parent field, 
// skip indexes of array and this field
func owner(node audit.Node) string {
	for i := len(node.Path)-1; i >= 0; i-- {
		if node.Path[i].IsIndex {
			continue
		}

		return normalize(node.Path[i].Key)
	}
	return ""
}

func finding(node audit.Node, id, severity, message, advice string) []audit.Finding {
	return []audit.Finding{{
		RuleID: id,
		Severity: severity,
		Path: node.Path.String(),
		Message: message,
		Recommendation: advice,
	}}
}