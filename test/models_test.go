package test

import (
    "testing"
    "gab/internal/models"
)

func TestArtifactIDDeterminism(t *testing.T) {
    a := models.Artifact{ContentType: "file", Data: []byte("hello")}
    a.ComputeID()
    b := models.Artifact{ContentType: "file", Data: []byte("hello")}
    b.ComputeID()
    if a.ID != b.ID { t.Fatalf("expected same id, got %s vs %s", a.ID, b.ID) }
}

func TestStateIDDeterminism(t *testing.T) {
    s1 := models.State{Resources: map[string]string{"a": "1", "b": "2"}}
    s1.ComputeID()
    s2 := models.State{Resources: map[string]string{"b": "2", "a": "1"}}
    s2.ComputeID()
    if s1.ID != s2.ID { t.Fatalf("expected same id, got %s vs %s", s1.ID, s2.ID) }
}


