package Manager

import (
	"encoding/json"
	"fmt"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/client"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/model"
	"strconv"
	"v2ray.com/core/common/errors"
	"v2ray.com/core/transport/internet"
)

type Manager struct {
	HandlerServiceClient  *client.HandlerServiceClient
	StatsServiceClient    *client.StatsServiceClient
	UserRuleServiceClient *client.UserRuleServerClient
	CurrentNodeInfo       *model.NodeInfo
	NextNodeInfo          *model.NodeInfo
	UserChanged           bool
	UserToBeMoved         map[string]model.UserModel
	UserToBeAdd           map[string]model.UserModel
	Users                 map[string]model.UserModel
	Id2PrefixedIdmap      map[uint]string
	Id2DisServer          map[uint]string
	MainAddress           string
	MainListenPort        uint16
	NodeID                uint
	CheckRate             int
	SpeedTestCheckRate    int
}

func (manager *Manager) GetUsers() map[string]model.UserModel {
	return manager.Users
}

func (manager *Manager) Add(user model.UserModel) {
	manager.UserChanged = true
	manager.UserToBeAdd[user.PrefixedId] = user
}

func (manager *Manager) Remove(prefixedId string) bool {
	user, ok := manager.Users[prefixedId]
	if ok {
		manager.UserChanged = true
		manager.UserToBeMoved[user.PrefixedId] = user
		return true
	}
	return false
}

func (manager *Manager) UpdataUsers() {
	var successfully_removed, successfully_add []string
	if manager.CurrentNodeInfo.Server_raw != "" {
		if manager.CurrentNodeInfo.Sort == 0 {
			// SS server
			/// remove inbounds
			for key, value := range manager.UserToBeMoved {
				if err := manager.HandlerServiceClient.RemoveInbound(value.PrefixedId); err == nil {
					newErrorf("Successfully remove user %s ", key).AtInfo().WriteToLog()
					successfully_removed = append(successfully_removed, key)
				} else {
					newError(err).AtDebug().WriteToLog()
					successfully_removed = append(successfully_removed, key)
				}
			}
		} else if manager.CurrentNodeInfo.Sort == 11 || manager.CurrentNodeInfo.Sort == 12 {
			// VMESS
			// Remove users
			for key, value := range manager.UserToBeMoved {
				if err := manager.HandlerServiceClient.DelUser(value.Email); err == nil {
					newErrorf("Successfully remove user %s ", key).AtInfo().WriteToLog()
					successfully_removed = append(successfully_removed, key)
				} else {
					newError(err).AtDebug().WriteToLog()
					successfully_removed = append(successfully_removed, key)
				}
			}
		}

	}
	if manager.NextNodeInfo.Server_raw != "" {
		if manager.NextNodeInfo.Sort == 0 {
			// SS server
			/// add inbounds
			for key, value := range manager.UserToBeAdd {
				if err := manager.HandlerServiceClient.AddSSInbound(value); err == nil {
					newErrorf("Successfully add user %s ", key).AtInfo().WriteToLog()
					successfully_add = append(successfully_add, key)
				} else {
					newError(err).AtDebug().WriteToLog()
					successfully_add = append(successfully_add, key)
				}
			}
		} else if manager.NextNodeInfo.Sort == 11 || manager.NextNodeInfo.Sort == 12 {
			// VMESS
			// add users
			for key, value := range manager.UserToBeAdd {
				if err := manager.HandlerServiceClient.AddUser(value); err == nil {
					newErrorf("Successfully add user %s ", key).AtInfo().WriteToLog()
					successfully_add = append(successfully_add, key)
				} else {
					newError(err).AtDebug().WriteToLog()
					successfully_add = append(successfully_add, key)
				}
			}
		}
	}
	for index := range successfully_removed {
		delete(manager.Users, successfully_removed[index])
		delete(manager.UserToBeMoved, successfully_removed[index])
	}
	for index := range successfully_add {
		manager.Users[successfully_add[index]] = manager.UserToBeAdd[successfully_add[index]]
		delete(manager.UserToBeAdd, successfully_add[index])
	}
	manager.Id2PrefixedIdmap = map[uint]string{}
	for key, value := range manager.Users {
		manager.Id2PrefixedIdmap[value.UserID] = key
	}
}

