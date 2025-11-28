package models

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt int64
	UpdatedAt int64
	DeletedAt gorm.DeletedAt `gorm:"index"`
}