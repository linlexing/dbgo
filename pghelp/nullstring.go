package pghelp

import (
	"database/sql"
)

type NullString struct {
	sql.NullString
}

func (this NullString) IsNull() bool {
	return !this.Valid
}
