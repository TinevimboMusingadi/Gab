package main

import (
    "fmt"
    "gab/internal/index"
)

import "github.com/spf13/cobra"

func newTimelineCmd() *cobra.Command {
    var (
        agentID string
        limit int
    )
    cmd := &cobra.Command{
        Use:   "timeline --agent <id>",
        Short: "List recent events for an agent in time order",
        RunE: func(cmd *cobra.Command, args []string) error {
            idx, err := index.Open(dataDir)
            if err != nil { return err }
            defer idx.Close()
            entries, err := idx.ListByTime(agentID, limit)
            if err != nil { return err }
            for _, e := range entries {
                fmt.Printf("%s %s\n", e.Time.Format("2006-01-02T15:04:05Z07:00"), e.EventHash)
            }
            return nil
        },
    }
    cmd.Flags().StringVar(&agentID, "agent", "", "Agent identifier")
    cmd.Flags().IntVar(&limit, "limit", 20, "Max entries to show")
    _ = cmd.MarkFlagRequired("agent")
    return cmd
}


