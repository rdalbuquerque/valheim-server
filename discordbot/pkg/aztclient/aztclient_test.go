package aztclient

import (
	"godin/pkg/statestorageinterface"
	"reflect"
	"testing"

	"github.com/joho/godotenv"
)

type TestState struct {
	Attributes statestorageinterface.StateAttributes
	storage    TableClientInterface
}

func NewTestState(stgclient TableClientInterface) statestorageinterface.StateInterface {
	return &TestState{
		storage: stgclient,
	}
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

func (ts *TestState) AddOnlinePlayer(player string) {}

func (ts *TestState) RemoveOnlinePlayer(player string) {}

func (ts *TestState) SetIp(ip string) {
	ts.Attributes.Ip = ip
}

func (ts *TestState) GetIp() string {
	return ts.Attributes.Ip
}

func (ts *TestState) GetAttributes() statestorageinterface.StateAttributes {
	return ts.Attributes
}

func (ts *TestState) Load() error {
	return nil
}

func (ts *TestState) Save() error {
	if err := ts.storage.Write(ts.Attributes); err != nil {
		return err
	}
	return nil
}

func TestGenEntity(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Errorf("error loading environment variables: %v", err)
	}
	tc, _ := NewTableClient("test", "test", "test")
	state := NewTestState(tc)
	state.SetIp("4.201.60.16")
	state.SetStatus("stopped")
	state.AddOnlinePlayer("player1")

	entity := tc.(*TableClient).genEntity(state.GetAttributes())
	expectedPropertiesLength := 3
	if len(entity.Properties) != expectedPropertiesLength {
		t.Errorf("wrong number of elements in map, expected %d but was %d", expectedPropertiesLength, len(entity.Properties))
	}
	pk := entity.PartitionKey
	rk := entity.RowKey
	if pk == "" || rk == "" {
		t.Errorf("patition key, and row key cannot be zero values")
	}
	expectedIp := "4.201.60.16"
	expectedStatus := "stopped"
	expectedOnlinePlayers := []string{"player1"}
	actualOnlinePlayers, ok := entity.Properties["online_players"].([]string)
	if !ok {
		t.Errorf("something is wrong with online_players type")
	}
	if entity.Properties["ip"] != expectedIp || entity.Properties["status"] != expectedStatus || !reflect.DeepEqual(actualOnlinePlayers, expectedOnlinePlayers) {
		t.Errorf("expected ip, status and online_players to be %s, %s and %v but they were %s, %s and %v",
			expectedIp, expectedStatus, expectedOnlinePlayers, entity.Properties["ip"], entity.Properties["status"], entity.Properties["online_players"],
		)
	}
}
