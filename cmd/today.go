package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/svanellewee/historian/pkg/storage"
	bolt "go.etcd.io/bbolt"
)

func init() {
	rootCmd.AddCommand(todayCmd)
}

type result struct {
	directory []byte
	timestamp []byte
	command   []byte
}

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "today entry into the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		results := make([]result, 0, 10)
		var err error
		store, err := storage.NewStore(HistorianDatabase)
		if err != nil {
			return err
		}
		defer store.Close()
		today := time.Now()
		err = store.AllBucketsForDay(today, func(directory []byte, b *bolt.Bucket, timestamp []byte, command []byte) error {
			//fmt.Printf("[%s] key=%s, value=%s\n", string(directory), timestamp, command)
			results = append(results, result{
				directory: directory,
				timestamp: timestamp,
				command:   command,
			})
			return nil
		})
		if err != nil {
			return err
		}
		// Sort the list
		sort.Slice(results, func(i, j int) bool {
			t1, err := time.Parse(time.RFC3339, string(results[i].timestamp))
			if err != nil {
				logrus.Fatalf("could not parse time %s", results[i].timestamp)
			}
			t2, err := time.Parse(time.RFC3339, string(results[j].timestamp))
			if err != nil {
				logrus.Fatalf("could not parse time %s", results[j].timestamp)
			}
			return t1.Before(t2)
		})
		for _, element := range results {
			timeResult, err := time.Parse(time.RFC3339, string(element.timestamp))
			if err != nil {
				logrus.Fatalf("could not bind time %v", err)
			}
			fmt.Printf("[%s] %s %s\n", timeResult, element.directory, element.command)
		}
		return nil
	},
}

/*
function insert-hist () {
  $HOME/source/historian/historian insert "$(history 1)"
}
*/
