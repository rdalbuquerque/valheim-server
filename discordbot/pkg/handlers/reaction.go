package handlers

import (
	"encoding/json"
	"fmt"
	"godin/pkg/aztclient"
	"godin/pkg/disclient"
	"godin/pkg/statestorageinterface"
	"godin/pkg/steamapi"
	"godin/pkg/valheimstate"
	"godin/pkg/vmssclient"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

type invokeResponse struct {
	Logs []string
}

type triggerData struct {
	Data struct {
		Action string `json:"action"`
	} `json:"Data"`
}

func setInternalServerErrorWithLogs(w http.ResponseWriter, handlerErr error) {
	invokeResponse := invokeResponse{Logs: []string{handlerErr.Error()}}
	js, err := json.Marshal(invokeResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, handlerErr.Error(), http.StatusInternalServerError)
	w.Write(js)
}

func ReactionHandler(w http.ResponseWriter, r *http.Request) {
	var triggerData triggerData
	if err := json.NewDecoder(r.Body).Decode(&triggerData); err != nil {
		setInternalServerErrorWithLogs(w, fmt.Errorf("error decoding request body: %v", err))
		return
	}
	defer r.Body.Close()

	storageclient, err := aztclient.NewTableClient(os.Getenv("STATE_STORAGE_NAME"), "valheim-vmss", os.Getenv("WORLD_NAME"))
	if err != nil {
		setInternalServerErrorWithLogs(w, fmt.Errorf("error creating storageclient: %v", err))
		return
	}
	state := valheimstate.NewValheimState(storageclient)
	if err := state.Load(); err != nil {
		setInternalServerErrorWithLogs(w, fmt.Errorf("error loading state: %v", err))
		return
	}
	vmssclient, err := vmssclient.NewVmssClient(os.Getenv("VMSS_RESOURCE_GROUP_NAME"), os.Getenv("VMSS_NAME"), os.Getenv("AZURE_SUBSCRIPTION_ID"), state.GetIp())
	if err != nil {
		setInternalServerErrorWithLogs(w, fmt.Errorf("error creating vmssclient: %v", err))
		return
	}
	discordclient, err := disclient.NewDiscordClient(os.Getenv("DISCORD_BOT_TOKEN"), os.Getenv("DISCORD_CHANNEL_ID"))
	if err != nil {
		setInternalServerErrorWithLogs(w, fmt.Errorf("error creating discordclient: %v", err))
		return
	}
	steamclient := steamapi.NewClient(os.Getenv("STEAM_API_KEY"))

	ah := newActionHandler(discordclient, vmssclient, steamclient, state)

	if err := ah.handleAction(strings.Trim(triggerData.Data.Action, "\"")); err != nil {
		setInternalServerErrorWithLogs(w, fmt.Errorf("cought error: %v", err))
		return
	}

	invokeResponse := invokeResponse{Logs: []string{}}
	js, err := json.Marshal(invokeResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

type actionHandler struct {
	discordClient disclient.DiscordClientInterface
	vmssClient    vmssclient.VmssClientInterface
	steamClient   steamapi.ClientInterface
	state         statestorageinterface.StateInterface
}

func newActionHandler(
	discordclient disclient.DiscordClientInterface,
	vmssclient vmssclient.VmssClientInterface,
	steamclient steamapi.ClientInterface,
	state statestorageinterface.StateInterface,
) *actionHandler {
	return &actionHandler{
		discordClient: discordclient,
		vmssClient:    vmssclient,
		steamClient:   steamclient,
		state:         state,
	}
}

func (ah *actionHandler) handleAction(action string) error {
	if action == "start" {
		ah.state.SetStatus("starting")
		if err := ah.state.Save(); err != nil {
			return err
		}
		if err := ah.discordClient.SendMessage("Starting Valheim server"); err != nil {
			return err
		}
		if err := ah.vmssClient.ScaleUp(); err != nil {
			return err
		}
		ah.state.SetStatus("started")
		if err := ah.state.Save(); err != nil {
			return err
		}
		if err := ah.discordClient.SendMessage("Valheim server started"); err != nil {
			return err
		}
	} else if action == "stop" {
		ah.state.SetStatus("stopping")
		if err := ah.state.Save(); err != nil {
			return err
		}
		if err := ah.discordClient.SendMessage("Stopping Valheim server"); err != nil {
			return err
		}
		if err := ah.vmssClient.ScaleDown(); err != nil {
			return err
		}
		ah.state.SetStatus("stopped")
		if err := ah.state.Save(); err != nil {
			return err
		}
		if err := ah.discordClient.SendMessage("Valheim server stopped, hope you had a great time! :grin:"); err != nil {
			return err
		}
	} else if net.ParseIP(action) != nil {
		log.Printf("IP Address: %s", action)
		if err := ah.discordClient.SendMessage(fmt.Sprintf("Public IP address: `%s`", action)); err != nil {
			return err
		}
		ah.state.SetIp(action)
		if err := ah.state.Save(); err != nil {
			return err
		}
	} else if strings.Contains(action, "listening") {
		ah.state.SetStatus("listening")
		if err := ah.state.Save(); err != nil {
			return err
		}
		if err := ah.discordClient.SendMessage("Valheim server is ready, enjoy!"); err != nil {
			return err
		}
	} else if strings.Contains(action, "Got connection SteamID") {
		realname, err := ah.steamClient.GetUserRealName(action)
		if err != nil {
			return err
		}
		ah.state.AddOnlinePlayer(realname)
		if err := ah.state.Save(); err != nil {
			return err
		}
		if err := ah.discordClient.SendMessage(fmt.Sprintf("Greetings `%s`!", realname)); err != nil {
			return err
		}
	} else if strings.Contains(action, "Closing socket") {
		realname, err := ah.steamClient.GetUserRealName(action)
		if err != nil {
			return err
		}
		ah.state.RemoveOnlinePlayer(realname)
		if err := ah.state.Save(); err != nil {
			return err
		}
		if err := ah.discordClient.SendMessage(fmt.Sprintf("Farewell `%s`...", realname)); err != nil {
			return err
		}
	} else {
		if err := ah.discordClient.SendMessage(fmt.Sprintf("Received unknown action: %s", action)); err != nil {
			return err
		}
	}
	return nil
}
