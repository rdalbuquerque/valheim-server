package steamapi

import (
	"encoding/json"
	"fmt"
	"godin/pkg/utils"
	"io"
	"log"
	"net/http"
)

type ClientInterface interface {
	GetUserRealName(userid string) (string, error)
}

type Client struct {
	baseUrl string
	apiKey  string
}

func NewClient(apiKey string) ClientInterface {
	baseurl := "https://api.steampowered.com"
	return Client{
		baseUrl: baseurl,
		apiKey:  apiKey,
	}
}

type SteamResponse struct {
	Response struct {
		Players []struct {
			PersonaName string `json:"personaname"`
			RealName    string `json:"realname"`
		} `json:"players"`
	} `json:"response"`
}

func (c Client) GetUserRealName(action string) (string, error) {
	steamid, err := utils.ExtractSteamId(action)
	if err != nil {
		return "", err
	}
	url := c.baseUrl + fmt.Sprintf("/ISteamUser/GetPlayerSummaries/v2?steamids=%s", steamid)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("x-webapi-key", c.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	var steamresp SteamResponse
	respBody, err := io.ReadAll(resp.Body)
	log.Printf("response body: %s", respBody)
	if err != nil {
		return "", fmt.Errorf("error reading body: %v", err)
	}
	if err := json.Unmarshal(respBody, &steamresp); err != nil {
		return "", fmt.Errorf("error unmarshalling response: %v", err)
	}
	return steamresp.Response.Players[0].RealName, nil
}
