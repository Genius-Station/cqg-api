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
	mux.HandleFunc("/logout", handlers.LogoutHandler(service))
	mux.HandleFunc("/account/balance", handlers.AccountBalanceHandler(service))
	mux.HandleFunc("/account/symbol", handlers.AccountSymbolHandler(service))
	mux.HandleFunc("/order/new", handlers.NewOrderHandler(service))
	mux.HandleFunc("/order/cancel", handlers.CancelOrderHandler(service))
	
	return mux
}
