package main

import (
	"godin/pkg/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/interactions", handlers.InteractionHandler)
	mux.HandleFunc("/reactions", handlers.ReactionHandler)
	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}
	log.Printf("Listening on %s...", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, mux))
}
