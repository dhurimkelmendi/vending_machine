package migrations

import (
	"github.com/go-pg/migrations/v8"
	"github.com/sirupsen/logrus"
)

func init() {
	migrations.Register(func(db migrations.DB) error {
		logrus.Infoln("Creating products table")
		_, err := db.Exec(`
		CREATE TABLE products (
			id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
			name text UNIQUE NOT NULL,
			seller_id uuid REFERENCES users(id) NOT NULL,
			amount_available int NOT NULL,
			cost int NOT NULL
		);`)
		return err
	}, func(db migrations.DB) error {
		logrus.Infoln("Dropping products table")
		_, err := db.Exec(`
			DROP TABLE IF EXISTS products CASCADE;
		`)
		return err
	})
}
