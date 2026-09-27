package audit

type Analyzer struct {
	rules []Rule
}

func New(rules ...Rule) *Analyzer {
	return &Analyzer{rules: append([]Rule(nil), rules...)}
}

func (a *Analyzer) Analyze(root any) ([]Finding, error) {
	findings := make([]Finding, 0)
	err := Walk(Node{Value: root}, func(node Node) {
		for _, rule := range a.rules {
			findings = append(findings, rule.Check(node)...)
		}
	})

	if err != nil {
		return nil, err
	}

	return findings, nil
}
