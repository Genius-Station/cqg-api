package handlers

import (
	wsService "cqg-api/internal/websocket"
	"cqg-api/pkg/utils"
	"cqg-api/protos/WebAPI"
	"cqg-api/protos/WebAPI/common"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"google.golang.org/protobuf/types/known/timestamppb"

	"google.golang.org/protobuf/proto"
)

func AccountBalanceHandler(service *wsService.WebSocketService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := "userName1"
		
		balancesRequest := &WebAPI.LastStatementBalancesRequest{}
		informationRequest := &WebAPI.InformationRequest{
			Id:                           utils.GenerateMsgID(),
			LastStatementBalancesRequest: balancesRequest,
		}

		clientMsg := &WebAPI.ClientMsg{
			InformationRequests: []*WebAPI.InformationRequest{informationRequest},
		}

		 // Utiliser le service pour envoyer et recevoir
		 messageTypes := []string{"InformationReports"}
		 response, err := service.SendAndReceiveMessage(userID, clientMsg, messageTypes)
		 if err != nil {
			 log.Printf("Erreur : %v", err)
			 http.Error(w, err.Error(), http.StatusInternalServerError)
			 return
		 }
 

		 // Retourner la réponse au client
		 w.Header().Set("Content-Type", "application/json")
		 if err := json.NewEncoder(w).Encode(response); err != nil {
			 log.Printf("Erreur lors de l'encodage de la réponse JSON : %v", err)
			 http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		 }

	}
}


func AccountSymbolHandler(service *wsService.WebSocketService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := "userName1"
		
		symbolResolutionRequests :=  &WebAPI.SymbolResolutionRequest{
			Symbol: proto.String("F.US.ZUT"),
		}

		informationRequests := &WebAPI.InformationRequest{
			Id:  proto.Uint32(1),
			Subscribe: proto.Bool(true), 
			SymbolResolutionRequest: symbolResolutionRequests,
		}

		clientMsg := &WebAPI.ClientMsg{
			InformationRequests: []*WebAPI.InformationRequest{informationRequests},
		}

		// Utiliser le service pour envoyer et recevoir
		messageTypes := []string{"InformationReports"}
		response , err := service.SendAndReceiveMessage(userID, clientMsg, messageTypes)
		if err != nil {
			log.Printf("Erreur : %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		infoReport, ok := response.(*WebAPI.InformationReport)
		if ok {
		
			if infoReport.StatusCode != nil && *infoReport.StatusCode == 1 {
				if infoReport.SymbolResolutionReport != nil && 
				infoReport.SymbolResolutionReport.ContractMetadata != nil {
					
					metadata := infoReport.SymbolResolutionReport.ContractMetadata

					if metadata.ContractId != nil && metadata.Title != nil {
						service.AddSubscription(
							userID,
							*metadata.Title,                                
							*metadata.ContractId,       
						)
						log.Printf("Souscription ajoutée : Symbol=%s, ContractID=%d", *metadata.Title, *metadata.ContractId)
					} 
				}
			}
		}
	


		// Retourner la réponse au client
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Erreur lors de l'encodage de la réponse JSON : %v", err)
			http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		}

	}
}

