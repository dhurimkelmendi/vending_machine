package controllers_test

import (
	"strings"

	"github.com/brianvoe/gofakeit"
	"github.com/dhurimkelmendi/vending_machine/auth"
	"github.com/dhurimkelmendi/vending_machine/models"
	uuid "github.com/satori/go.uuid"
)

func GetInvalidAuthToken() (string, error) {
	stateless := auth.GetStatelessAuthenticationProviderDefaultInstance()

	mockUser := &models.User{}
	mockUser.ID = uuid.NewV4()
	mockUser.Username = strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
	mockUser.Role = models.UserRoleBuyer
	mockUser.Token = gofakeit.BS()
	mockUser.Deposit = gofakeit.Int32()
	return stateless.CreateUserAuthToken(mockUser)
}
