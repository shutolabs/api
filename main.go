package main

import (
	"log"
	"net/http"
	"os"

	"shuto-api/handler"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/joho/godotenv"
)

func main() {
	// Initialize the Vips library
	vips.Startup(nil)
	defer vips.Shutdown()

	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/image/", handler.ImageHandler)
	http.HandleFunc("/list/", handler.ListHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server is running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
