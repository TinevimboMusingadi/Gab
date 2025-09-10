package main

import (
    "context"
    "fmt"
    "os"
    "strings"
    "time"

    "gab/internal/index"
    "gab/internal/recorder"
    "gab/internal/storage"
)

import "github.com/spf13/cobra"

func newRecordCmd() *cobra.Command {
    var (
        agentID string
        scopeDir string
        cwd string
    )

    cmd := &cobra.Command{
        Use:   "record-cmd --agent <id> --scope <dir> --cwd <dir> -- <command> [args...]",
        Short: "Record execution of a shell command with pre/post state snapshot",
        Args:  cobra.ArbitraryArgs,
        RunE: func(cmd *cobra.Command, args []string) error {
            if agentID == "" { return fmt.Errorf("--agent is required") }
            if scopeDir == "" { return fmt.Errorf("--scope is required") }
            if cwd == "" { cwd = "." }
            if len(args) == 0 { return fmt.Errorf("command is required after --") }

            store, err := storage.Open(dataDir)
            if err != nil { return err }
            defer store.Close()

            idx, err := index.Open(dataDir)
            if err != nil { return err }
            defer idx.Close()

            rec := recorder.New(store, idx)
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
            defer cancel()

            ev, err := rec.RecordCommand(ctx, recorder.RecordCommandInput{
                AgentID: agentID,
                ScopeDir: scopeDir,
                Cwd: cwd,
                Command: args[0],
                Args: args[1:],
                Env: os.Environ(),
            })
            if err != nil { return err }

            fmt.Printf("Recorded event %s for agent %s\n", ev.ID, agentID)
            fmt.Printf("Artifacts: %s\n", strings.Join(ev.ResultingArtifactHashes, ", "))
            return nil
        },
    }

    cmd.Flags().StringVar(&agentID, "agent", "", "Agent identifier")
    cmd.Flags().StringVar(&scopeDir, "scope", "", "Directory scope to snapshot")
    cmd.Flags().StringVar(&cwd, "cwd", ".", "Working directory to execute command")
    return cmd
}


