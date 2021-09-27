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
	seller := fixture.User.CreateSellerUser(t)
	buyer := fixture.User.CreateBuyerUser(t)
	product := fixture.Product.CreateProduct(t, seller.ID)
	acceptableDepositAmountValues := config.GetDefaultInstance().AcceptableDepositAmountValues

	ctx := context.Background()

	t.Run("create user", func(t *testing.T) {
		t.Run("create user with all fields", func(t *testing.T) {
			t.Run("create buyer", func(t *testing.T) {
				userToCreate := &payloads.CreateUserPayload{}
				userToCreate.Username = strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
				userToCreate.Password = fmt.Sprintf("password_%d", rand.Intn(100000))
				userToCreate.Role = models.UserRoleBuyer
				userToCreate.Deposit = acceptableDepositAmountValues[rand.Intn(len(acceptableDepositAmountValues))]
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
			t.Run("create seller", func(t *testing.T) {
				userToCreate := &payloads.CreateUserPayload{}
				userToCreate.Username = strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
				userToCreate.Password = fmt.Sprintf("password_%d", rand.Intn(100000))
				userToCreate.Role = models.UserRoleSeller
				userToCreate.Deposit = acceptableDepositAmountValues[rand.Intn(len(acceptableDepositAmountValues))]
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
		t.Run("with existing username", func(t *testing.T) {
			userToCreate := &payloads.CreateUserPayload{}
			fakeDate := gofakeit.Date().Unix()
			if fakeDate < 0 {
				fakeDate = fakeDate * -1
			}
			userToCreate.Username = seller.Username
			userToCreate.Password = fmt.Sprintf("password_%d", rand.Intn(100000))
			userToCreate.Role = models.UserRoleBuyer
			userToCreate.Deposit = acceptableDepositAmountValues[rand.Intn(len(acceptableDepositAmountValues))]
			_, err := service.CreateUser(ctx, userToCreate)
			if err == nil {
				t.Fatalf("expected duplicate user to fail %+v", err)
			}
		})
		t.Run("with no role", func(t *testing.T) {
			userToCreate := &payloads.CreateUserPayload{}
			userToCreate.Username = strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
			userToCreate.Password = fmt.Sprintf("password_%d", rand.Intn(100000))
			userToCreate.Deposit = acceptableDepositAmountValues[rand.Intn(len(acceptableDepositAmountValues))]
			_, err := service.CreateUser(ctx, userToCreate)
			if err == nil {
				t.Fatal("expected create to fail without Role, create was allowed")
			}
		})
	})
	t.Run("login", func(t *testing.T) {
		t.Run("valid login", func(t *testing.T) {
			loginUser := &payloads.LoginUserPayload{}
			loginUser.Username = seller.Username
			loginUser.Password = "password"
			loggedInUser, err := service.LoginUser(ctx, loginUser)
			if err != nil || !loggedInUser.Equals(seller) {
				t.Fatalf("login failed: %+v, %+v", loginUser, err)
			}
		})

		t.Run("login non-existent user", func(t *testing.T) {
			loginUser := &payloads.LoginUserPayload{}
			loginUser.Username = gofakeit.FirstName()
			loginUser.Password = seller.Password
			loggedInUser, err := service.LoginUser(ctx, loginUser)
			if err == nil || loggedInUser.Equals(seller) {
				t.Fatalf("expected login to fail: %+v, %+v", loginUser, err)
			}
		})
		t.Run("login user with wrong password", func(t *testing.T) {
			loginUser := &payloads.LoginUserPayload{}
			loginUser.Username = seller.Username
			loginUser.Password = gofakeit.Password(true, false, false, false, false, 10)
			loggedInUser, err := service.LoginUser(ctx, loginUser)
			if err == nil || loggedInUser.Equals(seller) {
				t.Fatalf("expected login to fail with wrong password: %+v, %+v", loginUser, err)
			}
		})
	})
	t.Run("get user by id", func(t *testing.T) {
		_, err := service.GetUserByID(seller.ID)
		if err != nil {
			t.Fatalf("could not retreive existing user by ID: %d, %+v", seller.ID, err)
		}
	})

	t.Run("get user by username", func(t *testing.T) {
		_, err := service.GetUserByUsername(seller.Username)
		if err != nil {
			t.Fatalf("could not retreive existing user by username: %s, %+v", seller.Username, err)
		}
	})

	t.Run("get all users", func(t *testing.T) {
		_, err := service.GetAllUsers()
		if err != nil {
			t.Fatalf("could not retreive users: %+v", err)
		}
	})
	t.Run("deposit money", func(t *testing.T) {
		t.Run("as seller", func(t *testing.T) {
			userToUpdate := &payloads.DepositMoneyPayload{}
			newDepositAmount := acceptableDepositAmountValues[rand.Intn(len(acceptableDepositAmountValues))]
			userToUpdate.DepositAmount = newDepositAmount
			_, err := service.DepositMoney(ctx, userToUpdate, seller.ID)
			if err == nil {
				t.Fatalf("expected deposit money to fail with seller user, deposit allowed: %s", seller.ID.String())
			}
		})
		t.Run("as buyer", func(t *testing.T) {
			t.Run("deposit unacceptable amount", func(t *testing.T) {
				userToUpdate := &payloads.DepositMoneyPayload{}
				newDepositAmount := int32(123)
				userToUpdate.DepositAmount = newDepositAmount
				_, err := service.DepositMoney(ctx, userToUpdate, buyer.ID)
				if err == nil {
					t.Fatalf("expected deposit money to fail with unacceptable amount, deposit allowed: %d", newDepositAmount)
				}
			})
			t.Run("deposit acceptable amount", func(t *testing.T) {
				newDepositAmount := acceptableDepositAmountValues[rand.Intn(len(acceptableDepositAmountValues))]
				oldDepositAmount := buyer.Deposit
				userToUpdate := &payloads.DepositMoneyPayload{}
				userToUpdate.DepositAmount = newDepositAmount
				updatedUser, err := service.DepositMoney(ctx, userToUpdate, buyer.ID)
				if err != nil {
					t.Fatalf("deposit money failed: %+v", err)
				}
				if updatedUser.Deposit != (oldDepositAmount + newDepositAmount) {
					t.Fatalf("expected new deposit to be: %d, got: %+v", newDepositAmount, updatedUser.Deposit)
				}
			})
		})
	})
	t.Run("reset deposit", func(t *testing.T) {
		updatedUser, err := service.ResetDeposit(ctx, seller.ID)
		if err != nil {
			t.Fatalf("reset deposit failed: %+v", err)
		}
		if updatedUser.Deposit != 0 {
			t.Fatalf("expected new deposit to be: %d, got: %+v", 0, updatedUser.Deposit)
		}
	})
	t.Run("buy product", func(t *testing.T) {
		t.Run("with sufficient deposit", func(t *testing.T) {
			productPurchase := &payloads.UserProductPurchase{
				ProductID: product.ID,
				Amount:    2,
			}
			_, err := service.BuyProduct(ctx, productPurchase, buyer.ID)
			if err != nil {
				t.Fatalf("product purchase failed: %+v", err)
			}
		})
		t.Run("with insufficient deposit", func(t *testing.T) {
			productPurchase := &payloads.UserProductPurchase{
				ProductID: product.ID,
				Amount:    int32(rand.Intn(999999999)),
			}
			_, err := service.BuyProduct(ctx, productPurchase, buyer.ID)
			if err == nil {
				t.Fatal("expected product purchase to fail due to insufficient deposit, purchase was allowed")
			}
		})
	})
	t.Run("update user", func(t *testing.T) {
		t.Run("with basic attributes", func(t *testing.T) {
			userToUpdate := &payloads.UpdateUserPayload{}
			userToUpdate.ID = seller.ID
			newUsername := strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
			userToUpdate.Username = newUsername
			updatedUser, err := service.UpdateUser(ctx, userToUpdate)
			if err != nil {
				t.Fatalf("update user failed: %+v", err)
			}
			if updatedUser.Username != newUsername {
				t.Fatalf("expected username to be: %s, got: %+v", newUsername, updatedUser.Username)
			}
		})
		t.Run("with protected attributes", func(t *testing.T) {
			userToUpdate := &payloads.UpdateUserPayload{}

			newID := uuid.NewV4()
			userToUpdate.ID = newID
			newUsername := strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
			userToUpdate.Username = newUsername
			updatedUser, _ := service.UpdateUser(ctx, userToUpdate)
			if updatedUser.ID == newID {
				t.Fatal("expected id not to be updated, update was allowed")
			}
		})
	})

	t.Run("delete user", func(t *testing.T) {
		t.Run("existing user", func(t *testing.T) {
			err := service.DeleteUser(ctx, seller.ID)
			if err != nil {
				t.Fatalf("delete user failed: %+v", err)
			}
		})
	})

}
