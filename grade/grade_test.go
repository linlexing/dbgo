package grade

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestExport(t *testing.T) {
	p := NewPGHelper("host=localhost database=postgres user=meta password=meta123 sslmode=disable")
	err := p.Export(&ExportParam{
		TableName:        "lx_check",
		CurrentGrade:     Grade("root"),
		PathName:         "d:/temp/meta",
		FileColumns:      map[string]string{"script": ".js"},
		SqlWhere:         "'root' like grade||'%'",
		ImpAutoRemove:    true,
		SqlRunAtImport:   "",
		ImpRefreshStruct: true})
	if err != nil {
		t.Error(err)
	}
}
func TestImport(t *testing.T) {
	err := RunAtTrans("host=localhost database=postgres user=meta password=meta123 sslmode=disable",
		func(p *PGHelper) error {
			return p.Import("d:/temp/meta_cpy/lx_check")

		})
	if err != nil {
		t.Error(err)
	}
}
func TestVersion(t *testing.T) {
	p := NewPGHelper("host=localhost database=postgres user=meta password=meta123 sslmode=disable")
	spew.Dump(p.Version("root/tjj"))
}
func TestGradeVersion(t *testing.T) {
	v1 := GradeVersion{1, 2}
	if !v1.GE(GradeVersion{1, 2}) {
		t.Error("error")
	}
	if v1.GE(GradeVersion{1, 3}) {
		t.Error("error")
	}
	if v1.GE(GradeVersion{1, 2, 1}) {
		t.Error("error")
	}
	if !v1.GE(GradeVersion{1}) {
		t.Error("error")
	}
	if !v1.GE(GradeVersion{1, 1, 1}) {
		t.Error("error")
	}

}
