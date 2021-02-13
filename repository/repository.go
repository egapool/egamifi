package repository

import (
	"gorm.io/gorm"
)

// Repository ...
type Repository struct {
	db *gorm.DB
}

// NewRepository ...
func NewRepository() *Repository {
	return &Repository{
		db:             lib.GetDBConn().Db,
	}
}

