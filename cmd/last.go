package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/svanellewee/historian/pkg/storage"
)

func init() {
	rootCmd.AddCommand(lastCmd)
}

var lastCmd = &cobra.Command{
	Use:   "last",
	Short: "last entry into the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		numCount := 1
		if len(args) == 1 {
			numCount, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}
		}
		store, err := storage.NewStore(HistorianDatabase)
		if err != nil {
			return err
		}
		defer store.Close()

		currentDirectory, err := os.Getwd()
		if err != nil {
			return err
		}
		history, err := store.Last(currentDirectory, numCount)
		if err != nil {
			return err
		}
		for _, elem := range history {
			fmt.Printf("%s\n", elem)
		}
		return nil
	},
}

/*
function insert-hist () {
  $HOME/source/historian/historian insert "$(history 1)"
}
*/
