package routes

import (
	"cqg-api/internal/handlers"
	"cqg-api/internal/websocket"
	"net/http"
)

// RegisterRoutes configure toutes les routes de l'application.
func RegisterRoutes() *http.ServeMux {

	service := wsService.NewWebSocketService()

	// Démarrer le service WebSocket dans une goroutine
	go service.Start()

	mux := http.NewServeMux()

	// Définir les routes
	mux.HandleFunc("/login", handlers.LoginHandler(service))
	mux.HandleFunc("/account", handlers.AccountHandler(service))
	mux.HandleFunc("/account/symbol", handlers.AccountSymbolHandler(service))
	mux.HandleFunc("/account/summary", handlers.AccountSummaryHandler(service))
	mux.HandleFunc("/account/order", handlers.AccountOrderHandler(service))
	mux.HandleFunc("/account/place-order", handlers.NewOrderHandler(service))
	mux.HandleFunc("/account/position", handlers.AccountPositionHandler(service))
	mux.HandleFunc("/order/cancel", handlers.AccountPositionHandler(service))
	return mux
}
