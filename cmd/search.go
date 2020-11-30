package cmd

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/svanellewee/historian/pkg/storage"
)

func init() {
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "search an entry into the database, using regex. Add more regexes to filter further",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		store, err := storage.NewStore(HistorianDatabase)
		if err != nil {
			return err
		}
		defer store.Close()

		filters := make([]storage.FilterFunction, 0, 1)
		for _, arg := range args {
			filters = append(filters, func(directoryName []byte, key []byte, value []byte) bool {
				re, err := regexp.Compile(arg)
				if err != nil {
					return false
				}
				return re.Find(value) != nil
			})
		}

		history, err := store.All(filters...)
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
