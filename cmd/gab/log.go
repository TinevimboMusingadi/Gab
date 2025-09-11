package main

import (
    "fmt"
    "gab/internal/index"
    "gab/internal/models"
    "gab/internal/storage"
    "time"
)

import "github.com/spf13/cobra"

func newLogCmd() *cobra.Command {
    var agentID string
    cmd := &cobra.Command{
        Use:   "log --agent <id>",
        Short: "Show event history for an agent",
        RunE: func(cmd *cobra.Command, args []string) error {
            if agentID == "" { return fmt.Errorf("--agent is required") }
            store, err := storage.Open(dataDir)
            if err != nil { return err }
            defer store.Close()
            idx, err := index.Open(dataDir)
            if err != nil { return err }
            defer idx.Close()

            head, err := idx.GetHead(agentID)
            if err != nil { return err }
            if head == "" {
                fmt.Println("No history.")
                return nil
            }
            // Walk backwards
            current := head
            for current != "" {
                var ev models.Event
                if err := store.GetObject(current, &ev); err != nil { return err }
                ts := time.Unix(ev.Action.FinishedAtUnix, 0).UTC().Format("2006-01-02T15:04:05Z07:00")
                fmt.Printf("%s %s %s %s %s\n", ts, ev.ID, ev.Action.Type, ev.Action.Tool, ev.Action.Params["command"])
                current = ev.ParentEventHash
            }
            return nil
        },
    }
    cmd.Flags().StringVar(&agentID, "agent", "", "Agent identifier")
    return cmd
}


