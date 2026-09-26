package rules

import (
	"github.com/serebryakov1997/utility/audit"
)

type TLSRule struct{}

func (TLSRule) ID() string { return "tls-unsafe" }

func (r TLSRule) Check(node audit.Node) []audit.Finding {
	value, ok := boolean(node.Value)
	if !ok {
		return nil
	}

	key := normalize(node.Key)
	tlsOwner := oneOf(owner(node), "tls", "ssl")
	disabled := !value && (oneOf(key, "tls", "ssl", "enabled", "verify") || key == "enabled" && tlsOwner)
	if disabled {
		return finding(node, r.ID(), audit.High,
			"TLS is disabled.",
			"Enable TLS.")
	}

	return nil
}