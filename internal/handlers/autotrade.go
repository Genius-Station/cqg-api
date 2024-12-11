package handlers

import (
	"cqg-api/cmd/cqg-api/models/queries"
	"cqg-api/pkg/cqgapi"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"path"
	"strconv"
)


func AutotradeOrderHandler (w http.ResponseWriter, r *http.Request , db *sql.DB ){


	alertIDStr := path.Base(r.URL.Path) 
	alertID, err := strconv.ParseInt(alertIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid alert_id", http.StatusBadRequest)
		return
	}

	// TODO GET ALERT INFO 

	log.Printf("autotrade alert id %v", alertID)

	// TODO GET ALL ACCOUNTS VALIDS	
	accounts, err := queries.GetActiveAccounts(db)
	if err != nil {
		http.Error(w, "Get active accounts error", http.StatusInternalServerError)
		return
	}

	accountsJSON, err := json.Marshal(accounts)
	if err != nil {
		log.Printf("Error marshalling accounts to JSON: %v", err)
	} else {
		log.Printf("autotrade alert id %v, accounts: %s", alertID, accountsJSON)
	}

	// TODO: Filter autotrade config for this account
	for _, account := range accounts {
		todayProfit := float64(10) //TODO 
		summary := float64(100)   //TODO 
		alert := cqgapi.Alert{}
		if valid := cqgapi.AccountIsValid(db , *account, alert , todayProfit, summary); err != nil {
			log.Printf("Error validating account ID %d: %v", account.ID, err)
			continue
		} else if !valid {
			log.Printf("Account ID %d is not valid for autotrade", account.ID)
			continue
		}

		log.Printf("Account ID %d is valid for autotrade", account.ID)

		
		volume := cqgapi.GetQuantity(summary, alert, account.Config.MaxLots)

        log.Printf("Volume to trade for account ID %d: %f", account.ID, volume)

		// TODO: Send order for this account
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Autotrade processing completed"))

}




