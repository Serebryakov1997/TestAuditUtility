package restapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/serebryakov1997/utility/audit"
	"github.com/serebryakov1997/utility/config"
)

func TestAuditHTTP(t *testing.T) {
	tests := []struct {
		name, method, contentType, body string
		result                          []audit.Finding
		err                             error
		status                          int
		want                            string
		called                          bool
		format                          string
	}{
		{
			name:        "findings are successful analysis",
			method:      "POST",
			contentType: "application/json",
			body:        `{"log": { "level": "debug" }}`,
			result: []audit.Finding{{
				RuleID:         "debug-enabled",
				Severity:       "LOW",
				Path:           "$.log.level",
				Message:        "Logging in debug mode",
				Recommendation: "Use level info or higher",
			}},
			status: 200,
			want: `[{"rule_id":"debug-enabled","severity":"LOW",` +
				`"path":"$.log.level","message":"Logging in debug mode",` +
				`"recommendation":"Use level info or higher"}]`,
			called: true,
			format: config.JSON,
		},
		{
			name:        "nil becomes empty array",
			method:      "POST",
			contentType: "application/json; charset=utf-8",
			body:        `{}`,
			status:      200,
			want:        `[]`,
			called:      true,
			format:      config.JSON,
		},
		{
			name:        "yaml",
			method:      "POST",
			contentType: "application/yaml",
			body:        "debug: true",
			status:      200,
			want:        `[]`,
			called:      true,
			format:      config.YAML,
		},
		{name: "method", method: "GET", status: 405},
		{name: "unsuppported type", method: "POST", contentType: "text/plain", body: `{}`, status: 415},
		{name: "missing type", method: "POST", body: `{}`, status: 415},
		{name: "empty body", method: "POST", contentType: "application/json", body: " \n ", status: 400},
		{
			name:        "parser error",
			method:      "POST",
			contentType: "application/json",
			body:        `{`,
			err:         fmt.Errorf("%w: secret", ErrInvalidConfig),
			status:      400,
			want:        `{"error":"invalid configuration"}`,
			called:      true,
			format:      config.JSON,
		},
		{
			name:        "internal error",
			method:      "POST",
			contentType: "application/json",
			body:        `{}`,
			err:         errors.New("private detail"),
			status:      500,
			want:        `{"error":"audit failed"}`,
			called:      true,
			format:      config.JSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false

			h, err := NewHandler(
				func(ctx context.Context,
					data []byte,
					format string) ([]audit.Finding, error) {
					called = true
					if string(data) != tt.body || format != tt.format {
						t.Fatalf("unexpected input: %q, %q", data, format)
					}

					return tt.result, tt.err
				})

			if err != nil {
				t.Fatalf("create HTTP handler: %v", err)
			}

			r := httptest.NewRequest(tt.method, "/api/v1/audit", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d", w.Code, tt.status)
			}
			if called != tt.called {
				t.Fatalf("called = %v, want %v", called, tt.called)
			}

			if tt.want != "" && strings.TrimSpace(w.Body.String()) != tt.want {
				t.Fatalf("body = %q, want %q", w.Body.String(), tt.want)
			}
		})
	}
}

func TestAuditTooLargeBody(t *testing.T) {
	handler, err := NewHandler(func(
		ctx context.Context,
		data []byte,
		format string,
	) ([]audit.Finding, error) {
		t.Fatal("audit function must not be called")
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	body := bytes.Repeat([]byte("a"), int(config.MaxBytes)+1)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/audit",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d",
			rec.Code,
			http.StatusRequestEntityTooLarge,
		)
	}
}
