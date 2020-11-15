// Package cmd does cmd things [improve this TODO]
package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/svanellewee/historian/pkg/storage"
)

var (
	// HistorianConfigPath is the directory where all the historian data files are stored
	HistorianConfigPath string
	// HistorianDatabase is the actual location of the history file
	HistorianDatabase string
	rootCmd           = &cobra.Command{
		Use:   "historian",
		Short: "historian is a replacement for your bash history",
		Long:  `historian stores your history into a queryable database`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Historian")
			initHomeDir()
		},
	}
)

// Execute root command
func Execute() {
	initHomeDir()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initHomeDir() {
	home, err := homedir.Dir()
	if err != nil {
		logrus.Errorln(err)
	}

	HistorianConfigPath = path.Join(home, ".historian")
	if _, err := os.Stat(HistorianConfigPath); os.IsNotExist(err) {
		logrus.Infof("Creating directory at %s", HistorianConfigPath)
		os.Mkdir(HistorianConfigPath, 0777)
	}

	HistorianDatabase = path.Join(HistorianConfigPath, "history.db")
	if _, err = os.Stat(HistorianDatabase); err != nil {
		if os.IsNotExist(err) {
			logrus.Infof("Attempting database creation at %s", HistorianDatabase)
			store, err := storage.NewStore(HistorianDatabase)
			if err != nil {
				logrus.Errorf("could not create history database %v", err)
			}
			defer store.Close()
		} else {
			logrus.Errorf("Error in finding/creating database file %v", err)
		}
	}
}
