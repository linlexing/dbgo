package pghelp

const (
	SQL_TableColumns = `
		SELECT
		  a.attname as columnname,
		  a.attnotnull as notnull,
		  d.adsrc AS def,
		  pg_catalog.format_type(a.atttypid, a.atttypmod) AS datatype,
		  col_description(b.oid,a.attnum) as desc
		FROM
		  pg_catalog.pg_attribute a join
		  (SELECT  c.oid
		   FROM    pg_catalog.pg_class c LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
		   WHERE c.relname =$1 AND (n.nspname) = current_schema
		  ) b on a.attrelid = b.oid left join
		  pg_catalog.pg_attrdef d ON (a.attrelid, a.attnum) = (d.adrelid,  d.adnum)
		WHERE

		  a.attnum > 0 AND
		  NOT a.attisdropped
		ORDER BY
		  a.attnum`
	SQL_TableIndexes = `
		select
		  c.relname as indexname,
		  pg_get_indexdef(i.indexrelid) as define,
		  obj_description(c.oid) as desc
		from
		  pg_index i inner JOIN pg_class c ON c.oid = i.indexrelid
		where
		  indrelid = $1::regclass and
		  indisprimary = false`
	SQL_TablePrimaryKeys = `
		SELECT
		  pg_attribute.attname as columnname,
		  idx.relname as indexname
		FROM pg_index, pg_class, pg_attribute ,pg_class idx
		WHERE
		  pg_class.oid = $1::regclass AND
		  pg_index.indrelid = pg_class.oid AND
		  pg_attribute.attrelid = pg_class.oid AND
		  pg_index.indexrelid = idx.oid and
		  pg_attribute.attnum = any(pg_index.indkey) AND
		  indisprimary`
	SQL_TableExists = `
	SELECT EXISTS(
	    SELECT *
	    FROM information_schema.tables
	    WHERE
	      table_schema = current_schema AND
	      table_name = $1
	)`
	SQL_GetTableDesc            = "select obj_description($1::regclass,'pg_class')"
	SQL_GetCurrentSchemaAndDesc = "SELECT b.nspname,a.description FROM pg_namespace b left join pg_description a on a.objoid = b.oid WHERE b.nspname=current_schema"
	SQL_GetTableCheck           = "select id,displaylabel,level,fields,script,grade from lx_check where tablename=$1"

	SQL_DropConstraint    = "ALTER TABLE %v DROP CONSTRAINT %v"
	SQL_CreatePrimaryKey  = "ALTER TABLE %v ADD PRIMARY KEY(%v)"
	SQL_DropColumn        = "ALTER TABLE %v DROP COLUMN %v"
	SQL_DropColumnNotNull = "ALTER TABLE %v ALTER COLUMN %v DROP NOT NULL"
	SQL_SetColumnNotNull  = "ALTER TABLE %v ALTER COLUMN %v SET NOT NULL"
	SQL_DropColumnDefault = "ALTER TABLE %v ALTER COLUMN %v DROP DEFAULT"
	SQL_SetColumnDefault  = "ALTER TABLE %v ALTER COLUMN %v SET DEFAULT %v"
	SQL_RenameColumn      = "ALTER TABLE %v RENAME %v TO %v"
	SQL_AlterColumnType   = "ALTER TABLE %v ALTER COLUMN %v TYPE %v"
	SQL_CreateColumn      = "ALTER TABLE %v ADD COLUMN %v %v %v"
	SQL_CreateTable       = "CREATE TABLE %v()"
	SQL_DropIndex         = "DROP INDEX %v"
	SQL_AlterColumnDesc   = "COMMENT ON COLUMN %v.%v IS %s"
	SQL_AlterTableDesc    = "COMMENT ON TABLE %v IS %s"
	SQL_AlterIndexDesc    = "COMMENT ON INDEX %v IS %s"
	SQL_AlterSchemaDesc   = "COMMENT ON SCHEMA %v IS %s"
)
