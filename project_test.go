package main

import (
	"github.com/linlexing/dbgo/grade"
	"os"
	"path/filepath"
	"testing"
)

func TestExportData(t *testing.T) {
	p := NewProject("meta", TranslateString{}, "postgres host=localhost database=postgres user=meta password=meta123 sslmode=disable", "meta")
	expFile, err := os.Create(filepath.Join(os.TempDir(), "export_test.zip"))
	if err != nil {
		t.Error(err)
	}
	defer expFile.Close()
	if err := p.ExportData("package", expFile, grade.GRADE_ROOT); err != nil {
		t.Error(err)
	}
}
