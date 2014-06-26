package pghelp

import (
	"github.com/lib/pq"
)

type NullTime struct {
	pq.NullTime
}

func (this NullTime) IsNull() bool {
	return !this.Valid
}
