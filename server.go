package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pricetra/api/database"
	"github.com/pricetra/api/setup"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Could not find .env file. Proceeding with using normal env variables.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := database.NewDbConnection()
	if err != nil {
		log.Println("❌ Could not connect to database.")
		panic(err)
	}
	defer db.Close()
	router := chi.NewRouter()
	server := setup.NewServer(db, router)

	log.Printf("🚀 Server running http://localhost:%s/playground\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), server.Router); err != nil {
		log.Fatal("❌ Could not start server", err)
	}
}
