package service

import "modul6/helper"

// CanAccessStudent memutuskan apakah seseorang boleh menyentuh data student tertentu.
// Fungsi murni: tidak mengimpor fiber maupun repository, sehingga bisa diuji
// hanya dengan memanggilnya langsung (unit test tanpa server/database).
func CanAccessStudent(current helper.AuthUser, ownerID int, perms *helper.PermissionSet, anyPermission string) bool {
	if current.UserID == ownerID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}
