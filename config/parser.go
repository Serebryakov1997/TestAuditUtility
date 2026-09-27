package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

const (
	Auto string = "auto"
	JSON string = "json"
	YAML string = "yaml"
)

func Parse(r io.Reader, format string) (any, error) {
	if format != Auto && format != JSON && format != YAML {
		return nil, errors.New("unsupported format")
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, errors.New("empty config")
	}

	if format == Auto {
		if data[0] == '{' || data[0] == '[' {
			format = JSON
		} else {
			format = YAML
		}
	}

	var parsingData any
	if format == JSON {
		parsingData, err = parseJSON(data)
	}
	if format == YAML {
		parsingData, err = parseYAML(data)
	}

	if err != nil {
		return nil, err
	}

	switch parsingData.(type) {
	case map[string]any, []any:
		return parsingData, nil
	default:
		return nil, errors.New("config data must be a map or array")
	}
}
