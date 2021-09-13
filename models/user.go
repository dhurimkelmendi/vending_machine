package models

import (
	"net/http"

	uuid "github.com/satori/go.uuid"
)

// UserRole represents the type of User.Role
type UserRole string

// Available user roles
const (
	UserRoleSeller UserRole = "seller"
	UserRoleBuyer  UserRole = "buyer"
)

// User is a struct that represents a db row of the Users table
type User struct {
	tableName struct{}   `pg:"users"`
	ID        uuid.UUID  `pg:"id,pk,type:uuid"`
	Username  string     `json:"username"`
	Password  string     `json:"password"`
	Token     string     `json:"token"`
	Role      UserRole   `json:"role"`
	Deposit   int32      `json:"deposit"`
	Products  []*Product `pg:"many2many:users_products"`
}

// Merge merges two instances of type User into one
func (u *User) Merge(secondUser User) {
	if u.Username == "" {
		u.Username = secondUser.Username
	}
	if u.Password == "" {
		u.Password = secondUser.Password
	}
	if u.Role == "" {
		u.Role = secondUser.Role
	}
	if u.Deposit == 0 {
		u.Deposit = secondUser.Deposit
	}
}

// Equals compares two instances of type User
func (u *User) Equals(secondUser *User) bool {
	if u.ID != secondUser.ID {
		return false
	}
	if u.Username != secondUser.Username {
		return false
	}
	if u.Role != secondUser.Role {
		return false
	}
	if u.Deposit != secondUser.Deposit {
		return false
	}
	return true
}

// Render is used by go-chi/renderer
func (u *User) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
