package models

import (
	"ops-message-unified-push/pkg/util"

	"gorm.io/gorm"
)

type MessageTargetScope struct {
	SoftDeleteModel
	ID              uint      `json:"id" gorm:"autoIncrement;type:integer;primaryKey"`
	MessageID       string    `json:"message_id" gorm:"type:varchar(64);not null;index:idx_message_scope,unique"`
	MessageCategory string    `json:"message_category" gorm:"type:varchar(32);not null;index:idx_message_scope,unique"`
	TargetType      string    `json:"target_type" gorm:"type:varchar(32);not null;index:idx_message_scope,unique"`
	TargetID        string    `json:"target_id" gorm:"type:varchar(64);not null;index:idx_message_scope,unique"`
	CreatedAt       util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime"`
}

func ReplaceMessageTargetScopes(category string, messageID string, scopes []MessageTargetScope) error {
	// Scope rows are replace-set associations and must be replaced atomically.
	return WithTransaction(func(tx *gorm.DB) error {
		var existing []MessageTargetScope
		if err := tx.Unscoped().Where("message_category = ? AND message_id = ?", category, messageID).Find(&existing).Error; err != nil {
			return err
		}

		desired := make(map[string]MessageTargetScope, len(scopes))
		for _, scope := range scopes {
			scope.MessageCategory = category
			scope.MessageID = messageID
			desired[scope.TargetType+"\x00"+scope.TargetID] = scope
		}
		restored := make(map[string]struct{}, len(desired))

		for i := range existing {
			scope := &existing[i]
			key := scope.TargetType + "\x00" + scope.TargetID
			if _, ok := desired[key]; ok {
				if scope.DeletedAt.Valid {
					if err := tx.Unscoped().Model(scope).Update("deleted_at", nil).Error; err != nil {
						return err
					}
				}
				restored[key] = struct{}{}
				continue
			}
			if !scope.DeletedAt.Valid {
				if err := tx.Delete(scope).Error; err != nil {
					return err
				}
			}
		}

		for key, scope := range desired {
			if _, ok := restored[key]; ok {
				continue
			}
			if err := tx.Create(&scope).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
