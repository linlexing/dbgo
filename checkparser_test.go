package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"
	"testing"
)

func TestJSParse(t *testing.T) {

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
		maina +'c' =='cc'&&
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

}
