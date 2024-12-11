package models

type Symbol struct {
	ID         int64  `db:"id"`
	SymbolName string `db:"symbol_name"`
	Name       string `db:"name"`
	Market     string `db:"market"`
	Category   string `db:"category"`
}
