// Package migrations is responsible for running Postgres database migrations.
package migrations

import (
	"github.com/dhurimkelmendi/vending_machine/internal/trace"
	"github.com/go-pg/migrations/v8"
	"github.com/sirupsen/logrus"
)

// Migrate executes the intended migration action on the provided DB. It will
// also report any changes in the migration version, or if there are no
// changes.
func Migrate(action string, db migrations.DB) {
	if _, _, err := migrations.Run(db, "init"); err != nil {
		logrus.Info("Initial migration has already been run")
	}

	oldVersion, newVersion, err := migrations.Run(db, action)
	if err != nil {
		logrus.Fatalf("%s: Failed to migrate database: %+v", trace.Getfl(), err)
	}

	if newVersion != oldVersion {
		logrus.Infof("Migrated database from version %d to %d", oldVersion, newVersion)
	} else {
		logrus.Infof("No migrations to execute, current version is %d", oldVersion)
	}
}

// Reset attempts to undo all of the database migrations. It is only called in
// the database test teardown method, and is intended as a utility method.
func Reset(db migrations.DB) {
	// We can't really use `Migrate("reset", db)` because it's too nice in that it tries to roll
	// back the migrations back to the beginning and then reapply all of them. The problem with
	// that is if some tests fails and leaves data in the table, those left over data would cause
	// some migrations' to fail to be rolled back.
	// Since the ultimate goal here is just to get the database back to a consistent initial state,
	// we can simply drop the schema and recreate it and that would let us start over with a blank
	// database.
	_, err := db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err != nil {
		logrus.Fatalf("Failed to reset database: %v", err)
	}
}
