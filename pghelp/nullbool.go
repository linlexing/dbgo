package pghelp

import (
	"database/sql"
)

type NullBool struct {
	sql.NullBool
}

func (this NullBool) IsNull() bool {
	return !this.Valid
}
