package audit

const (
	Low string  = "LOW"
	Medium string = "MEDIUM"
	High string = "HIGH"
)

type Finding struct {
	RuleID string `json:"rule_id"`
	Severity string `json:"severity"`
	Path string `json:"path"`
	Message string `json:"message"`
	Recommendation string `json:"recommendation"`
}