func (manager *Manager) UpdateMainAddressAndProt(node_info *model.NodeInfo) {
	if node_info.Sort == 11 || node_info.Sort == 12 {
		if node_info.Server["port"] == "0" || node_info.Server["port"] == "" {
			manager.MainAddress = "127.0.0.1"
			manager.MainListenPort = 10550
			if node_info.Server["inside_port"] != "" {
				port, err := strconv.ParseUint(node_info.Server["inside_port"].(string), 10, 0)
				if err == nil {
					manager.MainListenPort = uint16(port)
				}
			}
		} else {
			manager.MainAddress = "0.0.0.0"
			manager.MainListenPort = 10550
			if node_info.Server["port"] != "" {
				port, err := strconv.ParseUint(node_info.Server["port"].(string), 10, 0)
				if err == nil {
					manager.MainListenPort = uint16(port)
				}
			}
		}
	}
}
func (m *Manager) AddMainInbound() error {
	if m.NextNodeInfo.Server_raw != "" {
		if m.NextNodeInfo.Sort == 11 || m.NextNodeInfo.Sort == 12 {
			m.UpdateMainAddressAndProt(m.NextNodeInfo)
			var streamsetting *internet.StreamConfig
			streamsetting = &internet.StreamConfig{}

			if m.NextNodeInfo.Server["protocol"] == "ws" {
				host := "www.bing.com"
				path := "/"
				if m.NextNodeInfo.Server["path"] != "" {
					path = m.NextNodeInfo.Server["path"].(string)
				}
				if m.NextNodeInfo.Server["host"] != "" {
					host = m.NextNodeInfo.Server["host"].(string)
				}
				streamsetting = client.GetWebSocketStreamConfig(path, host)
			} else if m.NextNodeInfo.Server["protocol"] == "kcp" || m.NextNodeInfo.Server["protocol"] == "mkcp" {
				header_key := "noop"
				if m.NextNodeInfo.Server["protocol_param"] != "" {
					header_key = m.NextNodeInfo.Server["protocol_param"].(string)
				}
				streamsetting = client.GetKcpStreamConfig(header_key)
			}
			if err := m.HandlerServiceClient.AddVmessInbound(m.MainListenPort, m.MainAddress, streamsetting); err != nil {
				return err
			} else {
				newErrorf("Successfully add MAIN INBOUND %s port %d", m.MainAddress, m.MainListenPort).AtInfo().WriteToLog()
			}
		}

	}
	return nil
}
func (m *Manager) AddOuntBound(disnodeinfo *model.DisNodeInfo) error {
	if disnodeinfo.Server_raw != "" {
		if disnodeinfo.Sort == 11 || disnodeinfo.Sort == 12 {
			var streamsetting *internet.StreamConfig
			streamsetting = &internet.StreamConfig{}

			if disnodeinfo.Server["protocol"] == "ws" {
				host := "www.bing.com"
				path := "/"
				if m.NextNodeInfo.Server["path"] != "" {
					path = disnodeinfo.Server["path"].(string)
				}
				if m.NextNodeInfo.Server["host"] != "" {
					host = disnodeinfo.Server["host"].(string)
				}
				streamsetting = client.GetWebSocketStreamConfig(path, host)
			} else if m.NextNodeInfo.Server["protocol"] == "kcp" || m.NextNodeInfo.Server["protocol"] == "mkcp" {
				header_key := "noop"
				if m.NextNodeInfo.Server["protocol_param"] != "" {
					header_key = disnodeinfo.Server["protocol_param"].(string)
				}
				streamsetting = client.GetKcpStreamConfig(header_key)
			}
			if err := m.HandlerServiceClient.AddVmessOutbound(disnodeinfo.Server_raw+fmt.Sprintf("%d", disnodeinfo.UserId), disnodeinfo.Port, disnodeinfo.Server["server_address"].(string), streamsetting, m.HandlerServiceClient.ConvertVmessUser(
				m.Users[m.Id2PrefixedIdmap[disnodeinfo.UserId]])); err != nil {
				return err
			} else {
				newErrorf("Successfully add Outbound %s port %d", disnodeinfo.Server_raw+fmt.Sprintf("%d", disnodeinfo.UserId), disnodeinfo.Port).AtInfo().WriteToLog()
			}
		}
		if disnodeinfo.Sort == 0 {
			if err := m.HandlerServiceClient.AddSSOutbound(m.Users[m.Id2PrefixedIdmap[disnodeinfo.UserId]], disnodeinfo); err != nil {
				return newError("User Chipter %S", m.Users[m.Id2PrefixedIdmap[disnodeinfo.UserId]].Method).Base(err)
			} else {
				newErrorf("Successfully add user %s  outbound  %s ", m.Id2PrefixedIdmap[disnodeinfo.UserId], disnodeinfo.Server_raw).AtInfo().WriteToLog()

			}
		}
		m.AddUserRule(disnodeinfo.Server_raw+fmt.Sprintf("%d", disnodeinfo.UserId), m.Users[m.Id2PrefixedIdmap[disnodeinfo.UserId]].Email)
	}

	return nil
}
func (m *Manager) AddUserRule(tag, email string) {
	if err := m.UserRuleServiceClient.AddUserRelyRule(tag, []string{email}); err == nil {
		newErrorf("Successfully add user %s  %s server rule  ", email, tag).AtInfo().WriteToLog()
	} else {
		newError(err).AtDebug().WriteToLog()
	}
}
func (m *Manager) RemoveUserRule(email string) {

	if err := m.UserRuleServiceClient.RemveUserRelayRule([]string{email}); err == nil {
		newErrorf("Successfully remove user %s  rule", email).AtInfo().WriteToLog()
	} else {
		newError(err).AtDebug().WriteToLog()
	}
}

