package grpcapi

import (
	"context"
	"strings"

	"github.com/serebryakov1997/utility/audit"
	"github.com/serebryakov1997/utility/config"
	auditv1 "github.com/serebryakov1997/utility/grpc/proto/audit/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	auditv1.UnimplementedAuditServiceServer

	analyzer audit.Analyzer
}

var _ auditv1.AuditServiceServer = (*Handler)(nil)

func NewHandler(analyzer audit.Analyzer) *Handler {
	return &Handler{analyzer: analyzer}
}

func (h *Handler) Audit(
	ctx context.Context,
	req *auditv1.AuditRequest,
) (*auditv1.AuditResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}

	if req == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"request is required",
		)
	}

	if strings.TrimSpace(req.GetContent()) == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"configuration is empty",
		)
	}

	format := strings.ToLower(strings.TrimSpace(req.GetFormat()))

	if format == "" {
		format = config.Auto
	}

	switch format {
	case config.Auto, config.JSON, config.YAML:
	default:
		return nil, status.Error(
			codes.InvalidArgument,
			"format must be auto, json or yaml",
		)
	}

	parseConfig, err := config.Parse(
		strings.NewReader(req.GetContent()),
		format,
	)

	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"invalid configuration or parser limits exceed",
		)
	}

	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}

	findings, err := h.analyzer.Analyze(parseConfig)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			"configuration analysis failed",
		)
	}

	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}

	response := &auditv1.AuditResponse{
		Findings: make([]*auditv1.Finding, 0, len(findings)),
	}

	for _, finding := range findings {
		response.Findings = append(
			response.Findings,
			&auditv1.Finding{
				RuleId:         finding.RuleID,
				Severity:       string(finding.Severity),
				Path:           finding.Path,
				Message:        finding.Message,
				Recommendation: finding.Recommendation,
			},
		)
	}

	return response, nil
}
