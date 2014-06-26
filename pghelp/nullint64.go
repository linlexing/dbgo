package pghelp

import (
	"database/sql"
)

type NullInt64 struct {
	sql.NullInt64
}

func (this NullInt64) IsNull() bool {
	return !this.Valid
}
