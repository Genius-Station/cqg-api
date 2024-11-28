package main

import (
	"context"
	"cqg-api/pkg/utils"
	"cqg-api/protos/WebAPI"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	"cqg-api/internal/services"
	"cqg-api/config"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// Configuration de la connexion
var serverURL = os.Getenv("URL_WS_CQG")
var userName = os.Getenv("USER_NAME_WS_CQG")
var password = os.Getenv("PASSWORD_WS_CQG")
var clientAppId = os.Getenv("APP_ID_WS_CQG")

const clientVersion = "go-client-test-2-1"
const protocolVersionMajor = int32(2)
const protocolVersionMinor = int32(90)


type SymbolList struct {
	Symbols []string `json:"symbols"`
}

var (
	ws              *websocket.Conn
	contractIDToSymbolMap = make(map[uint32]string) 
	spotService *services.SpotService
)

func main() {

	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation de la base de données: %v", err)
	}
	defer db.Close()

	spotService = services.NewSpotService(db)

	for {
		if err := connect(); err != nil {
			log.Printf("Erreur de connexion : %v. Nouvelle tentative dans 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if err := handleMessages(); err != nil {
			log.Printf("Erreur : %v. Tentative de reconnexion...", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func connect() error {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Unrecoverable error intercepted: %v. Trying to reconnect...", r)
			time.Sleep(5 * time.Second)
		}
	}()

	log.Println("Tentative de connexion au WebSocket...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	ws, _, err = websocket.DefaultDialer.DialContext(ctx, serverURL, nil)
	if err != nil {
		return fmt.Errorf("échec de connexion : %w", err)
	}

	log.Println("Connexion réussie.")
	login()
	return nil
}

func handleMessages() error {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Unrecoverable error intercepted: %v. Trying to reconnect...", r)
			time.Sleep(5 * time.Second)
		}
	}()

	defer ws.Close()

	for {
		_, response, err := ws.ReadMessage()
		if err != nil {
			return fmt.Errorf("connexion perdue ou erreur de lecture : %w", err)
		}

		receivedMsg := &WebAPI.ServerMsg{}
		err = proto.Unmarshal(response, receivedMsg)
		if err != nil {
			log.Printf("Erreur lors du décodage de la réponse : %v", err)
			continue
		}

		processMessage(receivedMsg)
	}
}

func processMessage(receivedMsg *WebAPI.ServerMsg) {
	if receivedMsg.LogonResult != nil {
		if *receivedMsg.LogonResult.ResultCode != 0 {
			log.Printf("Échec du login : %v\n", *receivedMsg.LogonResult.ResultCode)
		} else {
			log.Printf("Login réussi : %+v\n", receivedMsg)
			symbols, err := getSymbolList("./cmd/cqg-exporter/symbols.json")
			if err != nil {
				log.Fatalf("Erreur lors de la lecture des symboles: %v", err)
			}
			symbolsSubscription(symbols)
		}
	}

	if receivedMsg.InformationReports != nil {
		log.Printf("TEST : %+v", receivedMsg)
		for _, infoReport := range receivedMsg.InformationReports {
			if infoReport.SymbolResolutionReport != nil && infoReport.SymbolResolutionReport.ContractMetadata != nil {
				contractId := *infoReport.SymbolResolutionReport.ContractMetadata.ContractId
				log.Printf("Symbol %s résolu avec succès, contract_id: %v", *infoReport.SymbolResolutionReport.ContractMetadata.ContractSymbol, contractId)

				contractID := *infoReport.SymbolResolutionReport.ContractMetadata.ContractId
				contractSymbol := *infoReport.SymbolResolutionReport.ContractMetadata.Title

				if _, ok := contractIDToSymbolMap[contractID]; !ok {
					contractIDToSymbolMap[contractID] = contractSymbol
					log.Printf("Souscription réussie pour contract_id: %d et contract en cours: %s", contractID, contractSymbol)
				} 

				go handleSymbolVerification(contractSymbol , infoReport.SymbolResolutionReport.ContractMetadata)
				marketDataSubscription(contractId)
			} else {
				log.Printf("Échec de la résolution du symbole : %+v", infoReport)
			}
		}
	}

	if receivedMsg.MarketDataSubscriptionStatuses != nil {
		for _, status := range receivedMsg.MarketDataSubscriptionStatuses {
			if status.StatusCode != nil && *status.StatusCode == uint32(*WebAPI.MarketDataSubscriptionStatus_STATUS_CODE_SUCCESS.Enum()) {
				log.Printf("Souscription réussie pour contract_id %v, niveau: TRADES", *status.ContractId)
			} else {
				log.Printf("Échec de souscription pour contract_id %v: %v", *status.ContractId, *status.StatusCode)
			}
		}
	}

	if receivedMsg.RealTimeMarketData != nil {
		for _, marketData := range receivedMsg.RealTimeMarketData {
			for _, quote := range marketData.Quotes {
				if quote.QuoteUtcTime == nil || quote.Volume == nil || quote.Volume.Significand == nil || quote.ScaledPrice == nil {
					continue
				}
				contractID := *marketData.ContractId
				quantity := *quote.Volume.Significand
				price := *quote.ScaledPrice
				time := *quote.QuoteUtcTime

				symbol, ok := contractIDToSymbolMap[contractID]
				if !ok {
					continue 
				}

				log.Printf("Time: %v, Symbol: %v, Quantity: %v, Price: %.2f",
					time, symbol, quantity, float64(price)/100.0)
				
					// TODO : ENVOIE NATS	
			}
		}
	}
}

func login() {
	logon := &WebAPI.Logon{
		UserName:             utils.StringPtr(userName),
		Password:             utils.StringPtr(password),
		ClientAppId:          utils.StringPtr(clientAppId),
		ClientVersion:        utils.StringPtr(clientVersion),
		ProtocolVersionMajor: utils.Uint32Ptr(protocolVersionMajor),
		ProtocolVersionMinor: utils.Uint32Ptr(protocolVersionMinor),
	}

	clientMsg := &WebAPI.ClientMsg{
		Logon: logon,
	}

	encodedMessage, err := proto.Marshal(clientMsg)
	if err != nil {
		log.Fatalf("Erreur lors de l'encodage du message : %v", err)
	}

	err = sendMessage(encodedMessage)
	if err != nil {
		log.Fatalf("Erreur lors de l'envoi du message : %v", err)
	}
}

func symbolsSubscription(symbols []string) {
	var symbolResolutionRequests []*WebAPI.SymbolResolutionRequest
	for _, symbol := range symbols {
		symbolResolutionRequests = append(symbolResolutionRequests, &WebAPI.SymbolResolutionRequest{
			Symbol: proto.String(symbol),
		})
	}

	var informationRequests []*WebAPI.InformationRequest
	for i, request := range symbolResolutionRequests {
		informationRequests = append(informationRequests, &WebAPI.InformationRequest{
			Id:                      proto.Uint32(uint32(i + 1)),
			Subscribe:               proto.Bool(true),
			SymbolResolutionRequest: request,
		})
	}

	clientMsg := &WebAPI.ClientMsg{
		InformationRequests: informationRequests,
	}

	encodedMessage, err := proto.Marshal(clientMsg)
	if err != nil {
		log.Fatalf("Erreur lors de l'encodage du message : %v", err)
	}

	err = sendMessage(encodedMessage)
	if err != nil {
		log.Fatalf("Erreur lors de l'envoi du message : %v", err)
	}
}

func marketDataSubscription(contractId uint32) {
	marketDataSubscription := &WebAPI.MarketDataSubscription{
		RequestId:  utils.GenerateMsgID(),
		ContractId: proto.Uint32(contractId),
		Level:      proto.Uint32(uint32(*WebAPI.MarketDataSubscription_LEVEL_TRADES.Enum())),
	}

	clientMsg := &WebAPI.ClientMsg{
		MarketDataSubscriptions: []*WebAPI.MarketDataSubscription{marketDataSubscription},
	}

	encodedMessage, err := proto.Marshal(clientMsg)
	if err != nil {
		log.Fatalf("Erreur lors de l'encodage du message : %v", err)
	}

	err = sendMessage(encodedMessage)
	if err != nil {
		log.Fatalf("Erreur lors de l'envoi du message : %v", err)
	}
}

func sendMessage(message []byte) error {
	if ws == nil {
		return fmt.Errorf("webSocket non initialisé")
	}
	return ws.WriteMessage(websocket.BinaryMessage, message)
}

func getSymbolList(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la lecture du fichier: %v", err)
	}

	var symbolList SymbolList
	err = json.Unmarshal(data, &symbolList)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la désérialisation des symboles: %v", err)
	}

	return symbolList.Symbols, nil
}

func handleSymbolVerification(contractSymbol string, contractMetadata *WebAPI.ContractMetadata) {
	
	// Vérifiez si le symbole existe
	_, exist, err := spotService.CheckSpotExists(contractSymbol)
	if err != nil {
		log.Printf("Erreur lors de la vérification du symbole %s: %v", contractSymbol, err)
		return
	}

	if !exist {
		log.Printf("Spot introuvable pour le symbole %s", contractSymbol)
		_ ,err = spotService.CreateSpot( contractMetadata ,contractSymbol) 
		if err != nil {
			log.Printf("Erreur lors de la création du spot pour %s: %v", contractSymbol, err)
		} else {
			log.Printf("Spot créé pour le symbole %s", contractSymbol)
		}
	} else {
		log.Printf("Spot existant pour le symbole %s", contractSymbol)
	}
}
