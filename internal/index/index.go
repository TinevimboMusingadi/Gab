package index

import (
    "encoding/json"
    "errors"
    "path/filepath"
    "strings"
    "time"

    badger "github.com/dgraph-io/badger/v4"
)

// Keys:
// head:<agentID> -> eventHash
// time:<agentID>:<reverse_ts>:<eventHash> -> 1 (for listing by time)

type Index struct { db *badger.DB }

func Open(dataDir string) (*Index, error) {
    opts := badger.DefaultOptions(filepath.Join(dataDir, "index")).WithLogger(nil)
    db, err := badger.Open(opts)
    if err != nil { return nil, err }
    return &Index{db: db}, nil
}

func (i *Index) Close() error { return i.db.Close() }

func (i *Index) SetHead(agentID, eventHash string) error {
    key := []byte("head:" + agentID)
    return i.db.Update(func(txn *badger.Txn) error { return txn.Set(key, []byte(eventHash)) })
}

func (i *Index) GetHead(agentID string) (string, error) {
    var out string
    err := i.db.View(func(txn *badger.Txn) error {
        item, err := txn.Get([]byte("head:" + agentID))
        if errors.Is(err, badger.ErrKeyNotFound) { return nil }
        if err != nil { return err }
        return item.Value(func(val []byte) error { out = string(val); return nil })
    })
    return out, err
}

func reverseTimestamp(ts time.Time) string {
    // Use UnixNano reversed for lexicographic descending order
    // Max int64 ~ 9,223,372,036,854,775,807
    const max = int64(9223372036854775807)
    v := max - ts.UnixNano()
    b, _ := json.Marshal(v) // stable decimal encoding
    return string(b)
}

func (i *Index) AddEventTime(agentID, eventHash string, ts time.Time) error {
    key := []byte("time:" + agentID + ":" + reverseTimestamp(ts) + ":" + eventHash)
    val, _ := json.Marshal(ts.UnixNano())
    return i.db.Update(func(txn *badger.Txn) error { return txn.Set(key, val) })
}

type TimeEntry struct {
    EventHash string
    Time      time.Time
}

func (i *Index) ListByTime(agentID string, limit int) ([]TimeEntry, error) {
    prefix := []byte("time:" + agentID + ":")
    entries := make([]TimeEntry, 0, limit)
    err := i.db.View(func(txn *badger.Txn) error {
        it := txn.NewIterator(badger.IteratorOptions{PrefetchValues: true, Prefix: prefix})
        defer it.Close()
        for it.Rewind(); it.ValidForPrefix(prefix); it.Next() {
            item := it.Item()
            k := string(item.Key())
            // key format: time:<agent>:<rev_ts>:<hash>
            // event hash is after last ':'
            idx := strings.LastIndex(k, ":")
            if idx <= 0 { continue }
            hash := k[idx+1:]
            var nsec int64
            if err := item.Value(func(v []byte) error { return json.Unmarshal(v, &nsec) }); err != nil { return err }
            entries = append(entries, TimeEntry{EventHash: hash, Time: time.Unix(0, nsec).UTC()})
            if limit > 0 && len(entries) >= limit { break }
        }
        return nil
    })
    return entries, err
}


