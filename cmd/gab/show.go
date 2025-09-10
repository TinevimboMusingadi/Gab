package main

import (
    "encoding/json"
    "fmt"
    "gab/internal/models"
    "gab/internal/storage"
)

import "github.com/spf13/cobra"

func newShowCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "show <event_hash>",
        Short: "Show full JSON for an event",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            store, err := storage.Open(dataDir)
            if err != nil { return err }
            defer store.Close()
            var ev models.Event
            if err := store.GetObject(args[0], &ev); err != nil { return err }
            out, _ := json.MarshalIndent(ev, "", "  ")
            fmt.Println(string(out))
            return nil
        },
    }
    return cmd
}