func (m *Manager) RemoveInbound() {
	if m.CurrentNodeInfo.Server_raw != "" {
		if m.CurrentNodeInfo.Sort == 11 || m.CurrentNodeInfo.Sort == 12 {
			m.UpdateMainAddressAndProt(m.CurrentNodeInfo)
			if err := m.HandlerServiceClient.RemoveInbound(m.HandlerServiceClient.InboundTag); err != nil {
				newError(err).AtWarning().WriteToLog()
			} else {
				newErrorf("Successfully remove main inbound %s", m.HandlerServiceClient.InboundTag).AtInfo().WriteToLog()
			}
		}
	}
}
func (m *Manager) RemoveOutBound(tag string, userid uint) {
	if err := m.HandlerServiceClient.RemoveOutbound(tag); err != nil {
		newError(err).AtWarning().WriteToLog()
	} else {
		newErrorf("Successfully remove outbound %s", tag).AtInfo().WriteToLog()
	}
	m.RemoveUserRule(m.Users[m.Id2PrefixedIdmap[userid]].Email)
}
func (m *Manager) RemoveAllUserOutBound() {
	for id, server := range m.Id2DisServer {
		m.RemoveOutBound(server+fmt.Sprintf("%d", id), id)
	}
}
func (m *Manager) CopyUsers() {
	jsonString, err := json.Marshal(m.Users)
	if err != nil {
		newError(err).AtWarning().WriteToLog()
	}
	var usertobemoved map[string]model.UserModel
	err = json.Unmarshal(jsonString, &usertobemoved)
	if err != nil {
		newError(err).AtWarning().WriteToLog()
	}
	m.UserToBeMoved = usertobemoved
	m.UserToBeAdd = map[string]model.UserModel{}
}
func (m *Manager) UpdateServer() error {
	m.CopyUsers()
	m.UpdataUsers()
	m.RemoveInbound()
	err := m.AddMainInbound()
	m.Users = map[string]model.UserModel{}
	return err
}

func newErrorf(format string, a ...interface{}) *errors.Error {
	return newError(fmt.Sprintf(format, a...))
}

func newError(values ...interface{}) *errors.Error {
	values = append([]interface{}{"SSPanelPlugin: "}, values...)
	return errors.New(values...)
}
