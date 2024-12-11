package handlers

import (
	"context"
	"cqg-api/cmd/cqg-api/models/queries"
	"cqg-api/internal/websocket"
	"cqg-api/pkg/cqgapi"
	"cqg-api/pkg/utils"
	"cqg-api/protos/WebAPI"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
	"database/sql"
	"path"
	"strconv"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func  LoginHandler(service *wsService.WebSocketService) http.HandlerFunc { 
	return func (w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

		var loginParams struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }

        err := json.NewDecoder(r.Body).Decode(&loginParams)
        if err != nil {
            http.Error(w, "Invalide data", http.StatusBadRequest)
            return
        }

		if loginParams.Username == "" {
			http.Error(w, "Param username required", http.StatusBadRequest)
			return 
		}
		if loginParams.Password == "" {
			http.Error(w, "Param password required", http.StatusBadRequest)
			return
		}
		
		serverURL := os.Getenv("URL_WS_CQG")
		userName := loginParams.Username
		password := loginParams.Password
		clientAppId := os.Getenv("APP_ID_WS_CQG")
		clientVersion := "python-client-test-2-1"
		protocolVersionMajor := int32(2)
		protocolVersionMinor := int32(90)
		sessionSettings := proto.Uint32(uint32(*WebAPI.Logon_SESSION_SETTING_ALLOW_SESSION_RESTORE.Enum()))
	
		// Création du message Logon
		logon := &WebAPI.Logon{
			UserName:             utils.StringPtr(userName),
			Password:             utils.StringPtr(password),
			ClientAppId:          utils.StringPtr(clientAppId),
			ClientVersion:        utils.StringPtr(clientVersion),
			ProtocolVersionMajor: utils.Uint32Ptr(protocolVersionMajor),
			ProtocolVersionMinor: utils.Uint32Ptr(protocolVersionMinor),
			SessionSettings:      []uint32{*sessionSettings},
		}
	
		clientMsg := &WebAPI.ClientMsg{
			Logon: logon,
		}
	
		encodedMessage, err := proto.Marshal(clientMsg)
		if err != nil {
			log.Fatalf("Erreur lors de l'encodage du message: %v", err)
		}
	
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	
		dialer := websocket.DefaultDialer
		ws, _, err := dialer.DialContext(ctx, serverURL, nil)
		if err != nil {
			log.Fatalf("Erreur lors de la connexion au serveur WebSocket: %v", err)
		}
		//defer ws.Close()
	
		
		// Envoi du message
		err = ws.WriteMessage(websocket.BinaryMessage, encodedMessage)
		if err != nil {
			log.Fatalf("Erreur lors de l'envoi du message: %v", err)
		}
	
		log.Println("Message envoyé, en attente de la réponse...")
	
		// Lecture de la réponse
		_, response, err := ws.ReadMessage()
		if err != nil {
			log.Fatalf("Erreur lors de la réception de la réponse: %v", err)
		}
	
		log.Printf("Réponse reçue (encodée) : %v\n", response)
	
		// Décodez la réponse si elle utilise un message Protobuf
		receivedMsg := &WebAPI.ServerMsg{}
		err = proto.Unmarshal(response, receivedMsg)
		if err != nil {
			log.Fatalf("Erreur lors du décodage de la réponse : %v", err)
		}
	
		log.Printf("Réponse décodée : %+v\n", receivedMsg)
	
	
		userID := "userName1" 
		if *receivedMsg.LogonResult.ResultCode == uint32(*WebAPI.LogonResult_RESULT_CODE_SUCCESS.Enum()) {

			
			clientConn := &wsService.ClientConnection{
				UserID: userID,
				Conn:   ws,
				SessionToken: *receivedMsg.LogonResult.SessionToken,
			}
	
			service.Register(clientConn)
		}

		
		json.NewEncoder(w).Encode(receivedMsg)

		// souscription orders position et summary 
	//	cqgapi.TradeSubscriptions(service,userID)
		cqgapi.SymbolSubscriptions(service,userID, []string{"ZUC","ZUT","ZUI"})

	}
} 

func  LogoutHandler(service *wsService.WebSocketService) http.HandlerFunc { 
	return func (w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

		userID := "userName1"

		logon := &WebAPI.Logoff{}
	
		clientMsg := &WebAPI.ClientMsg{
			Logoff: logon,
		}
	
		messageTypes := []string{"LoggedOff"}
		response, err := service.SendAndReceiveMessage(userID, clientMsg, messageTypes)
		if err != nil {
			log.Printf("Erreur : %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		service.Disconnect(userID)

		// Retourner la réponse au client
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Erreur lors de l'encodage de la réponse JSON : %v", err)
			http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		}

	}
} 

func CreateAccountHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var accountPayload queries.AccountPayload
		err := json.NewDecoder(r.Body).Decode(&accountPayload)
		if err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		accountID, err := queries.CreateAccount(db, &accountPayload)
		if err != nil {
			log.Printf("Erreur lors de la création du compte : %v", err)
			http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"message": "Compte créé avec succès",
			"account_id": accountID,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Erreur lors de l'encodage de la réponse JSON : %v", err)
			http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		}
	}
}


func GetAccountHandler (w http.ResponseWriter, r *http.Request, db *sql.DB){
	userIDStr := path.Base(r.URL.Path) 
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	account, err := queries.GetAccountByUserID(db, userID)
	if err != nil {
		if err.Error() == "account not found" {
			http.Error(w, "Account not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch account", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}




