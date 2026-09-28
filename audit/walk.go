package audit

// Walk visits every node one time
func Walk(node Node, visit func(Node)) error {
	visit(node)

	switch value := node.Value.(type) {
	case map[string]any:
		for key, child := range value {
			next := Node{
				Path:   append(node.Path, Segment{Key: key}),
				Key:    key,
				Value:  child,
				Parent: value,
			}

			if err := Walk(next, visit); err != nil {
				return err
			}
		}
	case []any:
		for i, child := range value {
			next := Node{
				Path:   append(node.Path, Segment{Index: i, IsIndex: true}),
				Key:    node.Key,
				Value:  child,
				Parent: node.Parent,
			}
			if err := Walk(next, visit); err != nil {
				return err
			}
		}
	}

	return nil
}
