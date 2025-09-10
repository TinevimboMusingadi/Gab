package models

import (
    "encoding/json"
    "gab/internal/hashutil"
)

type Artifact struct {
    ContentType string `json:"content_type"`
    Data        []byte `json:"data"`
    ID          string `json:"id"`
}

func (a *Artifact) ComputeID() {
    // Hash over content type + null + data bytes
    parts := []byte(a.ContentType)
    parts = append(parts, 0)
    parts = append(parts, a.Data...)
    a.ID = hashutil.HashBytes(parts)
}

type State struct {
    Resources map[string]string `json:"resources"`
    ID        string            `json:"id"`
}

func (s *State) ComputeID() {
    if s.Resources == nil { s.Resources = map[string]string{} }
    canonical := hashutil.CanonicalizeMap(s.Resources)
    s.ID = hashutil.HashStrings(append([]string{"State"}, canonical...))
}

type ActionDescriptor struct {
    Type    string            `json:"type"`
    Params  map[string]string `json:"params"`
    AgentID string            `json:"agent_id"`
    Tool    string            `json:"tool"`
    StartedAtUnix  int64      `json:"started_at_unix"`
    FinishedAtUnix int64      `json:"finished_at_unix"`
}

type Event struct {
    ParentEventHash    string           `json:"parent_event_hash"`
    PreActionStateHash string           `json:"pre_action_state_hash"`
    PostActionStateHash string          `json:"post_action_state_hash"`
    Action             ActionDescriptor `json:"action"`
    ResultingArtifactHashes []string    `json:"resulting_artifact_hashes"`
    ID                 string           `json:"id"`
}

func (e *Event) ComputeID() error {
    // Stable JSON encoding (Go map order is random; we avoid maps or pre-sort)
    // Marshal copy without ID field populated
    copy := *e
    copy.ID = ""
    b, err := json.Marshal(copy)
    if err != nil { return err }
    e.ID = hashutil.HashBytes(b)
    return nil
}


