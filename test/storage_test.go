package test

import (
    "os"
    "path/filepath"
    "testing"
    "gab/internal/storage"
)

func TestStoreRoundTrip(t *testing.T) {
    dir := t.TempDir()
    if err := storage.Init(dir); err != nil { t.Fatal(err) }
    st, err := storage.Open(dir)
    if err != nil { t.Fatal(err) }
    defer st.Close()
    sample := struct{ A string; B int }{"x", 42}
    id := "deadbeef"
    if err := st.PutObject(id, &sample); err != nil { t.Fatal(err) }
    var out struct{ A string; B int }
    if err := st.GetObject(id, &out); err != nil { t.Fatal(err) }
    if out.A != "x" || out.B != 42 { t.Fatalf("unexpected: %#v", out) }
    // Ensure DB files exist
    if _, err := os.Stat(filepath.Join(dir, "objects")); err != nil { t.Fatal(err) }
}


