package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

func parseJSON(data []byte) (any, error) {
	decode := json.NewDecoder(bytes.NewReader(data))
	
	config, err := jsonValue(decode)
	if err != nil {
		return nil, err
	}

	if _, err := decode.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("JSON must have only one document")
	}

	return config, nil
}

func jsonValue(decode *json.Decoder) (any, error) {
	token, err := decode.Token()
	if err != nil {
		return nil, err
	}

	delim, composite := token.(json.Delim)
	if !composite {
		return token, nil
	}

	switch delim {
	case '{':
		object := make(map[string]any)
		for decode.More() {
			t, err := decode.Token()
			if err != nil {
				return nil, err
			}

			key, ok := t.(string)
			if !ok {
				return nil, err
			}

			if _, exists := object[key]; exists {
				return nil, errors.New("duplicate JSON key")
			}

			value, err := jsonValue(decode)
			if err != nil {
				return nil, err
			}
			object[key] = value
		}

		if t, err := decode.Token(); err != nil || t != json.Delim('}') {
			return nil, err
		}

		return object, nil
	case '[':
		array := make([]any, 0)
		for decode.More() {
			value, err := jsonValue(decode)
			if err != nil {
				return nil, err
			}
			array = append(array, value)
		}

		if t, err := decode.Token(); err != nil || t != json.Delim(']') {
			return nil, err
		}

		return array, nil
	default:
		return nil, err
	}
}