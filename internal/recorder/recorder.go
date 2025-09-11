package recorder

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"time"

	"gab/internal/hashutil"
	"gab/internal/index"
	"gab/internal/models"
	"gab/internal/storage"
	"gab/internal/graph"

	"os"
	"io/fs"
)

type Recorder struct {
	store *storage.Store
	idx   *index.Index
	sink  graph.Sink
}

func New(store *storage.Store, idx *index.Index) *Recorder { return &Recorder{store: store, idx: idx, sink: graph.NewNoopSink()} }

func WithSink(r *Recorder, s graph.Sink) *Recorder { r.sink = s; return r }

type RecordCommandInput struct {
	AgentID string
	ScopeDir string
	Cwd string
	Command string
	Args []string
	Env []string
}

type RecordedEvent struct { models.Event }

func isExcluded(path string, excludes []string) bool {
	p := filepath.Clean(path)
	for _, ex := range excludes {
		if ex == "" { continue }
		e := filepath.Clean(ex)
		if p == e { return true }
		if len(p) > len(e) && p[:len(e)] == e {
			if p[len(e)] == filepath.Separator { return true }
		}
	}
	return false
}

func snapshotDir(scope string, excludes []string, st *storage.Store) (models.State, error) {
	resources := map[string]string{}
	err := filepath.WalkDir(scope, func(path string, d fs.DirEntry, err error) error {
		if err != nil { return err }
		if isExcluded(path, excludes) {
			if d.IsDir() { return fs.SkipDir }
			return nil
		}
		if d.IsDir() { return nil }
		data, err := os.ReadFile(path)
		if err != nil { return nil }
		art := models.Artifact{ContentType: "file", Data: data}
		art.ComputeID()
		if err := st.PutObject(art.ID, &art); err != nil { return err }
		rel, _ := filepath.Rel(scope, path)
		resources[filepath.ToSlash(rel)] = art.ID
		return nil
	})
	if err != nil { return models.State{}, err }
	s := models.State{Resources: resources}
	s.ComputeID()
	if err := st.PutObject(s.ID, &s); err != nil { return models.State{}, err }
	return s, nil
}

func (r *Recorder) RecordCommand(ctx context.Context, in RecordCommandInput) (*RecordedEvent, error) {
	// Default excludes to avoid ballooning artifacts: .git and gab_data under scope
	defaultExcludes := []string{
		filepath.Join(in.ScopeDir, ".git"),
		filepath.Join(in.ScopeDir, "gab_data"),
	}
	preState, err := snapshotDir(in.ScopeDir, defaultExcludes, r.store)
	if err != nil { return nil, err }

	start := time.Now().UTC()
	cmd := exec.CommandContext(ctx, in.Command, in.Args...)
	cmd.Dir = in.Cwd
	cmd.Env = in.Env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	finish := time.Now().UTC()

	// Result artifacts
	stdoutArt := models.Artifact{ContentType: "stdout", Data: stdout.Bytes()}
	stdoutArt.ComputeID()
	_ = r.store.PutObject(stdoutArt.ID, &stdoutArt)
	stderrArt := models.Artifact{ContentType: "stderr", Data: stderr.Bytes()}
	stderrArt.ComputeID()
	_ = r.store.PutObject(stderrArt.ID, &stderrArt)
	exitCode := 0
	if runErr != nil { if ee, ok := runErr.(*exec.ExitError); ok { exitCode = ee.ExitCode() } else { exitCode = -1 } }
	exitBytes, _ := json.Marshal(exitCode)
	exitArt := models.Artifact{ContentType: "exit_code", Data: exitBytes}
	exitArt.ComputeID()
	_ = r.store.PutObject(exitArt.ID, &exitArt)

	postState, err := snapshotDir(in.ScopeDir, defaultExcludes, r.store)
	if err != nil { return nil, err }

	// Build event
	action := models.ActionDescriptor{
		Type: "EXECUTE_COMMAND",
		Params: map[string]string{
			"command": in.Command,
			"args_hash": hashutil.HashStrings(in.Args),
			"cwd": in.Cwd,
		},
		AgentID: in.AgentID,
		Tool: "gab-cli",
		StartedAtUnix: start.Unix(),
		FinishedAtUnix: finish.Unix(),
	}

	head, _ := r.idx.GetHead(in.AgentID)
	ev := models.Event{
		ParentEventHash: head,
		PreActionStateHash: preState.ID,
		PostActionStateHash: postState.ID,
		Action: action,
		ResultingArtifactHashes: []string{stdoutArt.ID, stderrArt.ID, exitArt.ID},
	}
	_ = ev.ComputeID()
	if err := r.store.PutObject(ev.ID, &ev); err != nil { return nil, err }
	if err := r.idx.SetHead(in.AgentID, ev.ID); err != nil { return nil, err }
	if err := r.idx.AddEventTime(in.AgentID, ev.ID, finish); err != nil { return nil, err }
	_ = r.sink.WriteEvent(in.AgentID, &ev, &preState, &postState)
	return &RecordedEvent{Event: ev}, nil
}