func NewOrderHandler(service *wsService.WebSocketService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Identifier l'utilisateur
		userID := "userName1"

		if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

		
		var newOrderParams struct {
            ClOrderId string `json:"cl_order_id"`
        }

        err := json.NewDecoder(r.Body).Decode(&newOrderParams)
        if err != nil {
            http.Error(w, "Invalide data", http.StatusBadRequest)
            return
        }

		// Paramètres pour la requête NewOrder
		requestID := utils.GenerateMsgID()
		accountID :=  proto.Int32(17795820)
		contractID := proto.Uint32(1)
		clOrderID := &newOrderParams.ClOrderId
		orderType := proto.Uint32(uint32(*WebAPI.Order_ORDER_TYPE_MKT.Enum()))  // WebAPI.Order_ORDER_TYPE_LMT
		duration := proto.Uint32(2) // WebAPI.Order_DURATION_GTC
		side := proto.Uint32(uint32(*WebAPI.Order_SIDE_SELL.Enum()))  // Order_SIDE_SELL Order_SIDE_BUY
		qtySignificant := int64(1000)
		qtyExponent := int32(-2)
		isManual := proto.Bool(false)

		decimal := &common.Decimal{
			Significand: proto.Int64(qtySignificant),
			Exponent:   proto.Int32(qtyExponent),
		}

		// Ajouter un horodatage
		timestamp := timestamppb.New(time.Now())

		// Ajouter les attributs supplémentaires
		// extraAttributes := []*common.NamedValue{
		// 	{
		// 		Name:  proto.String("ALGO_CQG_cost_model"),
		// 		Value:  proto.String("1"),
		// 	},
		// 	{
		// 		Name:  proto.String("ALGO_CQG_percent_of_volume"),
		// 		Value: proto.String("0"),
		// 	},
		// }

		
		order := &WebAPI.Order{
			AccountId:       accountID,
			WhenUtcTimestamp: timestamp,
			ContractId:      contractID,
			ClOrderId:       clOrderID,
			OrderType:       orderType,
			ScaledLimitPrice: proto.Int64(10),
			Duration:        duration,
			Side:            side,
			Qty:             decimal,
			IsManual:        isManual,
		//	AlgoStrategy:    proto.String("CQG ARRIVALPRICE"),
		//	ExtraAttributes: extraAttributes,
		}

		newOrder := &WebAPI.NewOrder{
			Order: order,
		}

		orderRequest := &WebAPI.OrderRequest{
			RequestId: requestID,
			NewOrder:  newOrder,
		}

		clientMsg := &WebAPI.ClientMsg{
			OrderRequests: []*WebAPI.OrderRequest{orderRequest},
		}

		messageTypes := []string{"OrderStatuses", "OrderRequestRejects"}
		response, err := service.SendAndReceiveMessage(userID, clientMsg, messageTypes)
		if err != nil {
			log.Printf("Erreur : %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Erreur lors de l'encodage de la réponse JSON : %v", err)
			http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		}
	}
}

func CancelOrderHandler(service *wsService.WebSocketService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Identifier l'utilisateur
		userID := "userName1"


		if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

		
		var cancelOrderParams struct {
            OrderId string `json:"order_id"`
            ClOrderId string `json:"cl_order_id"`
			OrigClOrderId string `json:"orig_cl_order_id"`
			AccountId int32 `json:"account_id"`
        }

        err := json.NewDecoder(r.Body).Decode(&cancelOrderParams)
        if err != nil {
            http.Error(w, "Invalide data", http.StatusBadRequest)
            return
        }

		requestID := utils.GenerateMsgID()
	//	accountID :=  proto.Int32(17795820)
		
	
		now := time.Now()

		// Conversion en protobuf.Timestamp
		timestamp := timestamppb.New(now)

		// Créer la requête NewOrder
		cancelOrder := &WebAPI.CancelOrder{
			ClOrderId: &cancelOrderParams.ClOrderId,
			OrderId: &cancelOrderParams.OrderId,
			AccountId: &cancelOrderParams.AccountId,
			OrigClOrderId: &cancelOrderParams.OrigClOrderId,
			WhenUtcTimestamp: timestamp,
		}

		orderRequest := &WebAPI.OrderRequest{
			RequestId: requestID,
			CancelOrder:  cancelOrder,
		}

		// Créer le message ClientMsg
		clientMsg := &WebAPI.ClientMsg{
			OrderRequests: []*WebAPI.OrderRequest{orderRequest},
		}

		// Envoyer le message via WebSocket
		messageTypes := []string{"OrderStatuses", "OrderRequestRejects"}
		response, err := service.SendAndReceiveMessage(userID, clientMsg, messageTypes)
		if err != nil {
			log.Printf("Erreur : %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Retourner la réponse au client
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Erreur lors de l'encodage de la réponse JSON : %v", err)
			http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		}
	}
}