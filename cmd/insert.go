package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/svanellewee/historian/pkg/storage"
)

type entry struct {
	id   int64
	data string
}

// ErrIncorrectCount describes incorrect number of arguments
type ErrIncorrectCount struct {
	ArgumentCount int
	ExpectedCount int
}

func (e ErrIncorrectCount) Error() string {
	return fmt.Sprintf("incorrect number of elements, Expected [%d], found [%d]", e.ArgumentCount, e.ExpectedCount)
}

func convert(input string) (*entry, error) {
	trimmed := strings.Trim(input, " ")
	elements := strings.SplitN(trimmed, " ", 2)
	if len(elements) != 2 {
		return nil, ErrIncorrectCount{ArgumentCount: len(elements), ExpectedCount: 2}
	}
	number, err := strconv.ParseInt(elements[0], 10, 64)
	if err != nil {
		return nil, err
	}
	return &entry{
		id:   number,
		data: strings.Trim(elements[1], " "),
	}, nil
}

func init() {
	rootCmd.AddCommand(insertCmd)
}

var insertCmd = &cobra.Command{
	Use:   "insert",
	Short: "insert entry into the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		entry, err := convert(args[0])
		if err != nil {
			return err
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
		return store.Add(currentDirectory, time.Now(), []byte(entry.data))
	},
}

/*
function insert-hist () {
  $HOME/source/historian/historian insert "$(history 1)"
}
*/
