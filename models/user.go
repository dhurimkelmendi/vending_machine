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
	tableName struct{}  `pg:"users"`
	ID        uuid.UUID `pg:"id,pk,type:uuid"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Role      UserRole  `json:"role"`
	Deposit   int32     `json:"deposit"`
}

// Equals compares two instances of type User
func (p *User) Equals(secondUser *User) bool {
	if p.ID != secondUser.ID {
		return false
	}
	if p.Username != secondUser.Username {
		return false
	}
	if p.Role != secondUser.Role {
		return false
	}
	if p.Deposit != secondUser.Deposit {
		return false
	}
	return true
}

// Render is used by go-chi/renderer
func (p *User) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
