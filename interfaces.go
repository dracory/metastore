package metastore

import (
	"context"
	"database/sql"
)

// StoreInterface defines the interface for meta store operations
type StoreInterface interface {
	// MigrateDown drops the meta table
	MigrateDown(ctx context.Context, tx ...*sql.Tx) error

	// MigrateUp creates the meta table
	MigrateUp(ctx context.Context, tx ...*sql.Tx) error

	// EnableDebug enables or disables debug logging
	EnableDebug(debug bool)

	// FindByKey finds a meta entry by object type, object ID, and key
	FindByKey(objectType string, objectID string, key string) (*Meta, error)

	// Get retrieves a value by object type, object ID, and key
	Get(objectType string, objectID string, key string, valueDefault string) (string, error)

	// GetJSON retrieves and unmarshals a JSON value by object type, object ID, and key
	GetJSON(objectType string, objectID string, key string, valueDefault interface{}) (interface{}, error)

	// Remove deletes a meta entry by object type, object ID, and key
	Remove(objectType string, objectID string, key string) error

	// Set stores a string value for a given object type, object ID, and key
	Set(objectType string, objectID string, key string, value string) error

	// SetJSON marshals and stores a JSON value for a given object type, object ID, and key
	SetJSON(objectType string, objectID string, key string, value interface{}) error

	// SqlCreateTable returns the SQL statement for creating the meta table
	SqlCreateTable() string

	// GetMetaTableName returns the meta table name
	GetMetaTableName() string

	// GetDB returns the database connection
	GetDB() interface{}

	// IsAutomigrateEnabled returns whether automigrate is enabled
	IsAutomigrateEnabled() bool
}
