package helpers

import (
	"fmt"

	"github.com/dhurimkelmendi/vending_machine/models"
)

// UserRolesContains checks if a slice of type models.UserRole contains a given UserRole
func UserRolesContains(r []models.UserRole, role models.UserRole) bool {
	for _, v := range r {
		if v == role {
			return true
		}
	}
	return false
}

// StringsContains checks if a slice of type string contains a given string
func StringsContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// Int32sCointains checks if a slice of type int32 contains a given int32
func Int32sCointains(s []int32, str int32) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// StringsEquals checks if two string slices are equal
func StringsEquals(a, b []string) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// InterfacesToStrings converts a slice of interface{} type to a slice of string type
func InterfacesToStrings(t []interface{}) []string {
	s := make([]string, len(t))
	for i, v := range t {
		s[i] = fmt.Sprint(v)
	}
	return s
}
