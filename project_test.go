package main

import (
	"github.com/linlexing/dbgo/grade"
	"os"
	"path/filepath"
	"testing"
)

func TestExportData(t *testing.T) {
	p, err := NewProject("meat", "host=localhost database=postgres user=meta password=meta123 sslmode=disable", "meta")
	if err != nil {
		t.Error(err)
	}

	expFile, err := os.Create(filepath.Join(os.TempDir(), "export_test.zip"))
	if err != nil {
		t.Error(err)
	}
	defer expFile.Close()
	if err := p.ExportData("package", expFile, grade.GRADE_ROOT); err != nil {
		t.Error(err)
	}
}
