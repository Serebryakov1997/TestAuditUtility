package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/serebryakov1997/utility/config"
)

func TestRunLargeFile(t *testing.T) {

	const count = 10_000

	const payloadSize = 950
	padding := strings.Repeat("a", payloadSize)

	var builder strings.Builder

	builder.WriteString(`{"services":[`)

	for i := 0; i < count; i++ {
		if i > 0 {
			builder.WriteByte(',')
		}

		builder.WriteString(`{"payload":"`)
		builder.WriteString(padding)
		builder.WriteString(`","debug":`)

		if i == count-1 {
			builder.WriteString("true")
		} else {
			builder.WriteString("false")
		}

		builder.WriteByte('}')
	}

	builder.WriteString(`]}`)

	data := []byte(builder.String())

	if len(data) <= 9<<20 {
		t.Fatalf("file size = %d bytes, want more than 9 MiB", len(data))
	}

	if int64(len(data)) > config.MaxBytes {
		t.Fatalf(
			"file size = %d bytes exceeds limit %d bytes",
			len(data),
			config.MaxBytes,
		)
	}

	path := filepath.Join(t.TempDir(), "large.json")

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatalf("Chmod() error: %v", err)
	}

	cases := []struct {
		name string
		args []string
		code int
	}{
		{
			name: "regular",
			args: []string{path},
			code: ExitFindings,
		},
		{
			name: "silent",
			args: []string{"--silent", path},
			code: ExitOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer

			code := Run(
				tc.args,
				strings.NewReader("stdin must not be read"),
				&out,
				&errOut,
			)

			if code != tc.code {
				t.Fatalf(
					"code = %d, want %d; stderr = %q",
					code,
					tc.code,
					errOut.String(),
				)
			}

			if errOut.Len() != 0 {
				t.Fatalf("unexpected stderr: %q", errOut.String())
			}

			want := fmt.Sprintf(
				"LOW $.services[%d].debug",
				count-1,
			)

			if !strings.Contains(out.String(), want) {
				t.Fatalf(
					"missing finding %q; stdout = %q",
					want,
					out.String(),
				)
			}

			if got := strings.Count(
				out.String(),
				"debug-enabled",
			); got != 1 {
				t.Fatalf(
					"debug-enabled occurrences = %d, want 1",
					got,
				)
			}

			if strings.Contains(out.String(), "file-permissions") {
				t.Fatal("unexpected file-permissions finding")
			}
		})
	}
}
