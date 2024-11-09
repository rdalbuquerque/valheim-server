package steamapi

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestGetUsername(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Errorf("error loading environment variables: %v", err)
	}
	apikey := os.Getenv("STEAM_API_KEY")
	client := NewClient(apikey)
	player, err := client.GetUserRealName("76561198073103840")
	if err != nil {
		t.Errorf("error getting player username: %v", err)
	}
	log.Printf("Player is %s", player)
}
