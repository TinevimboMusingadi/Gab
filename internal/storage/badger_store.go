package storage

import (
    "bytes"
    "compress/zlib"
    "encoding/json"
    "fmt"
    "io"
    "path/filepath"

    badger "github.com/dgraph-io/badger/v4"
)

const (
    objectsDir = "objects"
)

type Store struct {
    db *badger.DB
}

func Init(dataDir string) error {
    opts := badger.DefaultOptions(filepath.Join(dataDir, objectsDir))
    _, err := badger.Open(opts.WithLogger(nil))
    if err == nil {
        // Close immediately; create folders/manifest
        // Note: open-test to create DB
        _ = badger.DefaultOptions("")
    }
    // We return nil if open succeeded
    if err == nil {
        // Close db properly by re-opening with Open, defer Close
        db, err2 := badger.Open(opts.WithLogger(nil))
        if err2 != nil { return err2 }
        defer db.Close()
        return nil
    }
    return err
}

func Open(dataDir string) (*Store, error) {
    opts := badger.DefaultOptions(filepath.Join(dataDir, objectsDir)).WithLogger(nil)
    db, err := badger.Open(opts)
    if err != nil { return nil, err }
    return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// PutObject stores an object under its ID hash. The object must have field ID string.
func (s *Store) PutObject(id string, v any) error {
    raw, err := json.Marshal(v)
    if err != nil { return err }
    var buf bytes.Buffer
    zw := zlib.NewWriter(&buf)
    if _, err := zw.Write(raw); err != nil { return err }
    if err := zw.Close(); err != nil { return err }
    return s.db.Update(func(txn *badger.Txn) error {
        return txn.Set([]byte(id), buf.Bytes())
    })
}

func (s *Store) GetObject(id string, out any) error {
    return s.db.View(func(txn *badger.Txn) error {
        item, err := txn.Get([]byte(id))
        if err != nil { return err }
        return item.Value(func(val []byte) error {
            zr, err := zlib.NewReader(bytes.NewReader(val))
            if err != nil { return err }
            defer zr.Close()
            dec, err := io.ReadAll(zr)
            if err != nil { return err }
            if err := json.Unmarshal(dec, out); err != nil { return fmt.Errorf("unmarshal %s: %w", id, err) }
            return nil
        })
    })
}


