package audit

// Node is representation of the current node
// for the duration of the Check call
type Node struct {
	Path Path
	Key string
	Value any
	Parent map[string]any
}

type Rule interface {
	ID() string
	Check(Node) []Finding
}