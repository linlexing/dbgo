package pghelp

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"testing"
	"time"
)

func CreateTable() *DataTable {
	table := NewDataTable("table1")
	table.AddColumn(NewColumn("id", TypeString, true))
	table.AddColumn(NewColumn("name", TypeString, true))
	table.AddColumn(NewColumn("f", TypeFloat64))
	table.AddValues("01", "name1", 0.01)
	table.AddValues("02", "name2", 0.02)
	table.AcceptChange()
	table.SetPK("id", "name")
	return table
}

func Test_buildSql(t *testing.T) {
	table := CreateTable()
	table.AddValues("03", "name3", 0.003)
	table.SetValues(0, "01", "name4", 0.004)
	table.DeleteRow(1)
	fmt.Print("delete sql:\n", buildDeleteSql(table), "\ninsert sql:\n", buildInsertSql(table), "\nupdate sql:\n", buildUpdateSql(table))
}

func TestPQScaleToInterface(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost database=postgres user=meta password=meta123 sslmode=disable")
	if err != nil {
		t.Error(t)
	}
	defer func() { err = db.Close() }()
	if rows, err := db.Query("SELECT '{\"a\":1}'::jsonb as aa ,true as cc,cast(1212 as bigint) as bb, cast(b.nspname as text) as nspname,a.description ,now() as tt FROM pg_namespace b left join pg_description a on a.objoid = b.oid union all"+
		" SELECT null as aa ,null as cc,null as bb, cast(b.nspname as text) as nspname,a.description ,now() as tt FROM pg_namespace b left join pg_description a on a.objoid = b.oid", []interface{}{}...); err != nil {
		t.Error(err)
	} else {
		for rows.Next() {
			cols, err := rows.Columns()
			if err != nil {
				t.Error(err)
			}
			if _, err = scanValues(rows, len(cols)); err != nil {
				t.Error(err)
			}
		}
	}
}
func TestConvert(t *testing.T) {
	var v interface{}
	v = NullBytea{}
	if !v.(IsNull).IsNull() {
		t.Error("error")
	}
}
func TestPQScaleToNullType(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost database=postgres user=meta password=meta123 sslmode=disable")
	if err != nil {
		t.Error(t)
	}
	defer func() { err = db.Close() }()
	if rows, err := db.Query("SELECT null as aa ,true as cc,cast(1212 as bigint) as bb, b.nspname,a.description,now() as tt FROM pg_namespace b left join pg_description a on a.objoid = b.oid", []interface{}{}...); err != nil {
		t.Error(err)
	} else {
		for rows.Next() {
			_, err := rows.Columns()
			aa := sql.NullFloat64{}
			cc := sql.NullBool{}

			bb := sql.NullInt64{}
			nspname := sql.NullString{}
			desc := sql.NullString{}
			tt := pq.NullTime{}
			if err != nil {
				t.Error(err)
			}
			if err := rows.Scan(&aa, &cc, &bb, &nspname, &desc, &tt); err != nil {
				t.Error(err)
			}

		}
	}
}
func TestPQPutToNullType(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost database=postgres user=meta password=meta123 sslmode=disable")
	if err != nil {
		t.Error(t)
	}
	defer func() { err = db.Close() }()
	if rows, err := db.Query("SELECT cast($1 as character varying(10)) as aa ,cast($2 as float) as bb,cast($3 as timestamp with time zone) as cc", "dafadf", nil, time.Now()); err != nil {
		t.Error(err)
	} else {
		for rows.Next() {
			cols, err := rows.Columns()
			if err != nil {
				t.Error(err)
			}
			if v, err := scanValues(rows, len(cols)); err != nil {
				t.Error(err)
			} else {
				fmt.Print(v)
			}
		}
	}
}
func TestPQGetTable(t *testing.T) {
	ahelp := NewPGHelp("host=localhost database=postgres user=meta password=meta123 sslmode=disable")
	if tab, err := ahelp.GetDataTable("SELECT '{\"a\":1}'::jsonb as aa ,true as cc,cast(1212 as bigint) as bb, cast(b.nspname as text) as nspname,a.description ,now() as tt FROM pg_namespace b left join pg_description a on a.objoid = b.oid"); err != nil {
		t.Error(err)
	} else {
		fmt.Print(tab.AsTabText())
	}
	if tab, err := ahelp.GetDataTable("SELECT 'abc' as aa"); err != nil {
		t.Error(err)
	} else {
		if _, ok := tab.GetValue(0, 0).(string); !ok {
			t.Error("error", tab.GetValue(0, 0))
		}
	}

}
func TestCreateTable(t *testing.T) {
	ahelp := NewPGHelp("host=localhost database=postgres user=meta password=meta123 sslmode=disable")
	if err := ahelp.ExecuteSql("create table version()"); err != nil {
		t.Error(err)
	} else {
		ahelp.ExecuteSql("drop table version")
	}
}
func TestFillTableForArray(t *testing.T) {
	dburl := "host=localhost database=postgres user=meta password=meta123 sslmode=disable"
	ahelp := NewPGHelp(dburl)
	tab := NewDataTable("tab")
	tab.AddColumn(NewColumnT("c1", NewPGType(TypeStringSlice, 0, true), ""))
	tab.AddColumn(NewColumn("c2", TypeInt64Slice, true))
	tab.AddColumn(NewColumn("c3", TypeJSONSlice))
	tab.AddColumn(NewColumn("c4", TypeJSON))
	tab.SetPK("c1", "c2")
	if err := ahelp.FillTable(tab, `select '{"234","343a4"}'::text[] c1,'{123,11}'::bigint[] c2,'{"{\"a\":1,\"b\":\"334\"}","{\"a\":2,\"b\":\"aaa\"}"}'::jsonb[] c3,'{"a":1,"b":"333","c":{"c1":33.34}}'::jsonb c4`); err != nil {
		t.Error(err)
	}
	fmt.Println(tab.AsTabText())
}
