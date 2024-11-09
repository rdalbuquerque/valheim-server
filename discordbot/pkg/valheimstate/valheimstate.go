package valheimstate

import (
	"godin/pkg/aztclient"
	"godin/pkg/statestorageinterface"
	"godin/pkg/utils"
	"slices"
	"strings"
)

type State struct {
	Attributes statestorageinterface.StateAttributes
	storage    aztclient.TableClientInterface
}

func NewValheimState(aztclient aztclient.TableClientInterface) statestorageinterface.StateInterface {
	return &State{
		storage: aztclient,
	}
}

func (s *State) GetAttributes() statestorageinterface.StateAttributes {
	return s.Attributes
}

func (s *State) AddOnlinePlayer(player string) {
	if len(s.Attributes.OnlinePlayers) == 0 {
		s.Attributes.OnlinePlayers = player
	} else {
		players := strings.Split(s.Attributes.OnlinePlayers, ",")
		players = append(players, player)
		s.Attributes.OnlinePlayers = strings.Join(players, ",")
	}
}

func (s *State) RemoveOnlinePlayer(player string) {
	players := strings.Split(s.Attributes.OnlinePlayers, ",")
	players = slices.DeleteFunc(players, func(s string) bool { return s == player })
	s.Attributes.OnlinePlayers = strings.Join(players, ",")
}

func (s *State) GetOnlinePlayers() []string {
	players := strings.Split(s.Attributes.OnlinePlayers, ",")
	return players
}

func (s *State) SetStatus(status string) {
	s.Attributes.Status = status
}

func (s *State) GetStatus() string {
	return s.Attributes.Status
}

func (s *State) Save() error {
	return s.storage.Write(s.GetAttributes())
}

func (s *State) Load() error {
	state, err := s.storage.Read("ip", "online_players", "status")
	if err != nil {
		if utils.IsMissingColumnError(err) {
			return s.Save()
		}
		return err
	}
	s.Attributes.Ip = state["ip"].(string)
	s.Attributes.OnlinePlayers = state["online_players"].(string)
	s.Attributes.Status = state["status"].(string)
	return nil
}

func (s *State) SetIp(ip string) {
	s.Attributes.Ip = ip
}

func (s *State) GetIp() string {
	return s.Attributes.Ip
}
