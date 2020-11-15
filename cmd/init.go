package cmd

import (
	"github.com/spf13/cobra"
	"github.com/svanellewee/historian/pkg/storage"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Init a database at the given location",
	RunE: func(cmd *cobra.Command, args []string) error {
		//fmt.Println("init...", args)
		_, err := storage.NewStore(args[0])
		if err != nil {
			return err
		}
		return nil
	},
}
