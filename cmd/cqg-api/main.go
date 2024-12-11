package main

import (
	"fmt"
	"net/http"
	"cqg-api/internal/routes"
	"cqg-api/config"
	"log"
)

func main() {

	db, err := config.InitDBPg() 
	if err != nil {
		log.Fatalf("Erreur lors de la connexion à la base de données : %v", err)
	}
	defer db.Close() 


	// Initialiser les routes
	mux := routes.RegisterRoutes(db)

	

	// Lancer le serveur
	fmt.Println("Serveur démarré sur http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
