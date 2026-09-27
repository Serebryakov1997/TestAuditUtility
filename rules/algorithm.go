package rules

import (
	"fmt"

	"github.com/serebryakov1997/utility/audit"
)

type AlgorithmRule struct{}

func (AlgorithmRule) ID() string { return "weak-algorithm" }

var weakAlgorithms = map[string]string{
	"md4":       "MD4",
	"md5":       "MD5",
	"sha1":      "SHA-1",
	"des":       "DES",
	"3des":      "3DES",
	"tripledes": "3DES",
	"rc4":       "RC4",
}

func (r AlgorithmRule) Check(node audit.Node) []audit.Finding {
	value, ok := node.Value.(string)
	if !ok {
		return nil
	}

	name, weak := weakAlgorithms[normalize(value)]
	if !weak {
		return nil
	}

	return finding(node, r.ID(), audit.High,
		fmt.Sprintf("Set unsafed algorithm %s.", name),
		"Choose modern strong algorithm.")
}
