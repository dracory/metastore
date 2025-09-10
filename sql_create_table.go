package metastore

import "github.com/dracory/sb"

// SqlCreateTable returns a SQL string for creating the setting table
func (st *Store) SqlCreateTable() string {
	sql := sb.NewBuilder(sb.DatabaseDriverName(st.db)).
		Table(st.metaTableName).
		Column(sb.Column{
			Name:       COLUMN_ID,
			Type:       sb.COLUMN_TYPE_STRING,
			Length:     40,
			PrimaryKey: true,
		}).
		Column(sb.Column{
			Name:   COLUMN_OBJECT_TYPE,
			Type:   sb.COLUMN_TYPE_STRING,
			Length: 100,
		}).
		Column(sb.Column{
			Name:   COLUMN_OBJECT_ID,
			Type:   sb.COLUMN_TYPE_STRING,
			Length: 40,
		}).
		Column(sb.Column{
			Name: COLUMN_META_KEY,
			Type: sb.COLUMN_TYPE_STRING,
		}).
		Column(sb.Column{
			Name: COLUMN_META_VALUE,
			Type: sb.COLUMN_TYPE_LONGTEXT,
		}).
		Column(sb.Column{
			Name: COLUMN_CREATED_AT,
			Type: sb.COLUMN_TYPE_DATETIME,
		}).
		Column(sb.Column{
			Name: COLUMN_UPDATED_AT,
			Type: sb.COLUMN_TYPE_DATETIME,
		}).
		Column(sb.Column{
			Name:     COLUMN_DELETED_AT,
			Type:     sb.COLUMN_TYPE_DATETIME,
			Nullable: true,
		}).
		CreateIfNotExists()

	return sql
}
