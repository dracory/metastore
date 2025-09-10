package metastore

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlserver"
	"github.com/dracory/sb"
	"github.com/dracory/uid"
	"github.com/georgysavva/scany/sqlscan"
)

// Store defines a session store
type Store struct {
	metaTableName      string
	db                 *sql.DB
	dbDriverName       string
	debugEnabled       bool
	automigrateEnabled bool
}

type NewStoreOptions struct {
	MetaTableName      string
	DB                 *sql.DB
	DbDriverName       string
	AutomigrateEnabled bool
	DebugEnabled       bool
}

// StoreOption options for the cache store
type StoreOption func(*Store)

// NewStore creates a new entity store
func NewStore(opts NewStoreOptions) (StoreInterface, error) {
	store := &Store{
		metaTableName:      opts.MetaTableName,
		automigrateEnabled: opts.AutomigrateEnabled,
		db:                 opts.DB,
		dbDriverName:       opts.DbDriverName,
		debugEnabled:       opts.DebugEnabled,
	}

	if store.metaTableName == "" {
		return nil, errors.New("meta store: metaTableName is required")
	}

	if store.db == nil {
		return nil, errors.New("meta store: DB is required")
	}

	if store.dbDriverName == "" {
		store.dbDriverName = sb.DatabaseDriverName(store.db)
	}

	if store.automigrateEnabled {
		store.AutoMigrate()
	}

	return store, nil
}

// AutoMigrate auto migrate
func (st *Store) AutoMigrate() error {
	sql := st.SqlCreateTable()

	if sql == "" {
		return errors.New("meta table create sql is empty")
	}

	if st.debugEnabled {
		log.Println(sql)
	}

	_, err := st.db.Exec(sql)

	if err != nil {
		return err
	}

	return nil
}

// EnableDebug - enables the debug option
func (st *Store) EnableDebug(debug bool) {
	st.debugEnabled = debug
}

// FindByKey finds a cache by key
func (st *Store) FindByKey(objectType string, objectID string, key string) (*Meta, error) {
	sqlStr, params, err := goqu.Dialect(st.dbDriverName).
		From(st.metaTableName).
		Prepared(true).
		Where(goqu.C(COLUMN_OBJECT_TYPE).Eq(objectType), goqu.C(COLUMN_OBJECT_ID).Eq(objectID), goqu.C(COLUMN_META_KEY).Eq(key)).
		Limit(1).
		ToSQL()

	if err != nil {
		return nil, err
	}

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	var meta Meta
	err = sqlscan.Get(context.Background(), st.db, &meta, sqlStr, params...)

	if err != nil {
		if err == sql.ErrNoRows {
			// Looks like this is now outdated for sqlscan
			return nil, nil // not really an error, no such row
		}

		if sqlscan.NotFound(err) {
			return nil, nil // not really an error, no such row
		}

		return nil, err
	}

	return &meta, nil
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
	sqlStr, params, err := goqu.Dialect(st.dbDriverName).
		From(st.metaTableName).
		Prepared(true).
		Where(goqu.C(COLUMN_OBJECT_TYPE).Eq(objectType), goqu.C(COLUMN_OBJECT_ID).Eq(objectID), goqu.C(COLUMN_META_KEY).Eq(key)).
		Delete().
		ToSQL()

	if err != nil {
		return err
	}

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	_, err = st.db.Exec(sqlStr, params...)
	if err != nil {
		if err == sql.ErrNoRows {
			// Looks like this is now outdated for sqlscan
			return nil // not really an error, already not there
		}

		if sqlscan.NotFound(err) {
			return nil
		}

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

	var sqlStr string
	var params []interface{}
	var sqlErr error
	if meta == nil {
		var newMeta = &Meta{
			ID:         uid.HumanUid(),
			ObjectType: objectType,
			ObjectID:   objectID,
			Key:        key,
			Value:      value,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		sqlStr, params, sqlErr = goqu.Dialect(st.dbDriverName).
			Insert(st.metaTableName).
			Prepared(true).
			Rows(newMeta).
			ToSQL()
	} else {
		fields := map[string]interface{}{}
		fields["meta_value"] = value
		fields["updated_at"] = time.Now()
		sqlStr, params, sqlErr = goqu.Dialect(st.dbDriverName).
			Update(st.metaTableName).
			Prepared(true).
			Where(goqu.C(COLUMN_OBJECT_TYPE).Eq(objectType), goqu.C(COLUMN_OBJECT_ID).Eq(objectID), goqu.C(COLUMN_META_KEY).Eq(key)).
			Set(fields).
			ToSQL()
	}

	if sqlErr != nil {
		return sqlErr
	}

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	_, err = st.db.Exec(sqlStr, params...)

	if err != nil {
		return err
	}

	return nil
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
func (st *Store) GetDB() interface{} {
	return st.db
}

// IsAutomigrateEnabled returns whether automigrate is enabled
func (st *Store) IsAutomigrateEnabled() bool {
	return st.automigrateEnabled
}
