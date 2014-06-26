package pghelp

import (
	"database/sql"
)

type NullFloat64 struct {
	sql.NullFloat64
}

func (this NullFloat64) IsNull() bool {
	return !this.Valid
}
