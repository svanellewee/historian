package main

import (
	//"github.com/sirupsen/logrus"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/svanellewee/historian/cmd"
	"github.com/svanellewee/historian/pkg/store"
)

var (
	// HistorianConfigPath is the directory where all the historian data files are stored
	HistorianConfigPath string
	// HistorianDatabase is the actual location of the history file
	HistorianDatabase string
)

func init() {
	home, err := homedir.Dir()
	if err != nil {
		logrus.Errorln(err)
	}
	HistorianConfigPath = path.Join(home, ".historian")
	if _, err := os.Stat(HistorianConfigPath); os.IsNotExist(err) {
		os.Mkdir(HistorianConfigPath, 0600)
	}
	HistorianDatabase = path.Join(HistorianConfigPath, "history.db")

	_, err = store.NewStore(HistorianDatabase)
	if err != nil {
		logrus.Errorf("could not create history database %v", err)
	}
}

func main() {
	cmd.Execute()
}
