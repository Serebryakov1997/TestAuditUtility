package restapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/serebryakov1997/utility/audit"
)

var ErrInvalidConfig = errors.New("invalid configuration")

type AuditFunc func(
	ctx context.Context,
	data []byte,
	format string,
) ([]audit.Finding, error)

func NewHandler(auditFn AuditFunc) (http.Handler, error) {
	if auditFn == nil {
		return nil, errors.New("restapi: audit function is required")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/audit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			writeError(w, http.StatusUnsupportedMediaType,
				"use application/json or application/yaml")
			return
		}

		var format string
		switch mediaType {
		case "application/json":
			format = "json"
		case "application/yaml", "application/x-yaml", "text/yaml":
			format = "yaml"
		default:
			writeError(w, http.StatusUnsupportedMediaType,
				"use application/json or application/yaml")
			return
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "cannot read request body")
			return
		}

		if len(bytes.TrimSpace(data)) == 0 {
			writeError(w, http.StatusBadRequest, "configuration is empty")
			return
		}

		findings, err := auditFn(r.Context(), data, format)
		if err != nil {
			if errors.Is(err, ErrInvalidConfig) {
				writeError(w, http.StatusBadRequest, "invalid configuration")
			} else {
				writeError(w, http.StatusInternalServerError, "audit failed")
			}
			return
		}

		if findings == nil {
			findings = make([]audit.Finding, 0)
		}
		writeJSON(w, http.StatusOK, findings)
	})
	return mux, nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		status = http.StatusInternalServerError
		data = []byte(`{"error":"cannot encode response"}`)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	_, _ = w.Write(data)
}
