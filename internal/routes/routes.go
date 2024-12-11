package routes

import (
	"cqg-api/internal/handlers"
	"cqg-api/internal/websocket"
	"net/http"
	"database/sql"
)

// RegisterRoutes configure toutes les routes de l'application.
func RegisterRoutes(db *sql.DB) *http.ServeMux {

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
	
	mux.HandleFunc("/account", handlers.CreateAccountHandler(db))
	mux.HandleFunc("/account/", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetAccountHandler(w, r, db)
	})
	mux.HandleFunc("/symbol", handlers.GetSymbolsHandler(db))

	mux.HandleFunc("/autotrade/alert/", func(w http.ResponseWriter, r *http.Request) {
		handlers.AutotradeOrderHandler(w, r, db)
	})
	return mux
}
