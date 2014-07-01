package grade

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestExport(t *testing.T) {
	p := NewPGHelper("host=localhost database=postgres user=meta password=meta123 sslmode=disable")
	tab, _ := p.Table("lx_check")
	err := tab.Export(Grade("root"), "d:/temp/meta", []struct{ ColumnName, FileExt string }{struct{ ColumnName, FileExt string }{"script", ".js"}}, "", "")
	if err != nil {
		t.Error(err)
	}
}
func TestImport(t *testing.T) {
	err := RunAtTrans("host=localhost database=postgres user=meta password=meta123 sslmode=disable",
		func(p *PGHelper) error {
			return Import(p, "d:/temp/meta/lx_check")

		})
	if err != nil {
		spew.Dump(err)
		t.Error(err)
	}
}
