package deduce

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gab/internal/models"
)

// Writer appends EDB facts (Datalog-like) to a file for later ingestion.
// Facts are written one per line, ending with a period.
type Writer struct {
	factsPath string
}

func New(dataDir string) (*Writer, error) {
	ddir := filepath.Join(dataDir, "deduce")
	if err := os.MkdirAll(ddir, 0o755); err != nil {
		return nil, err
	}
	return &Writer{factsPath: filepath.Join(ddir, "facts.dl")}, nil
}

func (w *Writer) appendLine(line string) error {
	f, err := os.OpenFile(w.factsPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if !strings.HasSuffix(line, ".") {
		line += "."
	}
	_, err = fmt.Fprintln(f, line)
	return err
}

// EmitEventFacts writes core facts derived from an event and pre/post states.
func (w *Writer) EmitEventFacts(agentID string, ev *models.Event, pre *models.State, post *models.State, resultArtifacts map[string]string, finishedAt time.Time) error {
	ts := finishedAt.Unix()
	// Event facts
	if err := w.appendLine(fmt.Sprintf("Event_happened(\"%s\", \"%s\", \"%s\", %d)", ev.ID, agentID, ev.Action.Tool, ts)); err != nil {
		return err
	}
	if ev.ParentEventHash != "" {
		if err := w.appendLine(fmt.Sprintf("Parent_of(\"%s\", \"%s\")", ev.ID, ev.ParentEventHash)); err != nil {
			return err
		}
	}
	if cmd, ok := ev.Action.Params["command"]; ok {
		if err := w.appendLine(fmt.Sprintf("Action_executed(\"%s\", \"%s\")", ev.ID, escape(cmd))); err != nil {
			return err
		}
	}
	// Result artifacts (stdout, stderr, exit_code)
	for kind, art := range resultArtifacts {
		if err := w.appendLine(fmt.Sprintf("Action_produced_output(\"%s\", \"%s\", \"%s\")", ev.ID, art, kind)); err != nil {
			return err
		}
	}

	// Diff resources for created/modified/deleted
	preMap := map[string]string{}
	postMap := map[string]string{}
	if pre != nil && pre.Resources != nil {
		for k, v := range pre.Resources {
			preMap[k] = v
		}
	}
	if post != nil && post.Resources != nil {
		for k, v := range post.Resources {
			postMap[k] = v
		}
	}

	// Created or modified
	keys := make([]string, 0, len(postMap))
	for k := range postMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		newID := postMap[k]
		if oldID, ok := preMap[k]; !ok {
			if err := w.appendLine(fmt.Sprintf("File_created(\"%s\", \"%s\", \"%s\")", ev.ID, escape(k), newID)); err != nil {
				return err
			}
		} else if oldID != newID {
			if err := w.appendLine(fmt.Sprintf("File_modified(\"%s\", \"%s\", \"%s\", \"%s\")", ev.ID, escape(k), oldID, newID)); err != nil {
				return err
			}
		}
	}
	// Deleted
	keys = keys[:0]
	for k := range preMap {
		if _, ok := postMap[k]; !ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		if err := w.appendLine(fmt.Sprintf("File_deleted(\"%s\", \"%s\", \"%s\")", ev.ID, escape(k), preMap[k])); err != nil {
			return err
		}
	}

	return nil
}

func escape(s string) string {
	// Minimal string escaping for quotes and backslashes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}
