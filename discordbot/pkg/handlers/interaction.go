package handlers

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"godin/pkg/azqclient"
	"io"
	"log"
	"net/http"
	"os"
)

// Interaction types from Discord API
const (
	InteractionPing    = 1
	InteractionCommand = 2
)

// Interaction response types
const (
	ResponsePong       = 1
	ResponseChannelMsg = 4
)

type InteractionOutput struct {
	Outputs map[string]interface{} `json:"outputs,omitempty"`
}

// Interaction structure to parse JSON payload from Discord
type Interaction struct {
	Type int `json:"type"`
	Data struct {
		Name string `json:"name"`
	} `json:"data"`
}

// PublicKey to verify the request signature (stored as an environment variable)
var discordPublicKey = os.Getenv("DISCORD_PUBLIC_KEY")

// Verify request signature function
func verifyRequest(r *http.Request) bool {
	signature := r.Header.Get("X-Signature-Ed25519")
	timestamp := r.Header.Get("X-Signature-Timestamp")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}

	// Re-attach the body for reading later
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	message := append([]byte(timestamp), body...)
	decodedSig, err := hexDecodeString(signature)
	if err != nil {
		return false
	}

	decodedPubKey, err := hexDecodeString(discordPublicKey)
	if err != nil {
		return false
	}

	return ed25519.Verify(decodedPubKey, message, decodedSig)
}

// Hex decoding utility
func hexDecodeString(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

func InteractionHandler(w http.ResponseWriter, r *http.Request) {
	// Verify request
	if os.Getenv("COMPUTERNAME") != "RODSPC" && !verifyRequest(r) {
		http.Error(w, "Invalid request signature", http.StatusUnauthorized)
		return
	}

	// Parse the interaction
	var interaction Interaction
	err := json.NewDecoder(r.Body).Decode(&interaction)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Handle the interaction based on type
	response := map[string]interface{}{}
	switch interaction.Type {
	case InteractionPing:
		log.Println("Received ping validation")
		response = map[string]interface{}{
			"type": ResponsePong,
		}
	case InteractionCommand:
		log.Printf("Received command: %s", interaction.Data.Name)
		switch interaction.Data.Name {
		case "ping":
			response = responseChannelMsg("Pong!")
		case "start", "stop":
			azqclient, err := azqclient.NewQueueClient("events")
			if err != nil {
				log.Printf("Error creating queue client: %v", err)
				http.Error(w, fmt.Sprintf("Error creating queue client: %v", err), http.StatusInternalServerError)
				return
			}
			if err = azqclient.EnqueueMessage(interaction.Data.Name); err != nil {
				log.Printf("Error enqueuing message: %v", err)
				http.Error(w, fmt.Sprintf("Error enqueuing message: %v", err), http.StatusInternalServerError)
				response = responseChannelMsg("Failed to queue the action")
				break
			}
			response = responseChannelMsg(fmt.Sprintf("Will %s the Valheim server", interaction.Data.Name))
		default:
			response = responseChannelMsg(fmt.Sprintf("Unknown command: %s", interaction.Data.Name))
		}
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

type InteractionResponse struct {
	Outputs     map[string]interface{} `json:"outputs"`
	Logs        []string               `json:"logs"`
	ReturnValue string                 `json:"returnValue"`
}

func responseChannelMsg(msg string) map[string]interface{} {
	return map[string]interface{}{
		"type": ResponseChannelMsg,
		"data": map[string]string{
			"content": msg,
		},
	}
}
