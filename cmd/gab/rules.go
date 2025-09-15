package main

import (
    "fmt"
    "gab/internal/deduce"
)

import "github.com/spf13/cobra"

func newRulesCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "rules",
        Short: "Install default Datalog/Mangle-like rules into the data directory",
        RunE: func(cmd *cobra.Command, args []string) error {
            path, err := deduce.InstallDefaultRules(dataDir)
            if err != nil { return err }
            fmt.Println(path)
            return nil
        },
    }
    return cmd
}


