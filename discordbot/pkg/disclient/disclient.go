package disclient

import (
	"fmt"
	"godin/pkg/utils"

	"github.com/bwmarrin/discordgo"
)

const logprefix = "disclient"

type DiscordClientInterface interface {
	SendMessage(msg string) error
}

type DiscordClient struct {
	channelId string
	client    *discordgo.Session
}

func NewDiscordClient(bottoken, channelid string) (DiscordClientInterface, error) {
	discord, err := discordgo.New("Bot " + bottoken)
	if err != nil {
		utils.Log(logprefix, fmt.Sprintf("error creating discord client: %v", err))
		return nil, fmt.Errorf("error creating discord client: %v", err)
	}
	return &DiscordClient{
		client:    discord,
		channelId: channelid,
	}, nil
}

func (dc *DiscordClient) SendMessage(msg string) error {
	if _, err := dc.client.ChannelMessageSend(dc.channelId, msg); err != nil {
		return err
	}
	return nil
}
