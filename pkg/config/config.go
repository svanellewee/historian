package config

// var (
// 	// HistorianConfigPath is the directory where all the historian data files are stored
// 	HistorianConfigPath string
// 	// HistorianDatabase is the actual location of the history file
// 	HistorianDatabase string
// )

// func init() {
// 	home, err := homedir.Dir()
// 	if err != nil {
// 		logrus.Errorln(err)
// 	}
// 	HistorianConfigPath = path.Join(home, ".historian")
// 	if _, err := os.Stat(HistorianConfigPath); os.IsNotExist(err) {
// 		os.Mkdir(HistorianConfigPath, 0600)
// 	}
// 	HistorianDatabase = path.Join(HistorianConfigPath, "history.db")

// }
