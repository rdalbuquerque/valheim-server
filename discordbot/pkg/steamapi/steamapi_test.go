package steamapi

import (
	"fmt"
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
	fmt.Printf("apikey is %s", apikey)
	client := NewClient(apikey)
	connectedplayer, err := client.GetUserRealName("Got connection SteamID 76561198073103840")
	if err != nil {
		t.Errorf("error getting player username: %v", err)
	}
	fmt.Printf("Player is %s", connectedplayer)

	disconnectedplayer, err := client.GetUserRealName("Got connection SteamID 76561198073103840")
	if err != nil {
		t.Errorf("error getting player username: %v", err)
	}
	fmt.Printf("Player is %s", disconnectedplayer)
}
