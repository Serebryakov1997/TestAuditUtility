package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

func parseYAML(data []byte) (any, error) {
	decode := yaml.NewDecoder(bytes.NewReader(data))

	var document yaml.Node
	if err := decode.Decode(&document); err != nil {
		return nil, errors.New("uncorrected or empty YAML")
	}

	var config yaml.Node
	if err := decode.Decode(&config); !errors.Is(err, io.EOF) {
		return nil, errors.New("YAML must have only one document")
	}

	if len(document.Content) != 1 {
		return nil, errors.New("empty YAML document")
	}

	return yamlValue(document.Content[0])
}

func yamlValue(n *yaml.Node) (any, error) {
	switch n.Kind {
	case yaml.MappingNode:
		if n.Tag != "!!map" {
			return nil, invalidYAML(n)
		}

		object := make(map[string]any, len(n.Content)/2)
		for i := 0; i < len(n.Content); i += 2 {
			key := n.Content[i]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
				return nil, fmt.Errorf("key must be a string: string %d", key.Line)
			}

			if _, exists := object[key.Value]; exists {
				return nil, fmt.Errorf("duplicate key of YAML: string %d", key.Line)
			}

			value, err := yamlValue(n.Content[i+1])
			if err != nil {
				return nil, err
			}

			object[key.Value] = value
		}
		return object, nil
	case yaml.SequenceNode:
		if n.Tag != "!!seq" {
			return nil, invalidYAML(n)
		}

		array := make([]any, 0, len(n.Content))

		for _, child := range n.Content {
			value, err := yamlValue(child)
			if err != nil {
				return nil, err
			}

			array = append(array, value)
		}
		return array, nil
	case yaml.ScalarNode:
		switch n.Tag {
		case "!!str", "!!timestamp":
			return n.Value, nil
		case "!!null":
			return nil, nil
		case "!!bool":
			var value bool
			if err := n.Decode(&value); err != nil {
				return nil, invalidYAML(n)
			}
			return value, nil
		case "!!int", "!!float":
			var value any
			if err := n.Decode(&value); err != nil {
				return nil, invalidYAML(n)
			}

			number := json.Number(fmt.Sprint(value))
			if _, err := json.Marshal(number); err != nil {
				return nil, invalidYAML(n)
			}

			return number, nil
		default:
			return nil, invalidYAML(n)
		}
	case yaml.AliasNode:
		return nil, fmt.Errorf("YAML aliases not supported: string %d", n.Line)
	default:
		return nil, invalidYAML(n)
	}
}

func invalidYAML(n *yaml.Node) error {
	return fmt.Errorf("invalid value of YAML: string %d", n.Line)
}