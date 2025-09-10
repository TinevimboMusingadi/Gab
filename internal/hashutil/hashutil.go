package hashutil

import (
    "crypto/sha256"
    "encoding/hex"
    "sort"
)

// HashBytes returns sha256 hash hex of input bytes
func HashBytes(b []byte) string {
    sum := sha256.Sum256(b)
    return hex.EncodeToString(sum[:])
}

// HashStrings deterministically hashes an ordered list of strings using sha256
func HashStrings(parts []string) string {
    h := sha256.New()
    for _, p := range parts {
        h.Write([]byte{0})
        h.Write([]byte(p))
    }
    return hex.EncodeToString(h.Sum(nil))
}

// CanonicalizeMap returns stable key order and concatenated representation "k\x00v" per key.
func CanonicalizeMap(m map[string]string) []string {
    keys := make([]string, 0, len(m))
    for k := range m { keys = append(keys, k) }
    sort.Strings(keys)
    out := make([]string, 0, len(keys))
    for _, k := range keys {
        out = append(out, k+"\x00"+m[k])
    }
    return out
}


