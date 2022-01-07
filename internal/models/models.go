package models

import (
	"database/sql"
	"time"
)

// Models wrapper for all models
type Models struct {
	DB DBModel
}

// NewModel creates new instance model with database connection pool
func NewModel(db *sql.DB) Models {
	return Models{
		DB: DBModel{DB: db},
	}
}

// DBModel is the type for database connection values
type DBModel struct {
	DB *sql.DB
}

// Widget is the type for all widgets
type Widget struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	InventoryLevel int       `json:"inventory_level"`
	Price          int       `json:"price"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}
