package deduce

import (
    "os"
    "path/filepath"
)

const defaultRules = `
// Example Mangle/Datalog-like rules (placeholder syntax)
// Ancestor_of via transitive closure
Ancestor_of(A, D) :- Parent_of(D, A).
Ancestor_of(A, D) :- Parent_of(D, M), Ancestor_of(A, M).

// Taint propagation from Secret_content to outputs of the same event
Tainted_content(C) :- Content_read_by_event(SC, E), Secret_content(SC), Action_produced_output(E, A, _), Artifact_is_file(A, _, C).

// Potential exfiltration when tainted content is read and network call is made
Potential_exfiltration(E, C, IP) :- Action_made_network_call(E, IP, _), Content_read_by_event(C, E), Tainted_content(C).
`

// InstallDefaultRules writes starter rules into dataDir/deduce/rules.dl if not present.
func InstallDefaultRules(dataDir string) (string, error) {
    ddir := filepath.Join(dataDir, "deduce")
    if err := os.MkdirAll(ddir, 0o755); err != nil { return "", err }
    path := filepath.Join(ddir, "rules.dl")
    if _, err := os.Stat(path); err == nil { return path, nil }
    if err := os.WriteFile(path, []byte(defaultRules), 0o644); err != nil { return "", err }
    return path, nil
}


