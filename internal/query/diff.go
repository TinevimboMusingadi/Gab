package query

import (
	"fmt"
	"strings"

	"gab/internal/models"
	"gab/internal/storage"
)

// StateDiff represents the differences between two states
type StateDiff struct {
	Created  []FileChange
	Modified []FileChange
	Deleted  []FileChange
}

// FileChange represents a change to a file
type FileChange struct {
	Path           string
	PreArtifactID  string
	PostArtifactID string
	PreContent     []byte
	PostContent    []byte
}

// DiffEngine computes differences between states
type DiffEngine struct {
	store *storage.Store
}

// NewDiffEngine creates a new diff engine
func NewDiffEngine(store *storage.Store) *DiffEngine {
	return &DiffEngine{store: store}
}

// GetStateDiff computes the difference between two states
func (d *DiffEngine) GetStateDiff(preStateID, postStateID string) (*StateDiff, error) {
	var preState, postState models.State

	if preStateID != "" {
		if err := d.store.GetObject(preStateID, &preState); err != nil {
			return nil, fmt.Errorf("failed to load pre-state: %w", err)
		}
	}

	if postStateID != "" {
		if err := d.store.GetObject(postStateID, &postState); err != nil {
			return nil, fmt.Errorf("failed to load post-state: %w", err)
		}
	}

	diff := &StateDiff{
		Created:  make([]FileChange, 0),
		Modified: make([]FileChange, 0),
		Deleted:  make([]FileChange, 0),
	}

	preMap := make(map[string]string)
	if preState.Resources != nil {
		preMap = preState.Resources
	}

	postMap := make(map[string]string)
	if postState.Resources != nil {
		postMap = postState.Resources
	}

	// Find created and modified files
	for path, postArtID := range postMap {
		preArtID, exists := preMap[path]
		if !exists {
			// File was created
			var artifact models.Artifact
			if err := d.store.GetObject(postArtID, &artifact); err == nil {
				diff.Created = append(diff.Created, FileChange{
					Path:           path,
					PostArtifactID: postArtID,
					PostContent:    artifact.Data,
				})
			}
		} else if preArtID != postArtID {
			// File was modified
			var preArt, postArt models.Artifact
			preContent := []byte{}
			postContent := []byte{}

			if err := d.store.GetObject(preArtID, &preArt); err == nil {
				preContent = preArt.Data
			}
			if err := d.store.GetObject(postArtID, &postArt); err == nil {
				postContent = postArt.Data
			}

			diff.Modified = append(diff.Modified, FileChange{
				Path:           path,
				PreArtifactID:  preArtID,
				PostArtifactID: postArtID,
				PreContent:     preContent,
				PostContent:    postContent,
			})
		}
	}

	// Find deleted files
	for path, preArtID := range preMap {
		if _, exists := postMap[path]; !exists {
			// File was deleted
			var artifact models.Artifact
			preContent := []byte{}
			if err := d.store.GetObject(preArtID, &artifact); err == nil {
				preContent = artifact.Data
			}
			diff.Deleted = append(diff.Deleted, FileChange{
				Path:          path,
				PreArtifactID: preArtID,
				PreContent:    preContent,
			})
		}
	}

	return diff, nil
}

// FormatDiff returns a human-readable diff report
func (d *StateDiff) FormatDiff() string {
	var sb strings.Builder

	if len(d.Created) > 0 {
		sb.WriteString("Created files:\n")
		for _, change := range d.Created {
			sb.WriteString(fmt.Sprintf("  + %s (artifact: %s)\n", change.Path, change.PostArtifactID[:16]+"..."))
		}
	}

	if len(d.Modified) > 0 {
		sb.WriteString("Modified files:\n")
		for _, change := range d.Modified {
			sb.WriteString(fmt.Sprintf("  ~ %s\n", change.Path))
			sb.WriteString(fmt.Sprintf("    pre:  %s\n", change.PreArtifactID[:16]+"..."))
			sb.WriteString(fmt.Sprintf("    post: %s\n", change.PostArtifactID[:16]+"..."))
		}
	}

	if len(d.Deleted) > 0 {
		sb.WriteString("Deleted files:\n")
		for _, change := range d.Deleted {
			sb.WriteString(fmt.Sprintf("  - %s (artifact: %s)\n", change.Path, change.PreArtifactID[:16]+"..."))
		}
	}

	if len(d.Created) == 0 && len(d.Modified) == 0 && len(d.Deleted) == 0 {
		sb.WriteString("No changes detected.\n")
	}

	return sb.String()
}

// GetContentDiff returns a unified diff for a modified file
func (d *DiffEngine) GetContentDiff(change FileChange) (string, error) {
	if len(change.PreContent) == 0 && len(change.PostContent) == 0 {
		return "", nil
	}

	preLines := strings.Split(string(change.PreContent), "\n")
	postLines := strings.Split(string(change.PostContent), "\n")

	// Simple line-by-line diff
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- %s (pre)\n", change.Path))
	sb.WriteString(fmt.Sprintf("+++ %s (post)\n", change.Path))

	// For now, just show both versions
	// A proper diff algorithm (Myers, etc.) would be better
	maxLen := len(preLines)
	if len(postLines) > maxLen {
		maxLen = len(postLines)
	}

	for i := 0; i < maxLen; i++ {
		preLine := ""
		postLine := ""
		if i < len(preLines) {
			preLine = preLines[i]
		}
		if i < len(postLines) {
			postLine = postLines[i]
		}

		if preLine != postLine {
			if preLine != "" {
				sb.WriteString(fmt.Sprintf("-%s\n", preLine))
			}
			if postLine != "" {
				sb.WriteString(fmt.Sprintf("+%s\n", postLine))
			}
		} else if preLine != "" {
			sb.WriteString(fmt.Sprintf(" %s\n", preLine))
		}
	}

	return sb.String(), nil
}
