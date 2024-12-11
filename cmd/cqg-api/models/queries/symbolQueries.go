package queries

import (
	"database/sql"
	"log"
	"cqg-api/cmd/cqg-api/models"
)

func GetSymbolList(db *sql.DB) ([]models.Symbol, error) {
	
	var symbols []models.Symbol

	rows, err := db.Query("SELECT id, symbol_name, name, market, category FROM symbols")
	if err != nil {
		log.Println("Error executing query:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var symbol models.Symbol
		err := rows.Scan(&symbol.ID, &symbol.SymbolName, &symbol.Name, &symbol.Market, &symbol.Category)
		if err != nil {
			log.Println("Error get symbol list : scanning row:", err)
			return nil, err
		}
		symbols = append(symbols, symbol)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error get symbol list : iterating rows:", err)
		return nil, err
	}

	return symbols, nil
}


func GetSymbolIDByName(db *sql.DB, symbolName string) *int64 {
  
    var symbol models.Symbol

    row := db.QueryRow("SELECT id, symbol_name, name, market, category FROM symbols WHERE symbol_name = ?", symbolName)
    
    err := row.Scan(&symbol.ID, &symbol.SymbolName, &symbol.Name, &symbol.Market, &symbol.Category)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil
        }
        log.Println("Error scanning row:", err)
        return nil
    }

    return &symbol.ID
}
