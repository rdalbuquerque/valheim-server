package statestorageinterface

type StateAttributes struct {
	Ip            string `json:"ip"`
	OnlinePlayers string `json:"online_players"` // for now this will just be comma delimited list of player names
	Status        string `json:"status"`
}

type StateInterface interface {
	GetAttributes() StateAttributes
	Save() error
	Load() error
	GetIp() string
	SetIp(string)
	GetOnlinePlayers() []string
	AddOnlinePlayer(string)
	RemoveOnlinePlayer(string)
	GetStatus() string
	SetStatus(string)
}
