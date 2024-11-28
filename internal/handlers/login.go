package handlers

import (
	"context"
	"cqg-api/protos/WebAPI"
	"cqg-api/internal/websocket"
	//"fmt"
	"log"
	"net/http"
	"time"
	"encoding/json"
	"cqg-api/pkg/utils"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func  LoginHandler(service *wsService.WebSocketService) http.HandlerFunc { 
	return func (w http.ResponseWriter, r *http.Request) {
		// Configuration de la connexion
		serverURL := "wss://demoapi.cqg.com:443" // Remplacez par votre URL WebSocket
		userName := "GeniusWAPI"
		password := "GeniusWAPI"
		clientAppId := "WebApiTest"
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
	
		
		// Encoder le message avec google.golang.org/protobuf/proto
		encodedMessage, err := proto.Marshal(clientMsg)
		if err != nil {
			log.Fatalf("Erreur lors de l'encodage du message: %v", err)
		}
	
		// Connexion WebSocket
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	
	//	dialer := websocket.DefaultDialer
		dialer := *websocket.DefaultDialer 
		dialer.EnableCompression = false  
		dialer.ReadBufferSize = 8192
		ws, _, err := dialer.DialContext(ctx, serverURL, nil)
		if err != nil {
			log.Fatalf("Erreur lors de la connexion au serveur WebSocket: %v", err)
		}
		defer ws.Close()
	
		
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
		clientConn := &wsService.ClientConnection{
			UserID: userID,
			Conn:   ws,
			SessionToken: *receivedMsg.LogonResult.SessionToken,
		}
	
		service.Register(clientConn)

	
	// 	ws.Close()

	// 	// Simulation de la restauration de session
	// 	time.Sleep(5 * time.Second) // Pause avant la restauration

	// 	// Connexion WebSocket pour restauration
	// 	ws, _, err = dialer.DialContext(ctx, serverURL, nil)
	// 	if err != nil {
	// 		log.Fatalf("Erreur lors de la reconnexion au serveur WebSocket: %v", err)
	// 	}
	// //	defer ws.Close()

	// 	// Création du message RestoreOrJoinSession
	// 	restoreSession := &WebAPI.RestoreOrJoinSession{
	// 		SessionToken:         receivedMsg.LogonResult.SessionToken,
	// 		ClientAppId:          utils.StringPtr(clientAppId),
	// 		ProtocolVersionMajor: utils.Uint32Ptr(protocolVersionMajor),
	// 		ProtocolVersionMinor: utils.Uint32Ptr(protocolVersionMinor),
	// 	}

	// 	restoreMsg := &WebAPI.ClientMsg{
	// 		RestoreOrJoinSession: restoreSession,
	// 	}

	// 	encodedRestoreMessage, err := proto.Marshal(restoreMsg)
	// 	if err != nil {
	// 		log.Fatalf("Erreur lors de l'encodage du message de restauration: %v", err)
	// 	}

	// 	// Envoi du message RestoreOrJoinSession
	// 	err = ws.WriteMessage(websocket.BinaryMessage, encodedRestoreMessage)
	// 	if err != nil {
	// 		log.Fatalf("Erreur lors de l'envoi du message de restauration: %v", err)
	// 	}

	// 	// Lecture de la réponse RestoreOrJoinSessionResult
	// 	_, restoreResponse, err := ws.ReadMessage()
	// 	if err != nil {
	// 		log.Fatalf("Erreur lors de la réception de la réponse de restauration: %v", err)
	// 	}

	// 	restoreResult := &WebAPI.ServerMsg{}
	// 	err = proto.Unmarshal(restoreResponse, restoreResult)
	// 	if err != nil {
	// 		log.Fatalf("Erreur lors du décodage de la réponse de restauration: %v", err)
	// 	}

	// 	log.Printf("Réponse de restauration décodée : %+v\n", restoreResult)

		json.NewEncoder(w).Encode(receivedMsg)

	//	ws.Close()
	}
} 

