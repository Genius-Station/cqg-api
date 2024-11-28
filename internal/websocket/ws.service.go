package wsService

import (
	"fmt"
	"log"
	"sync"
	"reflect"
	"context"
	"time"


	"cqg-api/protos/WebAPI"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// WebSocketService gère toutes les connexions WebSocket
type WebSocketService struct {
	clients    map[string]*ClientConnection // Map pour gérer les connexions par userID
	mu         sync.RWMutex               // Mutex pour protéger l'accès concurrent à la map
	register   chan *ClientConnection     // Canal pour enregistrer une connexion
	unregister chan *ClientConnection     // Canal pour déconnecter un client
	broadcast  chan []byte                // Canal pour diffuser des messages à tous les clients
	muGo sync.Mutex
}

// ClientConnection représente une connexion WebSocket d'un utilisateur
type ClientConnection struct {
	UserID string
	Conn   *websocket.Conn
	SessionToken  string
	ResponseChan chan proto.Message 
}

// Crée un nouveau WebSocketService
func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		clients:    make(map[string]*ClientConnection),
		register:   make(chan *ClientConnection),
		unregister: make(chan *ClientConnection),
		broadcast:  make(chan []byte),
	}
}

// Méthode pour démarrer le WebSocketService et écouter les connexions
func (service *WebSocketService) Start() {
	go service.handleMessages()
}

// Gère les canaux pour les messages et connexions
func (service *WebSocketService) handleMessages() {
	for {
		select {
		case clientConn := <-service.register:
			service.mu.Lock()
			service.clients[clientConn.UserID] = clientConn
			service.mu.Unlock()
			log.Printf("Client %s connecté", clientConn.UserID)

		case clientConn := <-service.unregister:
			service.mu.Lock()
			if conn, ok := service.clients[clientConn.UserID]; ok {
				delete(service.clients, clientConn.UserID)
				conn.Conn.Close()
				log.Printf("Client %s déconnecté", clientConn.UserID)
			}
			service.mu.Unlock()

		case message := <-service.broadcast:
			// Diffuser le message à tous les clients
			service.mu.RLock()
			for _, conn := range service.clients {
				if err := conn.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Println("Erreur d'envoi du message:", err)
				}
			}
			service.mu.RUnlock()
		}
	}
}

func (service *WebSocketService) SendAndReceiveMessage(userID string, request proto.Message, messageTypes []string) (proto.Message, error) {
	// Récupérer la connexion WebSocket
	
	
	conn, exists := service.GetConnection(userID)
	if !exists {
		return nil,fmt.Errorf("utilisateur %s non trouvé", userID)
	}

	err := conn.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(1*time.Second))
    if err != nil {
        log.Printf("Connexion perdue pour l'utilisateur %s: %v", userID, err)
		err :=tryReconnect(conn)
		if err != nil {
			log.Printf("Échec de la reconnexion pour le client %s : %v", conn.UserID, err)
		} else {
			log.Printf("Reconnexion réussie pour le client %s", conn.UserID)
		}
	}

	// Encoder le message Protobuf
	encodedMessage, err := proto.Marshal(request)
	if err != nil {
		log.Printf("Erreur lors de l'encodage du message Protobuf: %v", err)
		return nil,fmt.Errorf("erreur d'encodage du message Protobuf: %w", err)
	}

	// Envoyer le message
	err = conn.Conn.WriteMessage(websocket.BinaryMessage, encodedMessage)
	if err != nil {
		log.Printf("Erreur lors de l'envoi du message1: %v", err)
		return nil,fmt.Errorf("erreur lors de l'envoi du message: %w", err)
	}
	log.Printf("Message envoyé à %s", userID)

	//TODO
	responseChan := conn.ResponseChan
	timeout := time.After(30 * time.Second)
	
	for{
		select {
		case receivedResponse := <-responseChan:
			// Associer la réponse reçue au message attendu
		//	log.Printf("Réponse reçue pour %s: %+v", userID, receivedResponse)
	
			// Décodez les champs du message dans le type approprié
			msgValue := reflect.ValueOf(receivedResponse).Elem()
			for _, messageType := range messageTypes {
				fieldValue := msgValue.FieldByName(messageType)
				if fieldValue.IsValid() {
					if fieldValue.Kind() == reflect.Slice {
						if fieldValue.Len() > 0 {
							return fieldValue.Index(0).Interface().(proto.Message), nil
						}
					} else if !fieldValue.IsNil() {
						// Si le champ n'est pas nil, retourner la valeur
						return fieldValue.Interface().(proto.Message), nil
					}
				}
			}
			log.Println("Message ne correspond pas aux types attendus, attente de la prochaine réponse...")
		case <-timeout: // Timeout si aucun message reçu dans un délai raisonnable
			return nil, fmt.Errorf("délai d'attente dépassé pour la réponse")
		}
	}
	
}

