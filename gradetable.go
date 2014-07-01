package main

import (
	"github.com/linlexing/pghelper"
)

type GradeTable struct {
	*pghelper.DataTable
	Grade string
}

func NewGradeTable() *GradeTable {
	return &GradeTable{}
}
