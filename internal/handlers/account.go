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

func AccountHandler(service *wsService.WebSocketService) http.HandlerFunc {
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

func AccountSummaryHandler(service *wsService.WebSocketService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := "userName1"
		

		accountSummaryParameters := &WebAPI.AccountSummaryParameters{
		 	RequestedFields: []uint32{6, 7, 8, 9, 10, 11, 12, 13}, 
		}

		tradeSubscription := &WebAPI.TradeSubscription{
		 	Id:                   proto.Uint32(12345), 
		 	SubscriptionScopes:   []uint32{4,1},
		 	Subscribe:            proto.Bool(true),   
		 	AccountSummaryParameters: accountSummaryParameters,
		}

		 clientMsg := &WebAPI.ClientMsg{
		 	TradeSubscriptions: []*WebAPI.TradeSubscription{tradeSubscription},
		}

		// Utiliser le service pour envoyer et recevoir
		messageTypes := []string{"TradeSubscriptionStatuses"}
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
			Symbol: proto.String("F.US.ZUI"),
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

		// Retourner la réponse au client
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Erreur lors de l'encodage de la réponse JSON : %v", err)
			http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		}

	}
}

func AccountOrderHandler(service *wsService.WebSocketService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := "userName1"
		
	   tradeSubscription := &WebAPI.TradeSubscription{
			Id:                   utils.GenerateMsgID(), 
			SubscriptionScopes:   []uint32{1},
			Subscribe:            proto.Bool(true),   
	   }

		clientMsg := &WebAPI.ClientMsg{
			TradeSubscriptions: []*WebAPI.TradeSubscription{tradeSubscription},
	   }

		// Utiliser le service pour envoyer et recevoir
		messageTypes := []string{"test"}
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

func AccountPositionHandler(service *wsService.WebSocketService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := "userName1"
		
	   tradeSubscription := &WebAPI.TradeSubscription{
			Id:                   utils.GenerateMsgID(), 
			SubscriptionScopes:   []uint32{2,1},
			Subscribe:            proto.Bool(true),   
	   }

		clientMsg := &WebAPI.ClientMsg{
			TradeSubscriptions: []*WebAPI.TradeSubscription{tradeSubscription},
	   }

		// Utiliser le service pour envoyer et recevoir
		messageTypes := []string{"PositionStatuses"}
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

func NewOrderHandler(service *wsService.WebSocketService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Identifier l'utilisateur
		userID := "userName1"

		// Paramètres pour la requête NewOrder
		requestID := utils.GenerateMsgID()
		accountID :=  proto.Int32(17795820)
		contractID := proto.Uint32(1)
		clOrderID := proto.String("20")
		orderType := proto.Uint32(uint32(*WebAPI.Order_ORDER_TYPE_MKT.Enum()))  // WebAPI.Order_ORDER_TYPE_LMT
		duration := proto.Uint32(2) // WebAPI.Order_DURATION_GTC
		side := proto.Uint32(1) // WebAPI.Order_SIDE_BUY
		qtySignificant := int64(1000)
		qtyExponent := int32(-2)
		isManual := proto.Bool(false)

		// Créer le champ qty en utilisant cqg.Decimal
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

		// Construire l'ordre
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

		// Créer la requête NewOrder
		newOrder := &WebAPI.NewOrder{
			Order: order,
		}

		orderRequest := &WebAPI.OrderRequest{
			RequestId: requestID,
			NewOrder:  newOrder,
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