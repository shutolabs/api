package main

import (
	"fmt"
	"net/http"
	"os"

	"shuto-api/config"
	"shuto-api/handler"
	"shuto-api/utils"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/joho/godotenv"
)

func main() {
	// Initialize logger first with error handling
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	if err := utils.InitLogger(logLevel); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Now we can safely use the logger
	utils.Info("Initializing VIPS library")
	customLoggingHandler := func(messageDomain string, messageLevel vips.LogLevel, message string) {
		// switch over messageLevel and use utils to log
		switch messageLevel {
		case vips.LogLevelError:
			utils.Error("VIPS error", "vips_message", message)
		case vips.LogLevelWarning:
			utils.Warn("VIPS warning", "vips_message", message)
		case vips.LogLevelInfo:
			utils.Info("VIPS info", "vips_message", message)
		}
	}

	// Set logging settings with custom handler and desired verbosity level
	vips.LoggingSettings(customLoggingHandler, vips.LogLevelWarning)

	vips.Startup(nil)
	defer vips.Shutdown()

	// Load .env file
	if err := godotenv.Load(); err != nil {
		utils.Warn("No .env file found, using environment variables")
	} else {
		utils.Debug("Loaded environment variables from .env file")
	}

	// Initialize services and utilities
	utils.Info("Initializing services")
	imageUtils := utils.NewImageUtils()
	executor := utils.NewCommandExecutor()
	configManager := config.NewDomainConfigManager(&config.FileConfigLoader{}, "config/domains.yaml")
	rclone := utils.NewRclone(executor, configManager)

	// Create handler functions
	imageHandler := func(w http.ResponseWriter, r *http.Request) {
		handler.ImageHandler(w, r, imageUtils, rclone)
	}

	listHandler := func(w http.ResponseWriter, r *http.Request) {
		handler.ListHandler(w, r, imageUtils, rclone)
	}

	downloadHandler := func(w http.ResponseWriter, r *http.Request) {
		handler.DownloadHandler(w, r, imageUtils, rclone)
	}

	// Register routes
	utils.Info("Registering HTTP routes")
	http.HandleFunc("/"+config.ApiVersion+"/image/", imageHandler)
	http.HandleFunc("/"+config.ApiVersion+"/list/", listHandler)
	http.HandleFunc("/"+config.ApiVersion+"/download/", downloadHandler)

	// host swaggerui
	fs := http.FileServer(http.Dir("./public/swaggerui"))
	http.Handle("/swaggerui/", http.StripPrefix("/swaggerui/", fs))

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		utils.Debug("No PORT environment variable found, using default", "port", "8080")
	}

	// Start server
	utils.Info("Starting server", "port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		utils.Fatal("Server failed to start", "error", err)
	}
}
