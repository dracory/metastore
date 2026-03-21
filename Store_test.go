package metastore

import (
	"os"
	"testing"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(filepath string) *sql.DB {
	os.Remove(filepath) // remove database
	dsn := filepath + "?parseTime=true"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		panic(err)
	}

	return db
}

func InitStore() *Store {
	db := InitDB("test_metastore_automigrate.db")
	return &Store{
		metaTableName:      "test_metastore_automigrate.db",
		db:                 db,
		automigrateEnabled: false,
	}
}

// func TestWithAutoMigrate(t *testing.T) {
// 	db := InitDB("test_metastore_automigrate.db")

// 	s := Store{
// 		metaTableName:      "log_with_automigrate_false",
// 		db:                 db,
// 		automigrateEnabled: false,
// 	}

// 	f := WithAutoMigrate(true)
// 	f(&s)

// 	if s.automigrateEnabled != true {
// 		t.Fatalf("automigrateEnabled: Expected [true] received [%v]", s.automigrateEnabled)
// 	}

// 	s = Store{
// 		metaTableName:      "log_with_automigrate_true",
// 		db:                 db,
// 		automigrateEnabled: true,
// 	}

// 	f = WithAutoMigrate(false)
// 	f(&s)

// 	if s.automigrateEnabled == true {
// 		t.Fatalf("automigrateEnabled: Expected [true] received [%v]", s.automigrateEnabled)
// 	}
// }

// func TestWithDb(t *testing.T) {
// 	s := Store{
// 		metaTableName:      "log_with_automigrate_true",
// 		db:                 nil,
// 		automigrateEnabled: true,
// 	}

// 	db := InitDB("test")
// 	f := WithDb(db)
// 	f(&s)

// 	if s.db == nil {
// 		t.Fatalf("DB: Expected Initialized DB, received [%v]", s.db)
// 	}

// }

// func TestWithTableName(t *testing.T) {
// 	s := Store{
// 		metaTableName:      "",
// 		db:                 nil,
// 		automigrateEnabled: false,
// 	}
// 	table_name := "Table1"
// 	f := WithTableName(table_name)
// 	f(&s)
// 	if s.metaTableName != table_name {
// 		t.Fatalf("Expected logTableName [%v], received [%v]", table_name, s.metaTableName)
// 	}
// 	table_name = "Table2"
// 	f = WithTableName(table_name)
// 	f(&s)
// 	if s.metaTableName != table_name {
// 		t.Fatalf("Expected logTableName [%v], received [%v]", table_name, s.metaTableName)
// 	}
// }

