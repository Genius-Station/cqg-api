package models

import (
	"time"
)

type AccountSymbol struct {
	ID        int64     `db:"id"`
	AccountID int64     `db:"account_id"`
	SymbolID  int64     `db:"symbol_id"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
