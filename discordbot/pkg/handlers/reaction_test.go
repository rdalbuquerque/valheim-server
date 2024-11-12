package handlers

import (
	"encoding/json"
	"godin/pkg/aztclient"
	"godin/pkg/godinerrors"
	"godin/pkg/statestorageinterface"
	"godin/pkg/utils"
	"log"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
)

type TestState struct {
	Attributes statestorageinterface.StateAttributes
	storage    aztclient.TableClientInterface
}

func NewTestState(stgclient aztclient.TableClientInterface) statestorageinterface.StateInterface {
	return &TestState{
		storage: stgclient,
	}
}

func (ts *TestState) GetAttributes() statestorageinterface.StateAttributes {
	return ts.Attributes
}

func (ts *TestState) GetStatus() string {
	return ts.Attributes.Status
}

func (ts *TestState) SetStatus(status string) {
	ts.Attributes.Status = status
}

func (ts *TestState) GetOnlinePlayers() []string {
	return []string{}
}

func (ts *TestState) AddOnlinePlayer(player string) {
	if len(ts.Attributes.OnlinePlayers) == 0 {
		ts.Attributes.OnlinePlayers = player
	} else {
		players := strings.Split(ts.Attributes.OnlinePlayers, ",")
		players = append(players, player)
		ts.Attributes.OnlinePlayers = strings.Join(players, ",")
	}
}

func (ts *TestState) RemoveOnlinePlayer(player string) {
	players := strings.Split(ts.Attributes.OnlinePlayers, ",")
	players = slices.DeleteFunc(players, func(s string) bool { return s == player })
	ts.Attributes.OnlinePlayers = strings.Join(players, ",")
}

func (ts *TestState) SetIp(ip string) {
	ts.Attributes.Ip = ip
}

func (ts *TestState) GetIp() string {
	return ts.Attributes.Ip
}

func (ts *TestState) Load() error {
	state, err := ts.storage.Read("ip", "online_players", "status")
	if err != nil {
		if utils.IsMissingColumnError(err) {
			return ts.Save()
		}
		return err
	}
	ts.Attributes.Ip = state["ip"].(string)
	ts.Attributes.OnlinePlayers = state["online_players"].(string)
	ts.Attributes.Status = state["status"].(string)
	return nil
}

type TestSteamClient struct{}

func (tsc TestSteamClient) GetUserRealName(action string) (string, error) {
	idusermap := map[string]string{
		"76561198073103840": "player1",
		"76561198073103841": "player2",
	}
	id, err := utils.ExtractSteamId(action)
	if err != nil {
		return "", err
	}
	return idusermap[id], nil
}

type TestDiscordClient struct {
	messagesSent []string
}

func (tdc *TestDiscordClient) SendMessage(msg string) error {
	tdc.messagesSent = append(tdc.messagesSent, msg)
	log.Println(msg)
	return nil
}

type TestVmssClient struct{}

func (tvc *TestVmssClient) ScaleUp() error {
	return nil
}

func (tvc *TestVmssClient) ScaleDown() error {
	return nil
}

type TestTableClient struct{}

func (ttc TestTableClient) Read(columns ...string) (map[string]interface{}, error) {
	contentBytes, err := os.ReadFile("testvalheimstate.json")
	if err != nil {
		return nil, err
	}
	stateMap := make(map[string]interface{})
	if err := json.Unmarshal(contentBytes, &stateMap); err != nil {
		return nil, err
	}
	if err := utils.ValidateColumns(stateMap, columns); err != nil {
		return nil, godinerrors.ReadError{
			Code:    godinerrors.MissingColumnError,
			Message: err.Error(),
		}
	}
	return stateMap, nil
}

func (ttc TestTableClient) Write(state statestorageinterface.StateAttributes) error {
	statefile := "testvalheimstate.json"
	statebytes, err := json.Marshal(&state)
	if err != nil {
		return err
	}
	log.Printf("saving state: %s", statebytes)
	err = os.WriteFile(statefile, statebytes, 0777)
	if err != nil {
		return err
	}
	log.Printf("saved state to: %s", statefile)
	return nil
}

func (ts *TestState) Save() error {
	if err := ts.storage.Write(ts.GetAttributes()); err != nil {
		return err
	}
	return nil
}

