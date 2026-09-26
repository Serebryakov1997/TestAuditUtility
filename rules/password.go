package rules

import (
	"regexp"
	"strings"

	"github.com/serebryakov1997/utility/audit"
)

var regChecking = regexp.MustCompile(`^\$\{A-Za-z_[A-Za-z0-9_]*\}$`)

type PasswordRule struct{}

func (PasswordRule) ID() string { return "plaintext-password" }

func (r PasswordRule) Check(node audit.Node) []audit.Finding {	
	key := normalize(node.Key)
	if !(oneOf(key, "password", "passwd") || 
		strings.HasSuffix(key, "password") || strings.HasSuffix(key, "passwd")) {
		return nil
	}

	switch v := node.Value.(type) {
	case string:
		if strings.TrimSpace(v) == "" || regChecking.MatchString(v) {
			return nil
		}
	case float64:
	default:
		return nil
	}

	return finding(node, r.ID(), audit.High,
		"Founded password in cleartext.",
		"Move the password to a secret repository or use link to env value.")
}