package grpcapi

import (
	"context"
	"testing"
	"time"

	"github.com/serebryakov1997/utility/audit"
	"github.com/serebryakov1997/utility/config"
	auditv1 "github.com/serebryakov1997/utility/grpc/proto/audit/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type analyzerFunc func(config any) ([]audit.Finding, error)

func (f analyzerFunc) Analyze(config any) ([]audit.Finding, error) {
	return f(config)
}

func TestHandlerAudit_Formats(t *testing.T) {
	cases := []struct {
		name    string
		content string
		format  string
	}{
		{
			name:    "json",
			content: `{"debug": true}`,
			format:  config.JSON,
		},
		{
			name:    "yaml",
			content: "debug: true\n",
			format:  config.YAML,
		},
		{
			name:    "auto json",
			content: `{"debug": true}`,
			format:  config.Auto,
		},
		{
			name:    "auto yaml",
			content: "debug: true\n",
			format:  config.Auto,
		},
		{
			name:    "empty format defaults to auto",
			content: `{"debug": true}`,
			format:  "",
		},
		{
			name:    "format normalization",
			content: `{"debug": true}`,
			format:  " JSON ",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0

			analyzer := analyzerFunc(func(config any) ([]audit.Finding, error) {
				calls++

				object, ok := config.(map[string]any)
				if !ok {
					t.Fatalf("config type = %T, want map[string]any", config)
				}

				if len(object) != 1 {
					t.Fatalf("object fields = %d, want 1", len(object))
				}

				value, exists := object["debug"]
				if !exists {
					t.Fatal("field debug is missing")
				}

				debug, ok := value.(bool)
				if !ok {
					t.Fatalf("debug type = %T, want bool", value)
				}

				if !debug {
					t.Fatal("debug = false, want true")
				}

				return nil, nil
			})

			handler := NewHandler(analyzer)

			response, err := handler.Audit(
				context.Background(),
				&auditv1.AuditRequest{
					Content: tt.content,
					Format:  tt.format,
				},
			)

			if err != nil {
				t.Fatalf("Audit() error: %v", err)
			}

			if response == nil {
				t.Fatal("Audit() returned nil response")
			}

			if len(response.GetFindings()) != 0 {
				t.Fatalf("unexpected findings: %v", response.GetFindings())
			}

			if calls != 1 {
				t.Fatalf("Analyze() calls = %d, want 1", calls)
			}
		})
	}
}

func TestHandlerAudit_Findings(t *testing.T) {
	analyzer := analyzerFunc(func(config any) ([]audit.Finding, error) {
		return []audit.Finding{
			{
				RuleID:         "debug-enabled",
				Severity:       "LOW",
				Path:           "$.debug",
				Message:        "Debug is enabled.",
				Recommendation: "Disable debug.",
			},
			{
				RuleID:         "plaintext-password",
				Severity:       "HIGH",
				Path:           "$.password",
				Message:        "Plaintext password.",
				Recommendation: "Use a secret store.",
			},
		}, nil
	})

	handler := NewHandler(analyzer)

	response, err := handler.Audit(
		context.Background(),
		&auditv1.AuditRequest{
			Content: `{"debug":true,"password":"test_password"}`,
			Format:  config.JSON,
		},
	)

	if err != nil {
		t.Fatalf("Audit() error: %v", err)
	}

	want := &auditv1.AuditResponse{
		Findings: []*auditv1.Finding{
			{
				RuleId:         "debug-enabled",
				Severity:       "LOW",
				Path:           "$.debug",
				Message:        "Debug is enabled.",
				Recommendation: "Disable debug.",
			},
			{
				RuleId:         "plaintext-password",
				Severity:       "HIGH",
				Path:           "$.password",
				Message:        "Plaintext password.",
				Recommendation: "Use a secret store.",
			},
		},
	}

	if !proto.Equal(response, want) {
		t.Fatalf("response = %v, want %v", response, want)
	}
}

func TestHandlerAudit_InvalidRequest(t *testing.T) {
	cases := []struct {
		name string
		req  *auditv1.AuditRequest
	}{
		{
			name: "nil request",
			req:  nil,
		},
		{
			name: "empty content",
			req:  &auditv1.AuditRequest{Format: config.JSON},
		},
		{
			name: "whitespace content",
			req: &auditv1.AuditRequest{
				Content: " \n\t ",
				Format:  config.YAML,
			},
		},
		{
			name: "unsupported format",
			req: &auditv1.AuditRequest{
				Content: `{}`,
				Format:  "xml",
			},
		},
		{
			name: "invalid json",
			req: &auditv1.AuditRequest{
				Content: `{"debug":`,
				Format:  "json",
			},
		},
		{
			name: "invalid yaml",
			req: &auditv1.AuditRequest{
				Content: "debug: [",
				Format:  config.YAML,
			},
		},
		{
			name: "scalar root",
			req: &auditv1.AuditRequest{
				Content: `42`,
				Format:  config.JSON,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(analyzerFunc(
				func(config any) ([]audit.Finding, error) {
					t.Fatal("Analyze() must not be called")
					return nil, nil
				},
			))

			response, err := handler.Audit(context.Background(), tt.req)

			assertAuditError(t, response, err, codes.InvalidArgument)
		})
	}
}

func TestHandlerAudit_ContextError(t *testing.T) {
	cases := []struct {
		name     string
		newCtx   func() (context.Context, context.CancelFunc)
		wantCode codes.Code
	}{
		{
			name: "canceled",
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			wantCode: codes.Canceled,
		},
		{
			name: "deadline exceed",
			newCtx: func() (context.Context, context.CancelFunc) {
				return context.WithDeadline(
					context.Background(),
					time.Now().Add(-time.Second),
				)
			},
			wantCode: codes.DeadlineExceeded,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.newCtx()
			defer cancel()

			handler := NewHandler(analyzerFunc(
				func(config any) ([]audit.Finding, error) {
					t.Fatal("Analyze() must not be called")
					return nil, nil
				},
			))

			response, err := handler.Audit(
				ctx,
				&auditv1.AuditRequest{
					Content: `{}`,
					Format:  config.JSON,
				},
			)

			assertAuditError(t, response, err, tt.wantCode)
		})
	}
}

func assertAuditError(
	t *testing.T,
	response *auditv1.AuditResponse,
	err error,
	wantCode codes.Code,
) {
	t.Helper()

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if got := status.Code(err); got != wantCode {
		t.Fatalf("status code = %v, want %v; error: %v", got, wantCode, err)
	}

	if response != nil {
		t.Fatalf("expected nil response, got %v", response)
	}
}
