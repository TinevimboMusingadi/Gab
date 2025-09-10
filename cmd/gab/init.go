package main

import (
    "fmt"
    "os"

    "gab/internal/storage"
)

import "github.com/spf13/cobra"

func newInitCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "init",
        Short: "Initialize Gab data directory",
        RunE: func(cmd *cobra.Command, args []string) error {
            if err := os.MkdirAll(dataDir, 0o755); err != nil {
                return err
            }
            if err := storage.Init(dataDir); err != nil {
                return err
            }
            fmt.Printf("Initialized Gab data at %s\n", dataDir)
            return nil
        },
    }
    return cmd
}


