package migrations

import (
	"github.com/go-pg/migrations/v8"
	"github.com/sirupsen/logrus"
)

func init() {
	migrations.Register(func(db migrations.DB) error {
		logrus.Infoln("The staging instances skip the first migration, so we do too")
		return nil
	}, func(db migrations.DB) error {
		return nil
	})
}