func TestHandleAction(t *testing.T) {
	type testcase struct {
		Action                  string
		ExpectedMessages        []string
		InitialStateJson        string
		ExpectedState           *TestState
		ExpectedStateProperties []string
	}
	testcases := []testcase{
		{
			Action:                  "start",
			ExpectedMessages:        []string{"Starting Valheim server", "Valheim server started"},
			ExpectedStateProperties: []string{"ip", "online_players", "status"},
			InitialStateJson:        `{}`,
			ExpectedState: &TestState{
				Attributes: statestorageinterface.StateAttributes{
					Ip:            "",
					OnlinePlayers: "",
					Status:        "started",
				},
			},
		},
		{
			Action:                  "stop",
			ExpectedMessages:        []string{"Stopping Valheim server", "Valheim server stopped, hope you had a great time! :grin:"},
			ExpectedStateProperties: []string{"ip", "online_players", "status"},
			InitialStateJson:        `{"ip":"192.168.0.1", "online_players": "", "status":"listening"}`,
			ExpectedState: &TestState{
				Attributes: statestorageinterface.StateAttributes{
					Ip:            "192.168.0.1",
					OnlinePlayers: "",
					Status:        "stopped",
				},
			},
		},
		{
			Action:                  "Server is now listening",
			ExpectedMessages:        []string{"Valheim server is ready, enjoy!"},
			ExpectedStateProperties: []string{"ip", "online_players", "status"},
			InitialStateJson:        `{"ip":"192.168.0.1", "online_players": "", "status":"started"}`,
			ExpectedState: &TestState{
				Attributes: statestorageinterface.StateAttributes{
					Ip:            "192.168.0.1",
					OnlinePlayers: "",
					Status:        "listening",
				},
			},
		},
		{
			Action:                  "Got connection SteamID 76561198073103840",
			ExpectedMessages:        []string{"Greetings `player1`!"},
			ExpectedStateProperties: []string{"ip", "online_players", "status"},
			InitialStateJson:        `{"ip":"192.168.0.1", "online_players": "", "status":"listening"}`,
			ExpectedState: &TestState{
				Attributes: statestorageinterface.StateAttributes{
					Ip:            "192.168.0.1",
					OnlinePlayers: "player1",
					Status:        "listening",
				},
			},
		},
		{
			Action:                  "Got connection SteamID 76561198073103841",
			ExpectedMessages:        []string{"Greetings `player2`!"},
			ExpectedStateProperties: []string{"ip", "online_players", "status"},
			InitialStateJson:        `{"ip":"192.168.0.1", "online_players": "player1", "status":"listening"}`,
			ExpectedState: &TestState{
				Attributes: statestorageinterface.StateAttributes{
					Ip:            "192.168.0.1",
					OnlinePlayers: "player1,player2",
					Status:        "listening",
				},
			},
		},
		{
			Action:                  "Closing socket 76561198073103840",
			ExpectedMessages:        []string{"Farewell `player1`..."},
			ExpectedStateProperties: []string{"ip", "online_players", "status"},
			InitialStateJson:        `{"ip":"192.168.0.1", "online_players": "player1", "status":"listening"}`,
			ExpectedState: &TestState{
				Attributes: statestorageinterface.StateAttributes{
					Ip:            "192.168.0.1",
					OnlinePlayers: "",
					Status:        "listening",
				},
			},
		},
		{
			Action:                  "Closing socket 76561198073103841",
			ExpectedMessages:        []string{"Farewell `player2`..."},
			ExpectedStateProperties: []string{"ip", "online_players", "status"},
			InitialStateJson:        `{"ip":"192.168.0.1", "online_players": "player1,player2", "status":"listening"}`,
			ExpectedState: &TestState{
				Attributes: statestorageinterface.StateAttributes{
					Ip:            "192.168.0.1",
					OnlinePlayers: "player1",
					Status:        "listening",
				},
			},
		},
	}
	storage := TestTableClient{}
	vmssclient := TestVmssClient{}
	steamclient := TestSteamClient{}
	for _, tc := range testcases {
		disclient := TestDiscordClient{}
		setState(tc.InitialStateJson)
		testState := NewTestState(storage)
		testState.Load()
		validationState := NewTestState(storage)
		ah := newActionHandler(&disclient, &vmssclient, steamclient, testState)
		err := ah.handleAction(tc.Action)
		if err != nil {
			t.Errorf("%s - error handling action: %v", tc.Action, err)
		}
		validationState.Load()
		validationAttributes := validationState.GetAttributes()
		if err != nil {
			t.Errorf("%s - error reading state: %v", tc.Action, err)
		}
		if !reflect.DeepEqual(tc.ExpectedState.Attributes, validationAttributes) {
			t.Errorf("%s - expected state attributes to be %v but was %v", tc.Action, tc.ExpectedState, validationAttributes)
		}
		sentMessages := disclient.messagesSent
		if !reflect.DeepEqual(sentMessages, tc.ExpectedMessages) {
			t.Errorf("%s - expected sent messages to be %v but were %v", tc.Action, tc.ExpectedMessages, sentMessages)
		}
	}
}

func setState(statejson string) {
	os.WriteFile("testvalheimstate.json", []byte(statejson), 0777)
}
