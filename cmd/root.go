// Package cmd does cmd things [improve this TODO]
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "historian",
		Short: "historian is a replacement for your bash history",
		Long:  `historian stores your history into a queryable database`,
		Run: func(cmd *cobra.Command, args []string) {
			//fmt.Println("Call your root here...")

		},
	}
)

// Execute root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
