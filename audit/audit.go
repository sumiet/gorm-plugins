package audit

import (
	"time"
)

// Model make Model Auditable, embed `audit.Model` into your model as anonymous field to make the model auditable
//    type User struct {
//      audit.Model
//    }
type Model struct {
	CreatedBy string     `gorm:"column:created_by;type:varchar(150)"`
	UpdatedBy string     `gorm:"column:updated_by;type:varchar(150)"`
	CreatedAt *time.Time `gorm:"column:created_at;type:timestamp;default:current_timestamp"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:timestamp;default:current_timestamp"`
}
