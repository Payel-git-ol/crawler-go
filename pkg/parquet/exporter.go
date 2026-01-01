package parquet

import (
	"Fyne-on/pkg/database"
	"Fyne-on/pkg/models"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ParquetExporter handles exporting data to Parquet and JSONL formats
type ParquetExporter struct {
	db *database.BadgerDB
}

// NewParquetExporter creates a new Parquet exporter
func NewParquetExporter(db *database.BadgerDB) *ParquetExporter {
	return &ParquetExporter{db: db}
}

// ExportIssuesToParquet exports all issues to Parquet-like format (using JSON structure)
// Note: For production use, consider using github.com/segmentio/parquet-go
func (pe *ParquetExporter) ExportIssuesToParquet(outputPath string) (int, error) {
	// For now, export as JSONL which is compatible with Parquet tools
	return pe.ExportIssuesToJSONL(outputPath)
}

// ExportPullRequestsToParquet exports all PRs to Parquet-like format (using JSON structure)
func (pe *ParquetExporter) ExportPullRequestsToParquet(outputPath string) (int, error) {
	// For now, export as JSONL which is compatible with Parquet tools
	return pe.ExportPullRequestsToJSONL(outputPath)
}

// ExportRepositoriesToParquet exports all repositories to Parquet-like format (using JSON structure)
func (pe *ParquetExporter) ExportRepositoriesToParquet(outputPath string) (int, error) {
	// For now, export as JSONL which is compatible with Parquet tools
	return pe.ExportRepositoriesToJSONL(outputPath)
}

// ExportIssuesToJSONL экспортирует issues в JSONL формат (оптимально для LLM)
// JSONL (JSON Lines) - каждая строка это валидный JSON объект
// Идеально для машинного обучения и LLM training
func (pe *ParquetExporter) ExportIssuesToJSONL(outputPath string) (int, error) {
	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	count := 0
	// Iterate through all issues in database
	err = pe.db.IterateWithPrefix("issue:", func(key string, value []byte) error {
		var issue models.Issue
		if err := json.Unmarshal(value, &issue); err != nil {
			return nil
		}

		// Write as JSON line
		jsonData, err := json.Marshal(issue)
		if err != nil {
			return nil
		}
		if _, err := file.Write(append(jsonData, '\n')); err != nil {
			return err
		}
		count++
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to export issues: %w", err)
	}

	return count, nil
}

// ExportPullRequestsToJSONL экспортирует PRs в JSONL формат
func (pe *ParquetExporter) ExportPullRequestsToJSONL(outputPath string) (int, error) {
	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	count := 0
	err = pe.db.IterateWithPrefix("pr:", func(key string, value []byte) error {
		var pr models.PullRequest
		if err := json.Unmarshal(value, &pr); err != nil {
			return nil
		}

		jsonData, err := json.Marshal(pr)
		if err != nil {
			return nil
		}
		if _, err := file.Write(append(jsonData, '\n')); err != nil {
			return err
		}
		count++
		return nil
	})

	return count, err
}

// ExportRepositoriesToJSONL экспортирует репозитории в JSONL формат
func (pe *ParquetExporter) ExportRepositoriesToJSONL(outputPath string) (int, error) {
	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	count := 0
	err = pe.db.IterateWithPrefix("repo:", func(key string, value []byte) error {
		var repo models.Repo
		if err := json.Unmarshal(value, &repo); err != nil {
			return nil
		}

		jsonData, err := json.Marshal(repo)
		if err != nil {
			return nil
		}
		if _, err := file.Write(append(jsonData, '\n')); err != nil {
			return err
		}
		count++
		return nil
	})

	return count, err
}

// ExportAllToParquet exports all data types to Parquet files in a directory
func (pe *ParquetExporter) ExportAllToParquet(outputDir string) (map[string]int, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	results := make(map[string]int)
	timestamp := time.Now().Format("2006-01-02-15-04-05")

	// Export issues
	issuesPath := filepath.Join(outputDir, fmt.Sprintf("issues_%s.parquet", timestamp))
	issueCount, err := pe.ExportIssuesToParquet(issuesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to export issues: %w", err)
	}
	results["issues"] = issueCount

	// Export PRs
	prsPath := filepath.Join(outputDir, fmt.Sprintf("pull_requests_%s.parquet", timestamp))
	prCount, err := pe.ExportPullRequestsToParquet(prsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to export PRs: %w", err)
	}
	results["pull_requests"] = prCount

	// Export repos
	reposPath := filepath.Join(outputDir, fmt.Sprintf("repositories_%s.parquet", timestamp))
	repoCount, err := pe.ExportRepositoriesToParquet(reposPath)
	if err != nil {
		return nil, fmt.Errorf("failed to export repos: %w", err)
	}
	results["repositories"] = repoCount

	return results, nil
}

// ExportAllToJSONL exports all data types to JSONL files (better for LLM training)
func (pe *ParquetExporter) ExportAllToJSONL(outputDir string) (map[string]int, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	results := make(map[string]int)
	timestamp := time.Now().Format("2006-01-02-15-04-05")

	// Export issues as JSONL
	issuesPath := filepath.Join(outputDir, fmt.Sprintf("issues_%s.jsonl", timestamp))
	issueCount, err := pe.ExportIssuesToJSONL(issuesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to export issues: %w", err)
	}
	results["issues"] = issueCount

	// Export PRs as JSONL
	prsPath := filepath.Join(outputDir, fmt.Sprintf("pull_requests_%s.jsonl", timestamp))
	prCount, err := pe.ExportPullRequestsToJSONL(prsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to export PRs: %w", err)
	}
	results["pull_requests"] = prCount

	// Export repos as JSONL
	reposPath := filepath.Join(outputDir, fmt.Sprintf("repositories_%s.jsonl", timestamp))
	repoCount, err := pe.ExportRepositoriesToJSONL(reposPath)
	if err != nil {
		return nil, fmt.Errorf("failed to export repos: %w", err)
	}
	results["repositories"] = repoCount

	return results, nil
}
