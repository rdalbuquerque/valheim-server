package utils

import (
	"log"
	"testing"
)

func TestExtractSteamId(t *testing.T) {
	action := "Got connection SteamID 76561198073103840"
	id, err := ExtractSteamId(action)
	if err != nil {
		t.Errorf("error extracting steam id: %v", err)
	}
	log.Println(id)
}
