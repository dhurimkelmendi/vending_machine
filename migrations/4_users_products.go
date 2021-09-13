package migrations

import (
	"github.com/go-pg/migrations/v8"
	"github.com/sirupsen/logrus"
)

func init() {
	migrations.Register(func(db migrations.DB) error {
		logrus.Infoln("Creating users_products table")
		_, err := db.Exec(`
		CREATE TABLE users_products (
			id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id uuid REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
			product_id uuid REFERENCES products(id) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
			amount int NOT NULL
		);`)
		return err
	}, func(db migrations.DB) error {
		logrus.Infoln("Dropping users_products table")
		_, err := db.Exec(`
			DROP TABLE IF EXISTS users_products CASCADE;
		`)
		return err
	})
}
