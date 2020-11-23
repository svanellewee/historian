package cmd

import (
	"github.com/spf13/cobra"
	"github.com/svanellewee/historian/pkg/storage"
)

type entry struct {
	id   int64
	data string
}

// ErrIncorrectCount describes incorrect number of arguments

func init() {
	rootCmd.AddCommand(insertCmd)
}

var insertCmd = &cobra.Command{
	Use:   "insert",
	Short: "insert entry into the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		entry, err := storage.Convert(args[0])
		if err != nil {
			return err
		}

		store, err := storage.NewStore(HistorianDatabase)
		if err != nil {
			return err
		}
		defer store.Close()

		return store.Add(entry)
	},
}

/*
function insert-hist () {
  $HOME/source/historian/historian insert "$(history 1)"
}
*/
