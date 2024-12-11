package handlers

import (
	"cqg-api/cmd/cqg-api/models/queries"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)


func GetSymbolsHandler (db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		symbols, err := queries.GetSymbolList(db)

		if err != nil {
			http.Error(w, "Error retrieving symbols: "+err.Error(), http.StatusInternalServerError)
			log.Println("Error retrieving symbols:", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(symbols); err != nil {
			http.Error(w, "Error encoding response to JSON: "+err.Error(), http.StatusInternalServerError)
			log.Println("Error encoding response to JSON:", err)
			return
		}

	}
	
}




