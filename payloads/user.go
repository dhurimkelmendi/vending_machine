package payloads

import (
	"fmt"
	"net/http"

	"github.com/dhurimkelmendi/vending_machine/models"
	uuid "github.com/satori/go.uuid"
)

// UserList is a struct that contains a reference to a slice of type *models.UserDetails
type UserList struct {
	Users []*UserDetails `json:"users"`
}

// UserDetails simple response object
type UserDetails struct {
	ID       uuid.UUID       `json:"id"`
	Username string          `json:"username"`
	Role     models.UserRole `json:"role"`
	Deposit  int32           `json:"deposit"`
}

// Render is used by go-chi/renderer
func (u *UserDetails) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// MapUserToUserDetails convert a user model to a payload response
func MapUserToUserDetails(user *models.User) *UserDetails {
	return &UserDetails{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		Deposit:  user.Deposit,
	}
}

// Render is used by go-chi/renderer
func (ul *UserList) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// CreateUserPayload for registering a new user
type CreateUserPayload struct {
	Username string          `json:"username"`
	Password string          `json:"password"`
	Role     models.UserRole `json:"role"`
	Deposit  int32           `json:"deposit"`
}

// ToUserModel converts an instance of type *RegisterUserPayload to *models.User type
func (u *CreateUserPayload) ToUserModel() *models.User {
	return &models.User{
		Role:     u.Role,
		Username: u.Username,
		Password: u.Password,
		Deposit:  u.Deposit,
	}
}

// Validate ensures that all the required fields are present in an instance of *RegisterUserPayload
func (u *CreateUserPayload) Validate() error {
	if u.Username == "" {
		return fmt.Errorf("username is a required field")
	}
	if u.Password == "" {
		return fmt.Errorf("password is a required field")
	}
	if u.Username == u.Password {
		return fmt.Errorf("password can’t be the same as your username")
	}
	if u.Role == "" {
		return fmt.Errorf("role is a required field")
	}

	return nil
}

// Render is used by go-chi/renderer
func (u *CreateUserPayload) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// LoginUserPayload is a struct that represents the payload that is expected when logging a user in
type LoginUserPayload struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
}

// Validate ensures that all the required fields are present in an instance of *LoginUserPayload
func (u *LoginUserPayload) Validate() error {
	if u == nil {
		return fmt.Errorf("request body cannot be null")
	}
	return nil
}

// Render is used by go-chi/renderer
func (u *LoginUserPayload) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// UpdateUserPayload is a struct that represents the payload that is expected when updating a user
type UpdateUserPayload struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Deposit  int32     `json:"deposit"`
}

// ToUserModel converts an instance of type *UpdateUserPayload to *models.User type
func (u *UpdateUserPayload) ToUserModel() *models.User {
	return &models.User{
		ID:       u.ID,
		Username: u.Username,
		Deposit:  u.Deposit,
	}
}

// Validate ensures that all the required fields are present in an instance of *UpdateUserPayload
func (u *UpdateUserPayload) Validate() error {
	if u == nil {
		return fmt.Errorf("request body cannot be null")
	}
	return nil
}

// Render is used by go-chi/renderer
func (u *UpdateUserPayload) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
