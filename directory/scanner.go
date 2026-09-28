package directory

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/serebryakov1997/utility/audit"
	"github.com/serebryakov1997/utility/config"
)

type Analyzer interface {
	Analyze(config any) ([]audit.Finding, error)
}

type FileReport struct {
	Path     string          `json:"path"`
	Findings []audit.Finding `json:"findings"`
	Error    string          `json:"error,omitempty"`
}

type Scanner struct {
	analyzer Analyzer
}

func New(analyzer Analyzer) (*Scanner, error) {
	if analyzer == nil {
		return nil, errors.New("analyzer must not be nil")
	}

	return &Scanner{
		analyzer: analyzer,
	}, nil
}

func (s *Scanner) Analyze(directoryPath string) ([]FileReport, error) {
	if strings.TrimSpace(directoryPath) == "" {
		return nil, errors.New("directory path is empty")
	}

	info, err := os.Stat(directoryPath)
	if err != nil {
		return nil, fmt.Errorf("stat directory %q: %w", directoryPath, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", directoryPath)
	}

	reports := make([]FileReport, 0)

	err = fs.WalkDir(
		os.DirFS(directoryPath),
		".",
		func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return fmt.Errorf("walk %q: %w", path, walkErr)
			}

			if entry.IsDir() {
				return nil
			}

			if entry.Type()&os.ModeSymlink != 0 {
				return nil
			}

			format, ok := formatByPath(path)
			if !ok {
				return nil
			}

			report, err := s.analyzeFile(
				directoryPath,
				path,
				format,
			)

			if err != nil {
				reports = append(reports, FileReport{
					Path:     reportPath(directoryPath, path),
					Findings: make([]audit.Finding, 0),
					Error:    err.Error(),
				})

				return nil
			}

			reports = append(reports, report)
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return reports, nil
}

func (s *Scanner) analyzeFile(
	rootPath string,
	filePath string,
	format string,
) (FileReport, error) {
	fullPath := filepath.Join(rootPath, filePath)

	file, err := os.Open(fullPath)
	if err != nil {
		return FileReport{
			Path: reportPath(rootPath, filePath),
		}, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	parseConfig, err := config.Parse(file, format)
	if err != nil {
		return FileReport{
			Path: reportPath(rootPath, filePath),
		}, fmt.Errorf("parse file: %w", err)
	}

	findings, err := s.analyzer.Analyze(parseConfig)
	if err != nil {
		return FileReport{
			Path: reportPath(rootPath, filePath),
		}, fmt.Errorf("analyze file: %w", err)
	}

	if findings == nil {
		findings = make([]audit.Finding, 0)
	}

	return FileReport{
		Path:     reportPath(rootPath, filePath),
		Findings: findings,
	}, nil
}

func formatByPath(path string) (string, bool) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return config.JSON, true
	case ".yaml", ".yml":
		return config.YAML, true
	default:
		return "", false
	}
}

func reportPath(rootPath, filePath string) string {
	relativePath, err := filepath.Rel(rootPath, filepath.Join(rootPath, filePath))
	if err != nil {
		return filepath.ToSlash(filePath)
	}

	return filepath.ToSlash(relativePath)
}
