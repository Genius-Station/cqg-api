package cqgapi

import (
	
	"cqg-api/pkg/utils"
	"cqg-api/protos/WebAPI"
	"google.golang.org/protobuf/proto"
	"cqg-api/internal/websocket"
)

func TradeSubscriptions( service *wsService.WebSocketService , userID string ){
	
		
		accountSummaryParameters := &WebAPI.AccountSummaryParameters{
			RequestedFields : []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30},
		}

		tradeSubscriptionOrder := &WebAPI.TradeSubscription{
				Id: utils.GenerateMsgID(), 
				SubscriptionScopes:   []uint32{uint32(*WebAPI.TradeSubscription_SUBSCRIPTION_SCOPE_ORDERS.Enum())},
				Subscribe: proto.Bool(true),   
		}

		tradeSubscriptionPosition := &WebAPI.TradeSubscription{
			Id: utils.GenerateMsgID(), 
			SubscriptionScopes:   []uint32{uint32(*WebAPI.TradeSubscription_SUBSCRIPTION_SCOPE_POSITIONS.Enum())},
			Subscribe: proto.Bool(true),   
		}

		tradeSubscriptionSummary := &WebAPI.TradeSubscription{
			Id: utils.GenerateMsgID(), 
			SubscriptionScopes:   []uint32{uint32(*WebAPI.TradeSubscription_SUBSCRIPTION_SCOPE_ACCOUNT_SUMMARY.Enum()),},
			Subscribe: proto.Bool(true),   
			AccountSummaryParameters: accountSummaryParameters,
		}

		clientMsg := &WebAPI.ClientMsg{
			TradeSubscriptions: []*WebAPI.TradeSubscription{tradeSubscriptionOrder,tradeSubscriptionPosition,tradeSubscriptionSummary},
		}


		service.SendMessage( userID , clientMsg)

}

func SymbolSubscriptions(service *wsService.WebSocketService, userID string, symbols []string) {
	
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

	service.SendMessage(userID, clientMsg)
}
