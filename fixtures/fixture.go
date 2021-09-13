package fixtures

import "github.com/dhurimkelmendi/vending_machine/db"

// Fixtures is a struct that contains references to all fixture instances.
type Fixtures struct {
	User    *UserFixture
	Product *ProductFixture
}

var fixturesDefaultInstance *Fixtures

// GetFixturesDefaultInstance returns the default instance of Fixtures
func GetFixturesDefaultInstance() *Fixtures {
	if fixturesDefaultInstance == nil {
		// Purposeful pre-initialize
		_ = db.GetDefaultInstance()

		fixturesDefaultInstance = &Fixtures{
			User:    GetUserFixtureDefaultInstance(),
			Product: GetProductFixtureDefaultInstance(),
		}
	}
	return fixturesDefaultInstance
}
