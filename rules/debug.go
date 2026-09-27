package rules

import (
	"strings"

	"github.com/serebryakov1997/utility/audit"
)

type DebugRule struct{}

func (DebugRule) ID() string { return "debug-enabled" }

func (r DebugRule) Check(node audit.Node) []audit.Finding {
	key := normalize(node.Key)
	if oneOf(key, "debug", "debugmode") {
		if enabled, ok := boolean(node.Value); ok && enabled {
			return finding(node, r.ID(), audit.Low,
				"Debug is enabled.",
				"Disable debug in work environment")
		}
	}

	isLevel := oneOf(key, "loglevel", "logginglevel") ||
		key == "level" &&
			oneOf(owner(node), "log", "level", "logs", "logger", "logging")

	level, ok := node.Value.(string)
	if isLevel && ok &&
		oneOf(strings.ToLower(strings.TrimSpace(level)), "debug") {
		return finding(node, r.ID(), audit.Low,
			"Logging in debug mode.",
			`Use level "info" or higher.`)
	}
	return nil
}
