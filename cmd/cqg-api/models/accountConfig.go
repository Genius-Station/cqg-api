package models

import (
	"time"
)

type AccountConfig struct {
	ID                     		int64      `db:"id"`
	AccountID              		int64      `db:"account_id"`
	DailyProfitCap         		*float64   `db:"daily_profit_cap"`
	DailyProfitCapQuantifier 	*string    `db:"daily_profit_cap_quantifier"`
	Monoside               		*string    `db:"monoside"`
	LooseLimit            		*float64   `db:"loose_limit"`
	Leverage              		*float64   `db:"leverage"`
	LooseLimitQuantifier   		*string    `db:"loose_limit_quantifier"`
	Profitability         		*float64   `db:"profitability"`
	MaxLots               		*int64     `db:"max_lots"`
	CreatedAt             		time.Time  `db:"created_at"`
	UpdatedAt             		time.Time  `db:"updated_at"`
}
