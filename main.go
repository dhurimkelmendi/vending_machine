package main

import (
	"os"

	"github.com/dhurimkelmendi/vending_machine/config"
	"github.com/dhurimkelmendi/vending_machine/db"
	"github.com/dhurimkelmendi/vending_machine/migrations"
	"github.com/dhurimkelmendi/vending_machine/server"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Infof("Server starting ...")

	if len(os.Args) > 1 {
		action := os.Args[1]

		if action == "migrate" {
			if len(os.Args) > 2 {
				migrate(os.Args[2])
			} else {
				logrus.Fatal("Missing migration action.")
			}
		} else {
			logrus.Fatalf("Unknown action: %s", action)
		}
	} else {
		run()
	}
}

func run() {
	config.GetDefaultInstance().SetLogLevel()
	config.GetDefaultInstance().LogConfigs()
	server.GetDefaultInstance().Start()
}

func migrate(action string) {
	logrus.Infof("Starting migration -- action: %s", action)
	config.GetDefaultInstance().SetLogLevel()
	config.GetDefaultInstance().LogConfigs()

	dbConn := db.GetDefaultInstance()

	if action == "reset" {
		migrations.Reset(dbConn.GetDB())
	} else {
		migrations.Migrate(action, dbConn.GetDB())
	}
}
