package service

import (
	"strings"

	"modul6/helper"
)

// CanAccessUser: dua jalur — kepemilikan (paling murah, dicek duluan) atau permission.
func CanAccessUser(current helper.AuthUser, targetID int, perms *helper.PermissionSet, anyPermission string) bool {
	if current.UserID == targetID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}

func ValidateAssignRole(current helper.AuthUser, targetID int, role string, perms *helper.PermissionSet) map[string]string {
	errs := map[string]string{}
	role = strings.TrimSpace(role)
	if role == "" {
		errs["role"] = "wajib diisi"
		return errs
	}
	if !perms.IsKnownRole(role) {
		errs["role"] = "role tidak dikenal, pilih salah satu dari: " + strings.Join(perms.KnownRoles(), ", ")
	}
	if current.UserID == targetID {
		errs["role"] = "tidak boleh mengubah role diri sendiri"
	}
	return errs
}
