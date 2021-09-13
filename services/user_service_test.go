package services_test

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/dhurimkelmendi/vending_machine/config"
	"github.com/dhurimkelmendi/vending_machine/fixtures"
	"github.com/dhurimkelmendi/vending_machine/models"
	"github.com/dhurimkelmendi/vending_machine/payloads"
	"github.com/dhurimkelmendi/vending_machine/services"
	uuid "github.com/satori/go.uuid"
)

func TestUserService(t *testing.T) {
	t.Parallel()
	fixture := fixtures.GetFixturesDefaultInstance()

	service := services.GetUserServiceDefaultInstance()
	user := fixture.User.CreateSellerUser(t)

	ctx := context.Background()

	t.Run("create user", func(t *testing.T) {
		t.Run("create user with all fields", func(t *testing.T) {
			t.Run("create as buyer", func(t *testing.T) {
				userToCreate := &payloads.CreateUserPayload{}
				userToCreate.Username = strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
				userToCreate.Password = fmt.Sprintf("password_%d", rand.Intn(100000))
				userToCreate.Role = models.UserRoleBuyer
				userToCreate.Deposit = gofakeit.Int32()
				createedUser, err := service.CreateUser(ctx, userToCreate)
				if err != nil {
					t.Fatalf("error while creating user %+v", err)
				}
				userToCreateModel := userToCreate.ToUserModel()
				userToCreateModel.ID = createedUser.ID
				userToCreateModel.Token = createedUser.Token
				if !userToCreateModel.Equals(createedUser) {
					t.Fatalf("create user failed: %+v \n received: %+v, %+v", userToCreateModel, createedUser, err)
				}
			})
			t.Run("create as seller", func(t *testing.T) {
				userToCreate := &payloads.CreateUserPayload{}
				userToCreate.Username = strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
				userToCreate.Password = fmt.Sprintf("password_%d", rand.Intn(100000))
				userToCreate.Role = models.UserRoleSeller
				userToCreate.Deposit = gofakeit.Int32()
				createdUser, err := service.CreateUser(ctx, userToCreate)
				if err != nil {
					t.Fatalf("error while creating user %+v", err)
				}
				userToCreateModel := userToCreate.ToUserModel()
				userToCreateModel.ID = createdUser.ID
				userToCreateModel.Token = createdUser.Token
				if !userToCreateModel.Equals(createdUser) {
					t.Fatalf("create user failed: %+v \n received: %+v, %+v", userToCreateModel, createdUser, err)
				}
			})
		})
		t.Run("create user with existing username", func(t *testing.T) {
			userToCreate := &payloads.CreateUserPayload{}
			fakeDate := gofakeit.Date().Unix()
			if fakeDate < 0 {
				fakeDate = fakeDate * -1
			}
			userToCreate.Username = user.Username
			userToCreate.Password = fmt.Sprintf("password_%d", rand.Intn(100000))
			userToCreate.Role = models.UserRoleBuyer
			userToCreate.Deposit = gofakeit.Int32()
			_, err := service.CreateUser(ctx, userToCreate)
			if err == nil {
				t.Fatalf("expected duplicate user to fail %+v", err)
			}
		})
		t.Run("create user with no role", func(t *testing.T) {
			userToCreate := &payloads.CreateUserPayload{}
			userToCreate.Username = strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
			userToCreate.Password = fmt.Sprintf("password_%d", rand.Intn(100000))
			userToCreate.Deposit = gofakeit.Int32()
			_, err := service.CreateUser(ctx, userToCreate)
			if err == nil {
				t.Fatal("expected create to fail without Role, create was allowed")
			}
		})
	})

	t.Run("login user", func(t *testing.T) {
		loginUser := &payloads.LoginUserPayload{}
		loginUser.Username = user.Username
		loginUser.Password = "password"
		loggedInUser, err := service.LoginUser(ctx, loginUser)
		if err != nil || !loggedInUser.Equals(user) {
			t.Fatalf("login failed: %+v, %+v", loginUser, err)
		}
	})

	t.Run("login non-existent user", func(t *testing.T) {
		loginUser := &payloads.LoginUserPayload{}
		loginUser.Username = gofakeit.FirstName()
		loginUser.Password = user.Password
		loggedInUser, err := service.LoginUser(ctx, loginUser)
		if err == nil || loggedInUser.Equals(user) {
			t.Fatalf("expected login to fail: %+v, %+v", loginUser, err)
		}
	})
	t.Run("login user with wrong password", func(t *testing.T) {
		loginUser := &payloads.LoginUserPayload{}
		loginUser.Username = user.Username
		loginUser.Password = gofakeit.Password(true, false, false, false, false, 10)
		loggedInUser, err := service.LoginUser(ctx, loginUser)
		if err == nil || loggedInUser.Equals(user) {
			t.Fatalf("expected login to fail with wrong password: %+v, %+v", loginUser, err)
		}
	})

	t.Run("get user by id", func(t *testing.T) {
		_, err := service.GetUserByID(user.ID)
		if err != nil {
			t.Fatalf("could not retreive existing user by ID: %d, %+v", user.ID, err)
		}
	})

	t.Run("get user by username", func(t *testing.T) {
		_, err := service.GetUserByUsername(user.Username)
		if err != nil {
			t.Fatalf("could not retreive existing user by username: %s, %+v", user.Username, err)
		}
	})

	t.Run("get all users", func(t *testing.T) {
		_, err := service.GetAllUsers()
		if err != nil {
			t.Fatalf("could not retreive users: %+v", err)
		}
	})
	t.Run("deposit money", func(t *testing.T) {
		t.Run("deposit unacceptable amount", func(t *testing.T) {
			userToUpdate := &payloads.DepositMoneyPayload{}
			userToUpdate.ID = user.ID
			newDepositAmount := int32(123)
			userToUpdate.DepositAmount = newDepositAmount
			_, err := service.DepositMoney(ctx, userToUpdate)
			if err == nil {
				t.Fatalf("expected deposit money to fail with unacceptable amount, deposit allowed: %d", newDepositAmount)
			}
		})
		t.Run("deposit acceptable amount", func(t *testing.T) {
			acceptableAmountValues := config.GetDefaultInstance().AcceptableDepositAmountValues
			newDepositAmount := acceptableAmountValues[rand.Intn(len(acceptableAmountValues))]
			oldDepositAmount := user.Deposit
			userToUpdate := &payloads.DepositMoneyPayload{}
			userToUpdate.ID = user.ID
			userToUpdate.DepositAmount = newDepositAmount
			updatedUser, err := service.DepositMoney(ctx, userToUpdate)
			if err != nil {
				t.Fatalf("deposit money failed: %+v", err)
			}
			if updatedUser.Deposit != (oldDepositAmount + newDepositAmount) {
				t.Fatalf("expected new deposit to be: %d, got: %+v", newDepositAmount, updatedUser.Deposit)
			}
		})
	})
	t.Run("update user", func(t *testing.T) {
		t.Run("with basic attributes", func(t *testing.T) {
			userToUpdate := &payloads.UpdateUserPayload{}
			userToUpdate.ID = user.ID
			newDepositAmount := gofakeit.Int32()
			userToUpdate.Deposit = newDepositAmount
			updatedUser, err := service.UpdateUser(ctx, userToUpdate)
			if err != nil {
				t.Fatalf("update user failed: %+v", err)
			}
			if updatedUser.Deposit != newDepositAmount {
				t.Fatalf("expected deposit to be: %d, got: %+v", newDepositAmount, updatedUser.Deposit)
			}
		})
		t.Run("with protected attributes", func(t *testing.T) {
			userToUpdate := &payloads.UpdateUserPayload{}

			newID := uuid.NewV4()
			userToUpdate.ID = newID
			newDepositAmount := gofakeit.Int32()
			userToUpdate.Deposit = newDepositAmount
			updatedUser, _ := service.UpdateUser(ctx, userToUpdate)
			if updatedUser.ID == newID {
				t.Fatal("expected id not to be updated, update was allowed")
			}
		})
	})

	t.Run("delete user", func(t *testing.T) {
		t.Run("existing user", func(t *testing.T) {
			err := service.DeleteUser(ctx, user.ID)
			if err != nil {
				t.Fatalf("delete user failed: %+v", err)
			}
		})
	})

}
