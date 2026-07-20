package models

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"ops-message-unified-push/pkg/setting"
	"ops-message-unified-push/pkg/util"
)

var db *gorm.DB

type IDModel struct {
	ID uint `gorm:"autoIncrement;type:integer;primaryKey" json:"id"`

	CreatedBy  string    `json:"created_by" gorm:"type:varchar(100) ;default:'';"`
	ModifiedBy string    `json:"modified_by" gorm:"type:varchar(100) ;default:'';"`
	CreatedAt  util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime "`
	UpdatedAt  util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime ;"`
}

type UUIDModel struct {
	ID string `gorm:"type:varchar(12) ;primaryKey" json:"id"`

	CreatedBy  string    `json:"created_by" gorm:"type:varchar(100) ;default:'';"`
	ModifiedBy string    `json:"modified_by" gorm:"type:varchar(100) ;default:'';"`
	CreatedAt  util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime "`
	UpdatedAt  util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime ;"`
}

// SoftDeleteModel gives business records a recoverable lifecycle without
// applying soft deletion to RBAC join rows that are rebuilt transactionally.
type SoftDeleteModel struct {
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func notDeleted(table string) string {
	return table + ".deleted_at IS NULL"
}

// Setup initializes the database instance
func Setup() *gorm.DB {
	var err error

	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   setting.DatabaseSetting.TablePrefix,
			SingularTable: true,
		},
	}

	switch setting.DatabaseSetting.Type {
	case "mysql":
		connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FShanghai",
			setting.DatabaseSetting.User,
			setting.DatabaseSetting.Password,
			setting.DatabaseSetting.Host,
			setting.DatabaseSetting.Port,
			setting.DatabaseSetting.Name)
		if setting.DatabaseSetting.Ssl != "false" {
			connStr += fmt.Sprintf("&tls=%s", setting.DatabaseSetting.Ssl)
		}
		db, err = gorm.Open(mysql.Open(connStr), config)
	case "tidb":
		connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Asia%%2FShanghai",
			setting.DatabaseSetting.User,
			setting.DatabaseSetting.Password,
			setting.DatabaseSetting.Host,
			setting.DatabaseSetting.Port,
			setting.DatabaseSetting.Name)
		if setting.DatabaseSetting.Ssl != "false" {
			connStr += fmt.Sprintf("&tls=%s", setting.DatabaseSetting.Ssl)
		}
		db, err = gorm.Open(mysql.Open(connStr), config)
	case "sqlite":
		db, err = gorm.Open(sqlite.Open("conf/database.db"), config)
	}

	if err != nil {
		log.Fatalf("models.Setup err: %v", err)
	}

	if setting.DatabaseSetting.SqlDebug == "enable" {
		db = db.Debug()
	}
	return db
}

func GetSchema(table any) string {
	stmt := &gorm.Statement{DB: db}
	stmt.Parse(table)
	return stmt.Schema.Table
}

func GetDB() *gorm.DB {
	return db
}

// WithTransaction commits only when fn succeeds and explicitly rolls back on
// every error path so callers cannot accidentally leave a partial write set.
func WithTransaction(fn func(tx *gorm.DB) error) (err error) {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = tx.Rollback().Error
			err = fmt.Errorf("transaction callback failed: %v", recovered)
		}
	}()
	if err = fn(tx); err != nil {
		_ = tx.Rollback().Error
		return err
	}
	if err = tx.Commit().Error; err != nil {
		_ = tx.Rollback().Error
		return err
	}
	return nil
}
