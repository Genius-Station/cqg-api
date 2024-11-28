package main

import (
	"fmt"
	"net/http"
	"cqg-api/internal/routes"
)

func main() {
	// Initialiser les routes
	mux := routes.RegisterRoutes()

	// Lancer le serveur
	fmt.Println("Serveur démarré sur http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
