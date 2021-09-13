package fixtures

import (
	"context"
	"math/rand"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/dhurimkelmendi/vending_machine/db"
	"github.com/dhurimkelmendi/vending_machine/models"
	"github.com/dhurimkelmendi/vending_machine/payloads"
	"github.com/dhurimkelmendi/vending_machine/services"
	"github.com/go-pg/pg/v10"
	uuid "github.com/satori/go.uuid"
)

// UserFixture is a struct that contains references to the db and UserService
type UserFixture struct {
	db          *pg.DB
	userService *services.UserService
}

var userFixtureDefaultInstance *UserFixture

// GetUserFixtureDefaultInstance returns the default instance of UserFixture
func GetUserFixtureDefaultInstance() *UserFixture {
	if userFixtureDefaultInstance == nil {
		userFixtureDefaultInstance = &UserFixture{
			db:          db.GetDefaultInstance().GetDB(),
			userService: services.GetUserServiceDefaultInstance(),
		}
	}

	if userFixtureDefaultInstance.userService == nil {
		userFixtureDefaultInstance.userService = services.GetUserServiceDefaultInstance()
	}

	return userFixtureDefaultInstance
}

// CreateBuyerUser creates a user with fake data with buyer role
func (f *UserFixture) CreateBuyerUser(t *testing.T) *models.User {
	user := &payloads.CreateUserPayload{}
	user.Username = strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
	user.Password = gofakeit.Password(true, false, false, false, false, 10)
	user.Role = models.UserRoleBuyer

	user.Deposit = int32(rand.Intn(1000)+rand.Intn(1000)) * 5

	ctx := context.Background()

	if f.userService == nil {
		t.Log("CreateBuyerUser: fixture.UserService is nil!")
	}

	createdUser, err := f.userService.CreateUser(ctx, user)
	if err != nil {
		return nil
	}
	return createdUser
}

// CreateSellerUser creates a user with fake data with seller role
func (f *UserFixture) CreateSellerUser(t *testing.T) *models.User {
	user := &payloads.CreateUserPayload{}
	user.Username = strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
	user.Password = "password"
	user.Role = models.UserRoleSeller

	user.Deposit = int32(rand.Intn(1000)+rand.Intn(1000)) * 5

	ctx := context.Background()

	if f.userService == nil {
		t.Log("CreateSellerUser: fixture.UserService is nil!")
	}
	createdUser, err := f.userService.CreateUser(ctx, user)
	if err != nil {
		return nil
	}
	return createdUser
}
