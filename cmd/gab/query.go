package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gab/internal/deduce"

	"github.com/spf13/cobra"
)

func newQueryCmd() *cobra.Command {
	var q string
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Placeholder Datalog query executor (to be backed by Mangle/Cozo)",
		RunE: func(cmd *cobra.Command, args []string) error {
			facts := filepath.Join(dataDir, "deduce", "facts.dl")
			if _, err := os.Stat(facts); err != nil {
				return fmt.Errorf("no facts found; run record-cmd first")
			}
			engine := &deduce.PlaceholderEngine{FactsPath: facts}
			res, err := engine.Exec(q)
			if err != nil {
				return err
			}
			fmt.Println(res)
			return nil
		},
	}
	cmd.Flags().StringVarP(&q, "expr", "e", "", "Datalog/Mangle expression")
	_ = cmd.MarkFlagRequired("expr")
	return cmd
}
