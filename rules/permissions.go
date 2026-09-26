package rules

import (
	"strconv"
	"strings"

	"github.com/serebryakov1997/utility/audit"
)

type PermissionsRule struct{}

func (PermissionsRule) ID() string { return "broad-permissions" }

func (r PermissionsRule) Check(node audit.Node) []audit.Finding {
	if !oneOf(normalize(node.Key), "permissions", "filemode") {
		return nil
	}

	var mode uint64
	var err error
	switch value := node.Value.(type) {
	case string:
		value = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), "0o")
		if value == "" {
			return nil
		}
		mode, err = strconv.ParseUint(value, 8, 12)
	case float64:
		mode, err = strconv.ParseUint(
			strconv.FormatFloat(value, 'f', 0, 64), 10, 12)
	default:
		return nil
	}

	if err != nil || mode&0022 == 0 {
		return nil
	}

	return finding(node, r.ID(), audit.Medium,
		"Permissions allow writing to all users.",
		"Remove unnecessary write permissions and restrict access.")
}