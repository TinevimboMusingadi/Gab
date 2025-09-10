package index

import (
    "encoding/json"
    "errors"
    "path/filepath"
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
    return i.db.Update(func(txn *badger.Txn) error { return txn.Set(key, []byte("1")) })
}


