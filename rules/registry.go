package rules

import "github.com/serebryakov1997/utility/audit"

func RegisteredRules() []audit.Rule {
	return []audit.Rule{
		DebugRule{},
		PasswordRule{},
		BindRule{},
		TLSRule{},
		AlgorithmRule{},
		PermissionsRule{},
	}
}