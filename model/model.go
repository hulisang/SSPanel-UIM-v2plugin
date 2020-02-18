package model

type UserModel struct {
	UserID     uint   `json:"id"`
	Uuid       string `json:"uuid"`
	Email      string `json:"email"`
	Passwd     string `json:"passwd"`
	Method     string `json:"method"`
	Port       uint16 `json:"port"`
	AlterId    uint32
	PrefixedId string
}

type UserTrafficLog struct {
	UserID   uint   `json:"user_id"`
	Uplink   uint64 `json:"u"`
	Downlink uint64 `json:"d"`
}

type NodeInfo struct {
	NodeID     uint
	Server_raw string `json:"server"`
	Sort       uint   `json:"sort"`
	Server     map[string]interface{}
}

type UserOnLineIP struct {
	UserId uint   `json:"user_id"`
	Ip     string `json:"ip"`
}

type DisNodeInfo struct {
	Server_raw string `json:"dist_node_server"`
	Sort       uint   `json:"dist_node_sort"`
	Port       uint16 `json:"port"`
	Server     map[string]interface{}
	UserId     uint `json:"user_id"`
}
