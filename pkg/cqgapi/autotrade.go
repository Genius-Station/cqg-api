package cqgapi

import (
	"cqg-api/cmd/cqg-api/models"
	"cqg-api/cmd/cqg-api/models/queries"
	"cqg-api/pkg/utils"
	"log"
	"strings"
	"time"
	"database/sql"
)

type Alert struct {
	ID             int64       `json:"id"`               
	Symbol         string      `json:"symbol"`           
	PricePrecision int         `json:"pricePrecision"`   
	Asset          string      `json:"asset"`            
	Instrument     string      `json:"instrument"`       
	Strategy       string      `json:"strategy"`         
	Start          float64     `json:"start"`            
	Target         float64     `json:"target"`           
	Stop           float64     `json:"stop"`            
	Sortie         *float64    `json:"sortie"`          
	Gain           *float64    `json:"gain"`             
	RR             float64     `json:"rr"`              
	Status         string      `json:"status"`          
	CreatedAt      time.Time   `json:"createdAt"`        
	StartedAt      *time.Time  `json:"startedAt"`       
	ClosedAt       *time.Time  `json:"closedAt"`         
	Orders         []Order     `json:"orders"`          
}

type Order struct {

}

func (a *Alert) GetRatio() string {
	if a.RR > 4 {
		return "90%"
	} else if a.RR < 2 {
		return "50%"
	}
	return "75%"
}


func (a *Alert) GetSens() string {
	
	if a.Start != 0 && a.Target != 0 {
		if a.Target > a.Start {
			return "LONG"
		} else if a.Target < a.Start {
			return "SHORT"
		}
	}

	return "" 
}

func AccountIsValid(db *sql.DB ,account  models.Account , alert Alert, todayProfit, summary float64) bool {

	config := account.Config
	symbolID := queries.GetSymbolIDByName(db, alert.Symbol) 

	if config.Profitability != nil && *config.Profitability > 0 {
		ratio, err := utils.ParseFloat(alert.GetRatio())
		if err != nil {
			log.Printf("Error parsing alert ratio: %v", err)
			return false
		}
		if ratio < *config.Profitability {
			log.Printf("Profitability not allowed: alert_id=%d, account_id=%d", alert.ID, account.ID)
			return false
		}
	}

	// Vérification des symbols 
	if len(account.Symbols) > 0 {
		isAllowed := false
		for _, symbol := range account.Symbols {
			if symbol == *symbolID {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			log.Printf("Symbol not allowed: alert_id=%d, account_id=%d", alert.ID, account.ID)
			return false
		}
	}

	// Vérification de la monoside
	if *config.Monoside != "" && !strings.EqualFold(*config.Monoside, alert.GetSens()) {
		log.Printf("Not allowed side: alert_id=%d, account_id=%d", alert.ID, account.ID)
		return false
	}

	// Vérification du daily profit cap
	if config.DailyProfitCap != nil && *config.DailyProfitCap > 0 {
		var maxProfit float64
		if *config.DailyProfitCapQuantifier == "%" {
			maxProfit = *config.DailyProfitCap * summary / 100
		} else {
			maxProfit = *config.DailyProfitCap
		}
		if todayProfit > maxProfit {
			log.Printf("Daily profit cap reached: alert_id=%d, day_profit=%.2f, daily_profit_cap=%.2f, account_id=%d",
				alert.ID, todayProfit, *config.DailyProfitCap, account.ID)
			return false
		}
	}

	// Vérification du loose limit
	if config.LooseLimit != nil && *config.LooseLimit > 0 {
		var maxLoose float64
		if *config.LooseLimitQuantifier == "%" {
			maxLoose = *config.LooseLimit * summary / 100 * -1
		} else {
			maxLoose = *config.LooseLimit * -1
		}
		if todayProfit < maxLoose {
			log.Printf("Loose limit reached: alert_id=%d, day_profit=%.2f, loose_limit=%.2f, account_id=%d",
				alert.ID, todayProfit, *config.LooseLimit, account.ID)
			return false
		}
	}

	return true
}