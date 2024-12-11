package models

import (
	"time"
)

type Account struct {
	ID               int64          		`db:"id"`
	UserID           int64          		`db:"user_id"`
	Username         string         		`db:"username"`
	Password         string         		`db:"password"`
	AccessToken      *string   				`db:"access_token"`
	AutotradeActive  bool           		`db:"autotrade_active"`
	ValidCredentials bool           		`db:"valid_credentials"`
	CreatedAt        time.Time      		`db:"created_at"`
	UpdatedAt        time.Time      		`db:"updated_at"`
	Config          *AccountConfig  		`json:"config,omitempty"`
	Symbols         []int64         		`json:"symbols,omitempty"` 
}


