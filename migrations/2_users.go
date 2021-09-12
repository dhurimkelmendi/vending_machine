package migrations

import (
	"github.com/go-pg/migrations/v8"
	"github.com/sirupsen/logrus"
)

func init() {
	migrations.Register(func(db migrations.DB) error {
		logrus.Infoln("Creating users table")
		_, err := db.Exec(`
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";		

		CREATE TABLE users (
			id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
			username varchar(100) UNIQUE NOT NULL,
			password varchar(100) NOT NULL,
			role varchar(100) NOT NULL,
			token text,
			deposit int
		);`)
		return err
	}, func(db migrations.DB) error {
		logrus.Infoln("Dropping users table")
		_, err := db.Exec(`
			DROP TABLE IF EXISTS users CASCADE;
			DROP EXTENSION "uuid-ossp";
		`)
		return err
	})
}
