package main

import (
	"github.com/davecgh/go-spew/spew"
	"jsmvc/pghelp"
	"testing"
)

/*func TestJSParse(t *testing.T) {

	src := `
		$C1.count()`
	f, err := parser.ParseFunction("", src)
	if err != nil {
		t.Error(err)
	}
	spew.Dump(f.Body.(*ast.BlockStatement).List[0].(*ast.ExpressionStatement).Expression)
}

func TestPrint(t *testing.T) {
	src := `
		main.a +'c' =='cc'&&
		(substr(b+c,1,2)+'c'=="abc"||
		c+"test"=="test")&&
		COALESCE(a,0)==0&&
		!b||
		b==true&&
		a==null ||
		(c=='a\nb' &&
		a !=null &&
		a!=0 &&
		a > 0 ||
		a >=1 ||
		a < 2 ||
		a <= 3)`
	spew.Print(ParseToSql(src, "c"))

}*/
func TestBillCheck(t *testing.T) {
	mainT := pghelp.NewDataTable("测试")
	mainT.AddColumn(pghelp.NewColumn("s", pghelp.TypeString, true))
	mainT.AddColumn(pghelp.NewColumn("i", pghelp.TypeInt64))
	mainT.AddColumn(pghelp.NewColumn("f", pghelp.TypeFloat64))
	mainT.SetPK("s")
	childT := pghelp.NewDataTable("测试_dx")
	childT.AddColumn(pghelp.NewColumn("main_s", pghelp.TypeString, true))
	childT.AddColumn(pghelp.NewColumn("i", pghelp.TypeInt64, true))
	childT.AddColumn(pghelp.NewColumn("f", pghelp.TypeFloat64))
	childT.SetPK("main_s", "i")
	mainT.Desc = &pghelp.TableDesc{
		Relations: map[string]*pghelp.MainChildRelation{
			"测试_dx": &pghelp.MainChildRelation{
				MainColumns:  []string{"s"},
				ChildColumns: []string{"main_s"},
			},
		},
	}
	b := &Bill{
		Name: "测试",
		Tables: map[string]*pghelp.DataTable{
			"测试":    mainT,
			"测试_dx": childT,
		},
	}
	spew.Dump(b.ParseCheckToSql("测试", "true || false&& true"))
	spew.Dump(b.ParseCheckToSql("测试", "测试_dx.count()>0"))
	spew.Dump(b.ParseCheckToSql("测试", "s+i!=\"0\""))
	spew.Dump(b.ParseCheckToSql("测试", "exists(\"select 1 from aa where a=$1\",f)||query(\"select count(*) from aa where a=$1\",s)>0"))
	spew.Dump(b.ParseCheckToSql("测试_dx", "测试.i>0"))
}
