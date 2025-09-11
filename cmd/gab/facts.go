package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newFactsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "facts",
		Short: "Print path to the facts file (EDB) and show last lines",
		RunE: func(cmd *cobra.Command, args []string) error {
			facts := filepath.Join(dataDir, "deduce", "facts.dl")
			fmt.Println(facts)
			f, err := os.Open(facts)
			if err != nil {
				return nil
			}
			defer f.Close()
			buf := make([]byte, 4096)
			n, _ := f.Read(buf)
			if n > 0 {
				fmt.Println(string(buf[:n]))
			}
			return nil
		},
	}
	return cmd
}
