package queries

import (
	"database/sql"
	"errors"
	"fmt"
	"log"	
	"cqg-api/cmd/cqg-api/models"
)

func GetAccountByUserID(db *sql.DB, userID int64) (*models.Account, error) {
	var account models.Account
	var config models.AccountConfig
	query := `SELECT 
				a.id, a.user_id, a.username, a.password, a.access_token, a.autotrade_active, a.valid_credentials, a.created_at, a.updated_at,
				ac.id, ac.account_id, ac.daily_profit_cap, ac.daily_profit_cap_quantifier, ac.monoside, ac.loose_limit, ac.leverage, 
				ac.loose_limit_quantifier, ac.profitability, ac.max_lots, ac.created_at, ac.updated_at
			FROM accounts a 
			LEFT JOIN account_config ac ON a.id = ac.account_id 
			WHERE user_id = $1`
	err := db.QueryRow(query, userID).Scan(
		&account.ID,
		&account.UserID,
		&account.Username,
		&account.Password,
		&account.AccessToken,
		&account.AutotradeActive,
		&account.ValidCredentials,
		&account.CreatedAt,
		&account.UpdatedAt,
		&config.ID,
		&config.AccountID,
		&config.DailyProfitCap,
		&config.DailyProfitCapQuantifier,
		&config.Monoside,
		&config.LooseLimit,
		&config.Leverage,
		&config.LooseLimitQuantifier,
		&config.Profitability,
		&config.MaxLots,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("account not found")
		}
		log.Printf("get account error %v", err)
		return nil, err
	}

	account.Config = &config

	querySymbols := `SELECT symbol_id FROM account_symbols WHERE account_id = $1 AND is_active = TRUE`
	rows, err := db.Query(querySymbols, account.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []int64
	for rows.Next() {
		var symbolID int64
		if err := rows.Scan(&symbolID); err != nil {
			return nil, err
		}
		symbols = append(symbols, symbolID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	account.Symbols = symbols

	return &account, nil
}


type AccountPayload struct {
    ID                            int64     `json:"id,omitempty"`  
    UserID                        int64     `json:"user_id"`
    Username                      string    `json:"username"`
    Password                      string    `json:"password"`
    AccessToken                   *string   `json:"access_token,omitempty"`  
    AutotradeActive               bool      `json:"autotrade_active"`
    ValidCredentials              bool      `json:"valid_credentials"`
    DailyProfitCap                *float64  `json:"daily_profit_cap,omitempty"`
    DailyProfitCapQuantifier      string    `json:"daily_profit_cap_quantifier"`
    Monoside                      string    `json:"monoside"`
    LooseLimit                    *float64  `json:"loose_limit,omitempty"`
    Leverage                      *float64  `json:"leverage,omitempty"`
    LooseLimitQuantifier          string    `json:"loose_limit_quantifier"`
    Profitability                 *float64  `json:"profitability,omitempty"`
    MaxLots                       *int      `json:"max_lots,omitempty"`
	Symbols 					  []int64   `json:"symbols,omitempty"` 
}


func CreateAccount(db *sql.DB, data *AccountPayload) (int64, error) {
	
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() 

	queryAccount := `
		INSERT INTO accounts (user_id, username, password, access_token, autotrade_active, valid_credentials)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var accessToken interface{}
	if data.AccessToken != nil {
		accessToken = *data.AccessToken
	} else {
		accessToken = nil
	}

	validCredentials := false

	var accountID int64
	err = tx.QueryRow(
		queryAccount,
		data.UserID,
		data.Username,
		data.Password,
		accessToken,
		data.AutotradeActive,
		validCredentials,
	).Scan(&accountID)

	
	if err != nil {
		return 0, err
	}
	
	queryConfig := `
		INSERT INTO account_config (account_id, daily_profit_cap, daily_profit_cap_quantifier, monoside, loose_limit, leverage, 
			loose_limit_quantifier, profitability, max_lots)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var configID int64
	err = tx.QueryRow(
		queryConfig,
		accountID, 
		data.DailyProfitCap,
		data.DailyProfitCapQuantifier,
		data.Monoside,
		data.LooseLimit,
		data.Leverage,
		data.LooseLimitQuantifier,
		data.Profitability,
		data.MaxLots,
	).Scan(&configID)
	
	if err != nil {
		return 0, err
	}
	
	if len(data.Symbols) > 0 {
		querySymbol := `
			INSERT INTO account_symbols (account_id, symbol_id, is_active)
			VALUES ($1, $2, $3)
		`

		for _, symbolID := range data.Symbols {
			_, err := tx.Exec(querySymbol, accountID, symbolID, true)
			if err != nil {
				return 0, err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return accountID, nil
}



func UpdateAccount(db *sql.DB, data *AccountPayload) error {
	
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	
	defer tx.Rollback()

	
	queryAccount := `UPDATE accounts SET `
	paramsAccount := []interface{}{}
	counterAccount := 1

	if data.Username != "" {
		queryAccount += "username = $" + fmt.Sprintf("%d", counterAccount) + ", "
		paramsAccount = append(paramsAccount, data.Username)
		counterAccount++
	}

	if data.Password != "" {
		queryAccount += "password = $" + fmt.Sprintf("%d", counterAccount) + ", "
		paramsAccount = append(paramsAccount, data.Password)
		counterAccount++
	}

	if data.AccessToken != nil {
		queryAccount += "access_token = $" + fmt.Sprintf("%d", counterAccount) + ", "
		paramsAccount = append(paramsAccount, *data.AccessToken)
		counterAccount++
	}

	queryAccount += "updated_at = CURRENT_TIMESTAMP WHERE id = $" + fmt.Sprintf("%d", counterAccount)
	paramsAccount = append(paramsAccount, data.ID)

	_, err = tx.Exec(queryAccount, paramsAccount...)
	if err != nil {
		return err
	}

	queryConfig := `UPDATE account_config SET `
	paramsConfig := []interface{}{}
	counterConfig := 1

	fields := []struct {
		name  string
		value interface{}
	}{
		{"daily_profit_cap", data.DailyProfitCap},
		{"daily_profit_cap_quantifier", data.DailyProfitCapQuantifier},
		{"monoside", data.Monoside},
		{"loose_limit", data.LooseLimit},
		{"leverage", data.Leverage},
		{"loose_limit_quantifier", data.LooseLimitQuantifier},
		{"profitability", data.Profitability},
		{"max_lots", data.MaxLots},
	}

	for _, field := range fields {
		if field.value != nil {
			queryConfig += fmt.Sprintf("%s = $%d, ", field.name, counterConfig)
			paramsConfig = append(paramsConfig, field.value)
			counterConfig++
		}
	}

	if len(paramsConfig) > 0 {
		queryConfig = queryConfig[:len(queryConfig)-2]
	}

	queryConfig += " WHERE account_id = $" + fmt.Sprintf("%d", counterConfig)
	paramsConfig = append(paramsConfig, data.ID)

	_, err = tx.Exec(queryConfig, paramsConfig...)
	if err != nil {
		return err
	}

	if data.Symbols != nil { 
		
		disableSymbolsQuery := `
			UPDATE account_symbols 
			SET is_active = FALSE 
			WHERE account_id = $1
		`
		_, err = tx.Exec(disableSymbolsQuery, data.ID)
		if err != nil {
			return err
		}

		activateSymbolQuery := `
			INSERT INTO account_symbols (account_id, symbol_id, is_active)
			VALUES ($1, $2, TRUE)
			ON CONFLICT (account_id, symbol_id) 
			DO UPDATE SET is_active = EXCLUDED.is_active
		`
		for _, symbolID := range data.Symbols {
			_, err = tx.Exec(activateSymbolQuery, data.ID, symbolID)
			if err != nil {
				return err
			}
		}
	}


	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}


func GetActiveAccounts(db *sql.DB) ([]*models.Account, error) {
	var accounts []*models.Account
	var config models.AccountConfig

	query := `SELECT 
				a.id, a.user_id, a.username, a.password, a.access_token, a.autotrade_active, a.valid_credentials, a.created_at, a.updated_at,
				ac.id, ac.account_id, ac.daily_profit_cap, ac.daily_profit_cap_quantifier, ac.monoside, ac.loose_limit, ac.leverage, 
				ac.loose_limit_quantifier, ac.profitability, ac.max_lots, ac.created_at, ac.updated_at
			FROM accounts a 
			LEFT JOIN account_config ac ON a.id = ac.account_id 
			WHERE a.autotrade_active = TRUE AND a.valid_credentials = TRUE`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("get active accounts error %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var account models.Account
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Username,
			&account.Password,
			&account.AccessToken,
			&account.AutotradeActive,
			&account.ValidCredentials,
			&account.CreatedAt,
			&account.UpdatedAt,
			&config.ID,
			&config.AccountID,
			&config.DailyProfitCap,
			&config.DailyProfitCapQuantifier,
			&config.Monoside,
			&config.LooseLimit,
			&config.Leverage,
			&config.LooseLimitQuantifier,
			&config.Profitability,
			&config.MaxLots,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			log.Printf("error scanning account: %v", err)
			return nil, err
		}

		account.Config = &config

		querySymbols := `SELECT symbol_id FROM account_symbols WHERE account_id = $1 AND is_active = TRUE`
		symbolRows, err := db.Query(querySymbols, account.ID)
		if err != nil {
			log.Printf("error querying symbols: %v", err)
			return nil, err
		}
		defer symbolRows.Close()

		var symbols []int64
		for symbolRows.Next() {
			var symbolID int64
			if err := symbolRows.Scan(&symbolID); err != nil {
				log.Printf("error scanning symbol: %v", err)
				return nil, err
			}
			symbols = append(symbols, symbolID)
		}
		if err := symbolRows.Err(); err != nil {
			log.Printf("error iterating symbols: %v", err)
			return nil, err
		}

		account.Symbols = symbols

		accounts = append(accounts, &account)
	}

	if err := rows.Err(); err != nil {
		log.Printf("error iterating accounts: %v", err)
		return nil, err
	}

	return accounts, nil
}


func DeleteAccount(db *sql.DB, id int64) error {
	query := "DELETE FROM accounts WHERE id = $1"
	_, err := db.Exec(query, id)
	return err
}
