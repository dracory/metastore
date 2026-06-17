package metastore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/dracory/neat"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dracory/uid"
)

// Store defines a meta store
type Store struct {
	metaTableName      string
	db                 *neat.Database
	debugEnabled       bool
	automigrateEnabled bool
	logger             *slog.Logger
}

type NewStoreOptions struct {
	MetaTableName      string
	DB                 *sql.DB
	AutomigrateEnabled bool
	DebugEnabled       bool
}

// NewStore creates a new meta store
func NewStore(opts NewStoreOptions) (StoreInterface, error) {
	if opts.MetaTableName == "" {
		return nil, errors.New("meta store: metaTableName is required")
	}

	if opts.DB == nil {
		return nil, errors.New("meta store: DB is required")
	}

	neatDB, err := neat.NewFromSQLDB(opts.DB)
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	store := &Store{
		metaTableName:      opts.MetaTableName,
		db:                 neatDB,
		automigrateEnabled: opts.AutomigrateEnabled,
		debugEnabled:       opts.DebugEnabled,
		logger:             logger,
	}

	if store.automigrateEnabled {
		if err := store.MigrateUp(context.Background()); err != nil {
			return nil, err
		}
	}

	return store, nil
}

// MigrateUp creates the meta table
func (st *Store) MigrateUp(ctx context.Context, tx ...*sql.Tx) error {
	if st.db.Schema().HasTable(st.metaTableName) {
		if st.debugEnabled {
			st.logger.Info("MigrateUp: table already exists", "table", st.metaTableName)
		}
		return nil
	}

	err := st.db.Schema().Create(st.metaTableName, func(table contractsschema.Blueprint) {
		table.String(COLUMN_ID, 40)
		table.Primary(COLUMN_ID)
		table.String(COLUMN_OBJECT_TYPE, 100)
		table.String(COLUMN_OBJECT_ID, 40)
		table.String(COLUMN_META_KEY, 255)
		table.Text(COLUMN_META_VALUE)
		table.DateTime(COLUMN_CREATED_AT)
		table.DateTime(COLUMN_UPDATED_AT)
		table.DateTime(COLUMN_DELETED_AT).Nullable()
	})

	if err != nil {
		if st.debugEnabled {
			st.logger.Error("MigrateUp failed", "error", err)
		}
		return err
	}

	return nil
}

// MigrateDown drops the meta table
func (st *Store) MigrateDown(ctx context.Context, tx ...*sql.Tx) error {
	if !st.db.Schema().HasTable(st.metaTableName) {
		if st.debugEnabled {
			st.logger.Info("MigrateDown: table does not exist", "table", st.metaTableName)
		}
		return nil
	}

	err := st.db.Schema().Drop(st.metaTableName)
	if err != nil {
		if st.debugEnabled {
			st.logger.Error("MigrateDown failed", "error", err)
		}
		return err
	}

	return nil
}

// EnableDebug enables or disables debug mode
func (st *Store) EnableDebug(debug bool) {
	st.debugEnabled = debug
	if debug {
		st.db.EnableDebug()
		st.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		st.db.DisableDebug()
		st.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
}

// FindByKey finds a meta entry by key
func (st *Store) FindByKey(objectType string, objectID string, key string) (*Meta, error) {
	var rows []Meta
	err := st.db.Query().Table(st.metaTableName).
		Where(COLUMN_OBJECT_TYPE+" = ?", objectType).
		Where(COLUMN_OBJECT_ID+" = ?", objectID).
		Where(COLUMN_META_KEY+" = ?", key).
		Where(COLUMN_DELETED_AT + " IS NULL").
		Limit(1).
		Get(&rows)

	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, nil
	}

	return &rows[0], nil
}

// Get gets a key from cache
func (st *Store) Get(objectType string, objectID string, key string, valueDefault string) (string, error) {
	meta, err := st.FindByKey(objectType, objectID, key)

	if err != nil {
		return "", err
	}

	if meta != nil {
		return meta.Value, nil
	}

	return valueDefault, nil
}

// GetJSON gets a JSON key from cache
func (st *Store) GetJSON(objectType string, objectID string, key string, valueDefault interface{}) (interface{}, error) {
	meta, err := st.FindByKey(objectType, objectID, key)

	if err != nil {
		return nil, err
	}

	if meta != nil {
		jsonValue := meta.Value
		var intrfc interface{}
		jsonError := json.Unmarshal([]byte(jsonValue), &intrfc)
		if jsonError != nil {
			return valueDefault, jsonError
		}

		return intrfc, nil
	}

	return valueDefault, nil
}

// Remove deletes a meta key
func (st *Store) Remove(objectType string, objectID string, key string) error {
	_, err := st.db.Query().Table(st.metaTableName).
		Where(COLUMN_OBJECT_TYPE+" = ?", objectType).
		Where(COLUMN_OBJECT_ID+" = ?", objectID).
		Where(COLUMN_META_KEY+" = ?", key).
		Delete()

	if err != nil {
		return err
	}

	return nil
}

// Set sets new key value pair
func (st *Store) Set(objectType string, objectID string, key string, value string) error {
	meta, err := st.FindByKey(objectType, objectID, key)

	if err != nil {
		return err
	}

	if meta == nil {
		row := map[string]any{
			COLUMN_ID:          uid.HumanUid(),
			COLUMN_OBJECT_TYPE: objectType,
			COLUMN_OBJECT_ID:   objectID,
			COLUMN_META_KEY:    key,
			COLUMN_META_VALUE:  value,
			COLUMN_CREATED_AT:  time.Now(),
			COLUMN_UPDATED_AT:  time.Now(),
		}
		return st.db.Query().Table(st.metaTableName).Create(row)
	}

	row := map[string]any{
		COLUMN_META_VALUE: value,
		COLUMN_UPDATED_AT: time.Now(),
	}

	_, err = st.db.Query().Table(st.metaTableName).
		Where(COLUMN_OBJECT_TYPE+" = ?", objectType).
		Where(COLUMN_OBJECT_ID+" = ?", objectID).
		Where(COLUMN_META_KEY+" = ?", key).
		Update(row)

	return err
}

// SetJSON sets new key value pair
func (st *Store) SetJSON(objectType string, objectID string, key string, value interface{}) error {
	jsonValue, jsonError := json.Marshal(value)

	if jsonError != nil {
		return jsonError
	}

	return st.Set(objectType, objectID, key, string(jsonValue))
}

// GetMetaTableName returns the meta table name
func (st *Store) GetMetaTableName() string {
	return st.metaTableName
}

// GetDB returns the database connection
func (st *Store) GetDB() *sql.DB {
	db, _ := st.db.DB()
	return db
}

// IsAutomigrateEnabled returns whether automigrate is enabled
func (st *Store) IsAutomigrateEnabled() bool {
	return st.automigrateEnabled
}