func (service *WebSocketService) SendMessageToUser(userID string, message []byte) error {
	service.mu.RLock()
	defer service.mu.RUnlock()

	conn, ok := service.clients[userID]
	if !ok {
		return fmt.Errorf("client %s non trouvé", userID)
	}

	err := conn.Conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		return fmt.Errorf("erreur lors de l'envoi à %s: %v", userID, err)
	}

	return nil
}

func (service *WebSocketService) GetConnection(userID string) (*ClientConnection, bool) {
	service.mu.RLock()
	conn, exists := service.clients[userID]
	service.mu.RUnlock()

	return conn, exists
}

func (service *WebSocketService) GetResponseChan(userID string) (chan proto.Message, error) {
    service.mu.RLock() 
    defer service.mu.RUnlock()

    clientConn, exists := service.clients[userID]
    if !exists {
        return nil, fmt.Errorf("client %s non trouvé", userID)
    }

    return clientConn.ResponseChan, nil
}

func (service *WebSocketService) Register(clientConn *ClientConnection) {

	responseChan := make(chan proto.Message, 1)

	service.muGo.Lock()
	defer service.muGo.Unlock()

	service.register <- clientConn
	go func() {
		for {

			response := &WebAPI.ServerMsg{}

			_, message, err := clientConn.Conn.ReadMessage()
			if err != nil {
				log.Printf("Erreur de lecture pour le client %s: %v", clientConn.UserID, err)
				err := clientConn.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(1*time.Second))
				if err != nil {
					log.Printf("Connexion perdue pour l'utilisateur %s: %v", clientConn.UserID, err)
					err :=tryReconnect(clientConn)
					if err != nil {
						log.Printf("Échec de la reconnexion pour le client %s : %v", clientConn.UserID, err)
					} else {
						log.Printf("Reconnexion réussie pour le client %s", clientConn.UserID)
					}
				}
				break
			}

	
			err = proto.Unmarshal(message, response)
			if err != nil {
				log.Printf("Erreur lors du décodage de la réponse Protobuf: %v", err)
				continue
			}

			log.Printf("Message TEST reçu de %s:", clientConn.UserID)

			responseChan <- response

		}
	}()
	clientConn.ResponseChan = responseChan
}

func (service *WebSocketService) Unregister(clientConn *ClientConnection) {
	service.unregister <- clientConn
}

func tryReconnect( clientConn *ClientConnection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dialer := websocket.DefaultDialer
	ws, _, err := dialer.DialContext(ctx, "wss://demoapi.cqg.com:443", nil)
	if err != nil {
		return fmt.Errorf("erreur lors de la reconnexion WebSocket : %w", err)
	}

	// Message de restauration
	restoreSession := &WebAPI.RestoreOrJoinSession{
		SessionToken:         proto.String(clientConn.SessionToken),
		ClientAppId:          proto.String("WebApiTest"),
		ProtocolVersionMajor: proto.Uint32(2),
		ProtocolVersionMinor: proto.Uint32(90),
	}

	clientMsg := &WebAPI.ClientMsg{
		RestoreOrJoinSession: restoreSession,
	}
	encodedMessage, err := proto.Marshal(clientMsg)
	if err != nil {
		return fmt.Errorf("erreur d'encodage du message de restauration : %w", err)
	}

	err = ws.WriteMessage(websocket.BinaryMessage, encodedMessage)
	if err != nil {
		return fmt.Errorf("erreur d'envoi du message de restauration : %w", err)
	}

	_, response, err := ws.ReadMessage()
	if err != nil {
		return fmt.Errorf("erreur lors de la réception de la réponse de restauration : %w", err)
	}

	restoreResult := &WebAPI.ServerMsg{}
//	restoreResult := &WebAPI.RestoreOrJoinSessionResult{}
	err = proto.Unmarshal(response, restoreResult)
	if err != nil {
		return fmt.Errorf("erreur de décodage de la réponse : %w", err)
	}

	log.Printf("DEBUG %v", restoreResult)

	if restoreResult.RestoreOrJoinSessionResult != nil {
		if *restoreResult.RestoreOrJoinSessionResult.ResultCode != uint32(*WebAPI.LogonResult_RESULT_CODE_SUCCESS.Enum()) {
			return fmt.Errorf("échec de la restauration de session : %v", *restoreResult.RestoreOrJoinSessionResult.ResultCode)
		}
	}


	// Mise à jour de la connexion
	clientConn.Conn = ws
	clientConn.ResponseChan = make(chan proto.Message, 1)

	go func() {
        for {
            _, msg, err := clientConn.Conn.ReadMessage()
            if err != nil {
                log.Printf("Erreur lors de la lecture des messages pour %s: %v", clientConn.UserID, err)
                return
            }

            serverMsg := &WebAPI.ServerMsg{}
            err = proto.Unmarshal(msg, serverMsg)
            if err != nil {
                log.Printf("Erreur lors du décodage du message pour %s: %v", clientConn.UserID, err)
                continue
            }

            clientConn.ResponseChan <- serverMsg
        }
    }()

	log.Printf("Session restaurée avec succès pour le client %s", clientConn.UserID)
	return nil
}
