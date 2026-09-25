package rules

import (
	"net"
	"strings"

	"github.com/serebryakov1997/utility/audit"
)

type BindRule struct{}

func (BindRule) ID() string {return "all-interfaces"}

func (r BindRule) Check(node audit.Node) []audit.Finding {
	if !oneOf(normalize(node.Key), "host", "bind", "listen", "address") {
		return nil
	}

	address, ok := node.Value.(string)
	if !ok {
		return nil
	}

	host := strings.TrimSpace(address)
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	if host != "0.0.0.0" && host != "::" {
		return nil
	}

	return finding(node, r.ID(), audit.Medium,
		"Service can listen all network interfaces.",
		"Set need interface")
}