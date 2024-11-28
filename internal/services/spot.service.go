package services

import (
	"database/sql"
	"log"
	"time"
	"cqg-api/protos/WebAPI"
)

type SpotService struct {
	DB *sql.DB
}

func NewSpotService(db *sql.DB) *SpotService {
	return &SpotService{DB: db}
}


func (s *SpotService) CheckSpotExists(symbol string) (int64, bool, error) {
	var spotFound sql.NullInt64
	err := s.DB.QueryRow(`
		SELECT asset_id 
		FROM spots 
		WHERE symbol = ? AND deleted_at IS NULL AND asset_id > 0 AND collection = 'gain_futures'`,
		symbol).Scan(&spotFound)

	if err != nil && err != sql.ErrNoRows {
		return 0, false, err
	}

	if spotFound.Valid {
		return spotFound.Int64, true, nil
	}

	return 0, false, nil
}

func (s *SpotService) createFuture() (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO futures (created_at, updated_at) VALUES (now(), now())`)
	if err != nil {
		return 0, err
	}

	assetId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	log.Println("Création du futur id:", assetId)
	return assetId, nil
}

func (s *SpotService) getCountryID(countryCode string) (sql.NullInt64, error) {
	var countryId sql.NullInt64
	err := s.DB.QueryRow(`SELECT id FROM countries WHERE iso = ?`, countryCode).Scan(&countryId)
	if err != nil && err != sql.ErrNoRows {
		return countryId, err
	}

	return countryId, nil
}

func (s *SpotService) createSpot(symbol string, countryId sql.NullInt64, data *WebAPI.ContractMetadata, assetId int64) error {
	_, err := s.DB.Exec(`
		INSERT INTO spots (instrument_id, name, deleted_at, ticker_code, iex_code, logo_path, has_real_time, 
			live_rates_code, country_id, gf_code, symbol, collection, price_precision, asset_id, asset_type, tick)
		VALUES (3, ?, NULL, NULL, NULL, NULL, 1, NULL, ?, NULL, ?, 'gain_futures', ?, ?, 'App\\\\Models\\\\Future', ?)`,
		symbol, countryId, symbol, data.DisplayPriceScale, assetId, data.TickSize)

	if err != nil {
		return err
	}

	log.Println("Spot créé pour le symbole:", symbol)
	return nil
}

func (s *SpotService) CreateSpot(data *WebAPI.ContractMetadata, symbol string) (int64, error) {
	symbolParent := symbol[:len(symbol)-3]

	var assetIdParent int64

	assetIdParent , exist, err := s.CheckSpotExists(symbolParent)
	if err != nil {
		return 0, err
	}

	countryId, err := s.getCountryID(*data.CountryCode)
	if err != nil {
		return 0, err
	}

	if !exist {
		// Création d'un futur
		assetIdParent, err = s.createFuture()
		if err != nil {
			return 0, err
		}

		// Création du spot parent
		err = s.createSpot(symbolParent, countryId, data, assetIdParent)
		if err != nil {
			return 0, err
		}
	}

	// Vérification du spot existant pour le symbole complet
	_, spotExists, err := s.CheckSpotExists(symbol)
	if err != nil {
		return 0, err
	}

	if !spotExists {
		// Vérification et conversion de la date d'expiration
		var expirationDate sql.NullTime

		if data.MaturityDate != nil && *data.MaturityDate > 0 {
			// Utiliser MaturityDate en priorité
			parsedDate := time.Unix(*data.MaturityDate, 0)
			expirationDate = sql.NullTime{Time: parsedDate, Valid: true}
		} else if data.LastTradingDate != nil && *data.LastTradingDate > 0 {
			// Sinon utiliser LastTradingDate
			parsedDate := time.Unix(*data.LastTradingDate, 0)
			expirationDate = sql.NullTime{Time: parsedDate, Valid: true}
		} else {
			// Aucune date valide
			expirationDate = sql.NullTime{Valid: false}
		}
		
	
		// Création du contrat
		res, err := s.DB.Exec(`
			INSERT INTO contracts (
				symbol, created_at, updated_at, future_id, exchange, description,
				contract_group, expiration_date, tick, tick_value, initial_margin, maintenance_margin
			) VALUES (?, NOW(), NOW(), ?, ?, ?, NULL, ?, ?, ?, ?, ?)`,
			symbol, assetIdParent, data.MicDescription, data.Description, expirationDate,
			data.TickSize, data.TickValue, data.InitialMargin, data.MaintenanceMargin,
		)
		if err != nil {
			return 0, err
		}

		contractID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}

		// Création du spot pour le contrat
		err = s.createSpot(symbol, countryId, data, contractID)
		if err != nil {
			return 0, err
		}
	}

	return 0, nil
}