func Test_Store_AutoMigrate(t *testing.T) {
	db := InitDB("test_metastore_automigrate.db")

	s, err := NewStore(NewStoreOptions{
		DB:                 db,
		MetaTableName:      "metas",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("Error at AutoMigrate", err.Error())
	}

	s.AutoMigrate()

	if s.GetMetaTableName() != "metas" {
		t.Fatalf("Expected metaTableName [metas] received [%v]", s.GetMetaTableName())
	}
	if s.GetDB() == nil {
		t.Fatalf("DB Init Failure")
	}
	if s.IsAutomigrateEnabled() != true {
		t.Fatalf("Failure:  WithAutoMigrate")
	}
}

func Test_Store_Set(t *testing.T) {
	db := InitDB("test_metastore_set.db")
	s, err := NewStore(NewStoreOptions{
		DB:                 db,
		MetaTableName:      "metas",
		AutomigrateEnabled: true,
		DebugEnabled:       true,
	})

	if err != nil {
		t.Fatal("Error at AutoMigrate", err.Error())
	}

	objType := "Test_Obj"
	objID := "12345"
	key := "1234z"
	val := "123zx"
	errSet := s.Set(objType, objID, key, val)

	if errSet != nil {
		t.Fatal("Failure: Set", errSet.Error())
	}
}

func Test_Store_SetJSON(t *testing.T) {
	db := InitDB("test_metastore_set_json.db")
	s, err := NewStore(NewStoreOptions{
		DB:                 db,
		MetaTableName:      "metas",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("Error at AutoMigrate", err.Error())
	}

	objType := "Test_Obj"
	objID := "12345"
	key := "1234z"
	val := `{"a" : "b", "c" : "d"}`
	errSetJSON := s.SetJSON(objType, objID, key, val)

	if errSetJSON != nil {
		t.Fatal("Failure: SetJSON", errSetJSON.Error())
	}
}

func Test_Store_Remove(t *testing.T) {
	db := InitDB("test_metastore_remove.db")
	s, err := NewStore(NewStoreOptions{
		DB:                 db,
		MetaTableName:      "metas",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("Error at AutoMigrate", err.Error())
	}

	objType := "Test_Obj"
	objID := "12345"
	key := "1234z"
	val := "123zx"
	errSet := s.Set(objType, objID, key, val)

	if errSet != nil {
		t.Fatal("Failure at Remove: Set", errSet.Error())
	}

	errRemove := s.Remove(objType, objID, key)

	if errRemove != nil {
		t.Fatal("Failure at Remove: Remove", errRemove.Error())
	}

	ret, errGet := s.Get(objType, objID, key, "default")

	if errGet != nil {
		t.Fatal("Failure at Remove: Get", errGet.Error())
	}

	if ret != "default" {
		t.Fatal("Unable to delete!!! Entry Persists")
	}
}

func Test_Store_Get(t *testing.T) {
	db := InitDB("test_metastore_get.db")
	s, err := NewStore(NewStoreOptions{
		DB:                 db,
		MetaTableName:      "metas",
		AutomigrateEnabled: true,
		DebugEnabled:       true,
	})

	if err != nil {
		t.Fatal("Error at Test_Store_Get:", err.Error())
	}

	objType := "OBJECT_TYPE"
	objID := "OBJECT_ID"
	key := "OBJECT_KEY"
	val := "OBJECT_VALUE"
	errSet := s.Set(objType, objID, key, val)

	if errSet != nil {
		t.Fatal("Failure at Test_Store_Get: Set", errSet.Error())
	}

	ret, errGet := s.Get(objType, objID, key, "default")

	if errGet != nil {
		t.Fatal("Failure at Test_Store_Get:", errGet.Error())
	}

	if ret != val {
		t.Fatalf("Unable to Test_Store_Get: Expected [%v] Received [%v]", val, ret)
	}
}

func Test_Store_FindByKey(t *testing.T) {
	db := InitDB("test_metastore_findbykey.db")
	s, err := NewStore(NewStoreOptions{
		DB:                 db,
		MetaTableName:      "metas",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("Error at AutoMigrate", err.Error())
	}

	objType := "Test_Obj"
	objID := "12345"
	key := "1234z"
	val := "123zx"
	errSet := s.Set(objType, objID, key, val)

	if errSet != nil {
		t.Fatal("Failure at FindByKey: Set", errSet.Error())
	}

	meta, errFindByKey := s.FindByKey(objType, objID, key)

	if errFindByKey != nil {
		t.Fatal("Failure at FindByKey: FindbyKey", errFindByKey)
	}

	if meta.ObjectID != objID {
		t.Fatalf("Incorrect Record Received [%v]", meta)
	}
}
func Test_Store_GetJSON(t *testing.T) {
	db := InitDB("test_metastore_getjson.db")
	s, err := NewStore(NewStoreOptions{
		DB:                 db,
		MetaTableName:      "metas",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("Error at AutoMigrate", err.Error())
	}

	objType := "Test_Obj"
	objID := "12345"
	key := "1234z"
	val := `{"a" : "b", "c" : "d"}`
	errSetJSON := s.SetJSON(objType, objID, key, val)

	if errSetJSON != nil {
		t.Fatal("Failure as GetJSON: SetJSON", errSetJSON)
	}
	ret, errGetJSON := s.GetJSON(objType, objID, key, nil)

	if errGetJSON != nil {
		t.Fatal("Failure at GetJSON: GetJSON", errGetJSON.Error())
	}

	if ret == nil {
		t.Fatalf("Failure getting JSON value")
	}
}

func Test_Store_Update(t *testing.T) {
	db := InitDB("test_metastore_update.db")
	s, err := NewStore(NewStoreOptions{
		DB:                 db,
		MetaTableName:      "metas",
		AutomigrateEnabled: true,
		DebugEnabled:       true,
	})

	if err != nil {
		t.Fatal("Error at AutoMigrate", err.Error())
	}

	objType := "TESTOBJECT"
	objID := "OBJECTID"
	key := "OBJECTKEY"
	val := "OBJECTVALUE"
	val2 := "OBJECTVALUE2"
	errSet := s.Set(objType, objID, key, val)

	if errSet != nil {
		t.Fatal("Failure Update: Set", errSet.Error())
	}

	metaVal, errGet := s.Get(objType, objID, key, "")

	if errGet != nil {
		t.Fatal("Failure Update: Get", errGet.Error())
	}

	if metaVal != val {
		t.Fatal("Failure Update: Values do not match", metaVal, val)
	}

	errSet2 := s.Set(objType, objID, key, val2)

	if errSet2 != nil {
		t.Fatal("Failure Update: Set2", errSet2.Error())
	}

	metaVal2, errGet2 := s.Get(objType, objID, key, "")

	if errGet2 != nil {
		t.Fatal("Failure Update: Get2", errGet2.Error())
	}

	if metaVal2 != val2 {
		t.Fatal("Failure Update: Values 2 do not match", metaVal2, val2)
	}
}
