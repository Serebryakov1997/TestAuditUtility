package filecheck

import (
	"fmt"
	"os"

	"github.com/serebryakov1997/utility/audit"
)

func Permissions(path string) ([]audit.Finding, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat file %q: %w", path, err)
	}

	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%q is not a regular file", path)
	}

	permissions := info.Mode().Perm()

	const (
		groupWrite os.FileMode = 0o020
		otherWrite os.FileMode = 0o002
	)

	findings := make([]audit.Finding, 0)

	if permissions&otherWrite != 0 {
		findings = append(findings, audit.Finding{
			RuleID:   "file-permissions",
			Severity: audit.High,
			Path:     "@file.permissions",
			Message: fmt.Sprintf(
				"File permissions %04o allow writing by other users.",
				permissions,
			),
			Recommendation: "Remove write permission for other users; review group access.",
		})

		return findings, nil
	}

	if permissions&groupWrite != 0 {
		findings = append(findings, audit.Finding{
			RuleID:   "file-permissions",
			Severity: audit.Medium,
			Path:     "@file.permissions",
			Message: fmt.Sprintf(
				"File permissions %04o allow writing by the group.",
				permissions,
			),
			Recommendation: "Remove group write permission unless explicitly required.",
		})
	}

	return findings, nil
}
