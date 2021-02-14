package repository

import (
	"github.com/egapool/egamifi/database"
	"gorm.io/gorm"
)

// Repository ...
type Repository struct {
	db *gorm.DB
}

// NewRepository ...
func NewRepository() *Repository {
	return &Repository{
		db: database.GetDBConn().Db,
	}
}
