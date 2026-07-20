package models

import (
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type lifecycleTestRecord struct {
	ID uint `gorm:"primaryKey"`
	SoftDeleteModel
	Name string
}

func TestSoftDeleteModelHidesDeletedRecords(t *testing.T) {
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := testDB.AutoMigrate(&lifecycleTestRecord{}); err != nil {
		t.Fatalf("migrate test table: %v", err)
	}
	record := lifecycleTestRecord{Name: "recoverable"}
	if err := testDB.Create(&record).Error; err != nil {
		t.Fatalf("create record: %v", err)
	}
	if err := testDB.Delete(&record).Error; err != nil {
		t.Fatalf("soft delete record: %v", err)
	}

	var visible int64
	if err := testDB.Model(&lifecycleTestRecord{}).Count(&visible).Error; err != nil {
		t.Fatalf("count visible records: %v", err)
	}
	if visible != 0 {
		t.Fatalf("visible records = %d, want 0", visible)
	}

	var retained int64
	if err := testDB.Unscoped().Model(&lifecycleTestRecord{}).Count(&retained).Error; err != nil {
		t.Fatalf("count retained records: %v", err)
	}
	if retained != 1 {
		t.Fatalf("retained records = %d, want 1", retained)
	}
}

func TestWithTransactionRollbackAndCommit(t *testing.T) {
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := testDB.AutoMigrate(&lifecycleTestRecord{}); err != nil {
		t.Fatalf("migrate test table: %v", err)
	}

	previousDB := db
	db = testDB
	t.Cleanup(func() { db = previousDB })

	expectedErr := errors.New("rollback requested")
	err = WithTransaction(func(tx *gorm.DB) error {
		if err := tx.Create(&lifecycleTestRecord{Name: "rolled-back"}).Error; err != nil {
			return err
		}
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("transaction error = %v, want %v", err, expectedErr)
	}

	var count int64
	if err := testDB.Model(&lifecycleTestRecord{}).Count(&count).Error; err != nil {
		t.Fatalf("count rolled-back records: %v", err)
	}
	if count != 0 {
		t.Fatalf("records after rollback = %d, want 0", count)
	}

	if err := WithTransaction(func(tx *gorm.DB) error {
		return tx.Create(&lifecycleTestRecord{Name: "committed"}).Error
	}); err != nil {
		t.Fatalf("commit transaction: %v", err)
	}
	if err := testDB.Model(&lifecycleTestRecord{}).Count(&count).Error; err != nil {
		t.Fatalf("count committed records: %v", err)
	}
	if count != 1 {
		t.Fatalf("records after commit = %d, want 1", count)
	}
}

func TestWithTransactionRollsBackOnPanic(t *testing.T) {
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := testDB.AutoMigrate(&lifecycleTestRecord{}); err != nil {
		t.Fatalf("migrate test table: %v", err)
	}

	previousDB := db
	db = testDB
	t.Cleanup(func() { db = previousDB })

	err = WithTransaction(func(tx *gorm.DB) error {
		if err := tx.Create(&lifecycleTestRecord{Name: "panicked"}).Error; err != nil {
			return err
		}
		panic("rollback requested")
	})
	if err == nil || err.Error() != "transaction callback failed: rollback requested" {
		t.Fatalf("transaction error = %v, want recovered callback error", err)
	}

	var count int64
	if err := testDB.Model(&lifecycleTestRecord{}).Count(&count).Error; err != nil {
		t.Fatalf("count records after panic: %v", err)
	}
	if count != 0 {
		t.Fatalf("records after panic = %d, want 0", count)
	}
}

func TestRolePermissionReplacementPhysicallyReplacesRows(t *testing.T) {
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := testDB.AutoMigrate(&RbacRolePermission{}); err != nil {
		t.Fatalf("migrate role permissions: %v", err)
	}

	previousDB := db
	db = testDB
	t.Cleanup(func() { db = previousDB })

	if err := SetRolePermissions(7, []uint{11, 12}, "first"); err != nil {
		t.Fatalf("create role permissions: %v", err)
	}
	if err := SetRolePermissions(7, []uint{12, 13}, "second"); err != nil {
		t.Fatalf("replace role permissions: %v", err)
	}
	if err := SetRolePermissions(7, []uint{11}, "third"); err != nil {
		t.Fatalf("replace role permission: %v", err)
	}

	ids, err := GetPermissionIDsByRoleID(7)
	if err != nil {
		t.Fatalf("read active role permissions: %v", err)
	}
	if len(ids) != 1 || ids[0] != 11 {
		t.Fatalf("active permission ids = %v, want [11]", ids)
	}

	var retained int64
	if err := testDB.Unscoped().Model(&RbacRolePermission{}).Count(&retained).Error; err != nil {
		t.Fatalf("count retained role permissions: %v", err)
	}
	if retained != 1 {
		t.Fatalf("retained role permissions = %d, want 1", retained)
	}
}

func TestArchiveRolePhysicallyDeletesRoleAndRelations(t *testing.T) {
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := testDB.AutoMigrate(&RbacRole{}, &RbacRolePermission{}, &RbacUserRole{}, &RbacGroupRole{}); err != nil {
		t.Fatalf("migrate RBAC tables: %v", err)
	}

	previousDB := db
	db = testDB
	t.Cleanup(func() { db = previousDB })

	role := RbacRole{Code: "physical-delete", Name: "Physical delete", Status: 1}
	if err := testDB.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	if err := testDB.Create(&RbacRolePermission{RoleID: role.ID, PermissionID: 11}).Error; err != nil {
		t.Fatalf("create role permission: %v", err)
	}
	if err := testDB.Create(&RbacUserRole{UserID: 7, RoleID: role.ID}).Error; err != nil {
		t.Fatalf("create user role: %v", err)
	}
	if err := testDB.Create(&RbacGroupRole{GroupID: 9, RoleID: role.ID}).Error; err != nil {
		t.Fatalf("create group role: %v", err)
	}

	if err := ArchiveRoleByID(role.ID); err != nil {
		t.Fatalf("delete role: %v", err)
	}

	for name, model := range map[string]interface{}{
		"role":            &RbacRole{},
		"role permission": &RbacRolePermission{},
		"user role":       &RbacUserRole{},
		"group role":      &RbacGroupRole{},
	} {
		var count int64
		if err := testDB.Unscoped().Model(model).Count(&count).Error; err != nil {
			t.Fatalf("count %s rows: %v", name, err)
		}
		if count != 0 {
			t.Fatalf("retained %s rows = %d, want 0", name, count)
		}
	}
}

func TestMessageTargetScopeReplacementRestoresArchivedRows(t *testing.T) {
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := testDB.AutoMigrate(&MessageTargetScope{}); err != nil {
		t.Fatalf("migrate message target scopes: %v", err)
	}

	previousDB := db
	db = testDB
	t.Cleanup(func() { db = previousDB })

	first := []MessageTargetScope{
		{TargetType: "role", TargetID: "operator"},
		{TargetType: "user", TargetID: "42"},
	}
	if err := ReplaceMessageTargetScopes("system", "message-1", first); err != nil {
		t.Fatalf("create message target scopes: %v", err)
	}
	if err := ReplaceMessageTargetScopes("system", "message-1", []MessageTargetScope{{TargetType: "role", TargetID: "operator"}}); err != nil {
		t.Fatalf("replace message target scopes: %v", err)
	}
	if err := ReplaceMessageTargetScopes("system", "message-1", []MessageTargetScope{{TargetType: "user", TargetID: "42"}}); err != nil {
		t.Fatalf("restore message target scope: %v", err)
	}

	var active []MessageTargetScope
	if err := testDB.Where("message_category = ? AND message_id = ?", "system", "message-1").Find(&active).Error; err != nil {
		t.Fatalf("read active message target scopes: %v", err)
	}
	if len(active) != 1 || active[0].TargetType != "user" || active[0].TargetID != "42" {
		t.Fatalf("active message target scopes = %#v, want user 42", active)
	}

	var retained int64
	if err := testDB.Unscoped().Model(&MessageTargetScope{}).Count(&retained).Error; err != nil {
		t.Fatalf("count retained message target scopes: %v", err)
	}
	if retained != 2 {
		t.Fatalf("retained message target scopes = %d, want 2", retained)
	}
}

func TestAddRoleRestoresArchivedBusinessKey(t *testing.T) {
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := testDB.AutoMigrate(&RbacRole{}); err != nil {
		t.Fatalf("migrate roles: %v", err)
	}

	previousDB := db
	db = testDB
	t.Cleanup(func() { db = previousDB })

	role := &RbacRole{Code: "operator", Name: "Operator", Status: 1}
	if err := AddRole(role); err != nil {
		t.Fatalf("create role: %v", err)
	}
	if err := testDB.Delete(role).Error; err != nil {
		t.Fatalf("archive role: %v", err)
	}

	restored := &RbacRole{Code: "operator", Name: "Operations", Status: 1}
	if err := AddRole(restored); err != nil {
		t.Fatalf("restore role: %v", err)
	}
	if restored.ID != role.ID {
		t.Fatalf("restored role id = %d, want %d", restored.ID, role.ID)
	}

	loaded, err := GetRoleByCode("operator")
	if err != nil {
		t.Fatalf("load restored role: %v", err)
	}
	if loaded.Name != "Operations" {
		t.Fatalf("restored role name = %q, want Operations", loaded.Name)
	}
}
