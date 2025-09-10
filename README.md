# Meta Store <a href="https://gitpod.io/#https://github.com/dracory/metastore" style="float:right:"><img src="https://gitpod.io/button/open-in-gitpod.svg" alt="Open in Gitpod" loading="lazy"></a>


[![Tests Status](https://github.com/dracory/metastore/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/dracory/metastore/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dracory/metastore)](https://goreportcard.com/report/github.com/dracory/metastore)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dracory/metastore)](https://pkg.go.dev/github.com/dracory/metastore)

Meta stores meta information for any object to a database table.

Store to database additional information to anything using metas (key - value) pairs

## Installation
```
go get -u github.com/dracory/metastore
```

## Table Schema ##

The following schema is used for the database.

| meta        |                  |
|-------------|------------------|
| id          | String, UniqueId |
| object_type | String (100)     |
| object_id   | String (40)      |
| meta_key    | String (255)     |
| meta_value  | Long Text        |
| created_at  | DateTime         |
| updated_at  | DateTime         |
| deleted_at  | DateTime         |

## Setup

```go
metaStore, err := metastore.NewStore(metastore.NewStoreOptions{
	DB:                 databaseInstance,
	MetaTableName:      "my_meta",
	AutomigrateEnabled: true,
	DebugEnabled:       false,
})

if err != nil {
    t.Fatal("Error at AutoMigrate", err.Error())
}
```



## Usage

- Set a meta values (for user with ID 1)
```go
metaStore.Set("user", "1", "verified", "yes")
metaStore.Set("user", "1", "verified_at", "2021-03-12")
```

- Get meta values (for user with ID 1), if not found a default value is returned
```go
log.Println(metaStore.Get("user", "1", "verified", ""))
log.Println(metaStore.Get("user", "1", "verified_at", ""))
```

## Changelog

2022.12.07 - Changed setup to use struct

2022.12.07 - Updated dependencies, fixed package name

2022.01.02 - Removed GORM dependency

2021.12.29 - Added tests badge

2021.12.29 - Added tests
